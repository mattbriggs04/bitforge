# BitForge

Daily interview and exam practice for embedded systems, AI/ML, computer vision, computer architecture, and operating systems.

BitForge supports existing C function exercises plus validated generated questions: standalone C11, C++20, Python/NumPy and Rust programs, single choice, multiple select, and numerical answers. Language availability belongs to each problem; old C exercises are not automatically translated.

## Run locally

Prerequisites: Go 1.25+, Node/npm, and a running Docker daemon. No host compiler or LLM API key is required.

```bash
make sandbox-image
npm --prefix frontend install
npm run dev
```

`make dev` starts the API, frontend and a bounded judge worker. Open http://localhost:3000. The API runs on port 8080. SQLite lives at `backend/bitforge.db`; startup migrates and loads the legacy seed catalog without deleting generated questions.

To run services separately, use `make api`, `make worker`, and `make frontend`. A submission remains queued while no worker is running.

## Generate a new question with an agent

1. Edit `generation/request.json`: choose category, format, language, difficulty, topic, level, and time budget. Use `random` for category/format/language/difficulty. Seed `0` selects a fresh seed on each invocation; a nonzero seed reproduces the selection. `avoid` supplies topics or questions to exclude.
2. Run `make generation-prompt`. It writes the fully expanded request and instructions to `generation/next-question.txt`.
3. Give that text to your coding agent working in this repository. The agent authors a draft JSON bundle, runs validation, repairs failures, and imports it only after validation succeeds. There is no dependency on a particular LLM provider and no paid API call hidden in this command.

The agent contract is `backend/internal/authoring/bundle.go`. Complete examples live in `generation/examples/`. Each run requests exactly one question. The seed controls selection and generated inputs; it does not make LLM wording deterministic. Keep the actual bundle to reproduce the question.

Manual validation and publication:

```bash
cd backend
go run ./cmd/problems validate -bundle ../generation/drafts/my-question.json
go run ./cmd/problems import -bundle ../generation/drafts/my-question.json
```

Import always reruns validation; a supplied report or publication flag cannot bypass it. Rejected content stays outside the catalog. Existing hand-authored slugs cannot be overwritten by generated imports. Imported generated content can be revised under its own slug; queued submissions retain their frozen grading version.

### Try the examples

From `backend/`:

```bash
go run ./cmd/problems import -bundle ../generation/examples/stable-log-sum-exp.json
go run ./cmd/problems import -bundle ../generation/examples/adc-quantization-error.json
go run ./cmd/problems import -bundle ../generation/examples/pthread-counter-interleaving.json
```

These cover a numerically stable ML reduction in all four languages, an embedded error-budget calculation, and a pthread interleaving question. Validation launches isolated processes for each test and language, so a coding import can take a minute or more. Example validation reports record actual checked cases and image IDs; they are evidence, not import credentials.

### Validation contract

Coding bundles contain:

- A precise statement, constraints, objectives, compiling starters, and executable reference solutions for every advertised language.
- A different independent verifier and 2–20 hand-derived anchor cases, including visible and hidden cases.
- A deterministic input generator, supplied the recorded seed on stdin. It prints a JSON array of 20–100 uniquely named `{ "name": "...", "input": "..." }` objects.
- At least two distinct compiling mutants representing plausible incorrect implementations. All must be rejected by the generated suite.

The validator checks anchors against the independent implementation, freezes its expected outputs, compares every reference against those outputs, compiles starters, and challenges the suite with mutants. Ordinary submissions use the frozen cases and need no LLM or reference execution. Program outputs are checked in the trusted worker; programs cannot declare their own verdicts.

Comparators are trimmed exact text, one finite number, or recursive JSON. Numeric/JSON tolerances use `abs(actual-expected) <= max(absolute, relative*abs(expected))`; zero tolerances preserve exact integers, including integers beyond float64 precision. Numerical units and tolerances must appear in the statement. Multiple-select answers are sets of stable option IDs, with duplicate selections rejected.

Choice and numerical bundles include an executable independent calculation/model plus reasoning and option rationales. The validator verifies its output against the supplied answer. Conceptual truth and the quality of independent reasoning still need editorial judgment: agreement between programs is not a proof that an LLM-authored question is correct. No arbitrary generated checker code executes in the trusted worker.

## Mathematical notation

Question prose supports LaTeX math through `remark-math` and KaTeX:

- Inline: `$\nabla L(\mathbf{w})$`.
- Display: `$$` on separate lines around the equation.
- Supported in statements, constraints, explanations, option rationales, and visible sample explanations. Multiple-choice option text supports inline math.
- Keep titles, descriptions, tags, code, test inputs/outputs, and stored answer IDs plain. Programming code fences remain literal.

Use JSON serialization when authoring bundles: every LaTeX backslash must be doubled in JSON source. For example:

```json
{
  "statement": "Compute $\\nabla L(\\mathbf{w})$.\n\n$$\nF_1=\\frac{2\\,\\mathrm{TP}}{2\\,\\mathrm{TP}+\\mathrm{FP}+\\mathrm{FN}}\n$$"
}
```

Escape literal dollar signs in Markdown (`\$`), including currency. Use KaTeX-supported math commands without HTML, external resources, color commands, document classes or preambles. Invalid math stays readable rather than crashing the page. Preview equations before importing; the computational validator checks grading correctness, not LaTeX syntax. Long display equations scroll inside their pane on smaller screens. KaTeX fonts are bundled locally, and rendered formulas include MathML for assistive technology.

## Mixed-format interviews

An interview groups reusable questions into an ordered collection. Each member keeps its own format, tests, submission history, drafts, and explanation. The overview shows personal solved progress and resumes at the first unsolved question; question pages keep previous/next links and the interview context. Progress is cumulative problem completion, not a separate timed exam attempt or aggregate exam score.

Open **Interviews → Binary classifiers** for three AI/ML multiple-choice questions and one Python coding exercise (about 45 minutes). The source is `generation/interviews/binary-classifiers/interview.json`.

An interview manifest uses `schemaVersion`, `slug`, `title`, `category`, `description`, and `questions`: an ordered array of relative question-bundle filenames. It accepts 1–20 questions, including different formats. Referenced files must remain inside the manifest directory. Each question follows the existing generated-bundle contract.

```bash
cd backend
go run ./cmd/problems validate-interview -bundle ../generation/interviews/binary-classifiers/interview.json
go run ./cmd/problems import-interview -bundle ../generation/interviews/binary-classifiers/interview.json
```

Import validates every question first, then publishes all members and their order in one database transaction. A failed validation or import leaves the previous catalog and collection unchanged. Question slugs must be unique within a collection; a question may belong to multiple interviews. Unpublishing a member hides its interview rather than silently dropping or renumbering that question.

For another mixed interview, ask the agent to author the requested question bundles plus a manifest, then run `import-interview`. `make generation-prompt` still expands a single-question request.

## Practice experience

- The homepage selects an unfinished exercise and keeps that day's selection stable on the device.
- Search/filter the catalog by category, difficulty, or tag.
- Coding supports language selection, saved drafts per question/language/version, sample runs, submission, and per-case results.
- Exam questions use radio buttons, checkboxes, or a numeric input. Explanations and option rationales are retrieved only after the user's completed submission, rather than embedded in the public question payload.
- Existing legacy solution reveal panels remain available for hand-authored C questions.

## Judge and persistence

```text
Next.js → Go API → SQLite submission + snapshot + job (one transaction)
                           ↓
                   bounded judge workers
                           ↓
                 disposable Docker containers
                           ↓
                 trusted output comparison
```

`problem_versions` stores content-addressed snapshots of the public specification, grader configuration and cases. Submissions reference a version; editing or reseeding live tests cannot change queued work. Pre-migration historical submissions have no reconstructible version and need resubmission if still queued.

Jobs use atomic claims, ten-minute leases, fenced lease tokens, at most three attempts, and delayed retries. Each job has an eight-minute worker deadline. API restarts do not lose work. A crashed worker's lease is reclaimed; a stale worker cannot finalize the new owner's job. `JUDGE_WORKERS` defaults to 2 and is capped at 8 per process.

SQLite is appropriate for local practice. This is the first implementation milestone; catalog pagination/facets, broader curriculum coverage, richer review scheduling, and a Postgres deployment path remain follow-up work.

### Execution boundary

Compilation and execution both run as an unprivileged user in a read-only container with no network, no host credentials or Docker socket, dropped capabilities, and explicit CPU/memory/process/file/output/time limits. Only submitted source and its compiled artifact are mounted read-only. Writable directories are bounded tmpfs mounts. Expected outputs and reference code are never mounted into a generated submission's container. Each case starts a fresh container.

A daemon-side timeout also terminates the container if the worker crashes. The worker removes containers after execution; Docker Desktop must remain running. There is no fallback to executing code on the host.

Generated imports resolve `SANDBOX_IMAGE` to an immutable image ID and store it with the grader. Retain that local image to execute those problem versions. Rebuilding the image and reimporting creates a new execution environment. Image IDs also capture architecture; cross-platform workers need compatible images and explicit revalidation.

The legacy `c_assert_harness_v1` runner now executes inside Docker, but still combines test snippets and submitted code and reads harness verdicts from stdout. It remains suitable for personal legacy practice, not adversarial grading. New generated coding content uses `stdio_v1` with an external comparator; exam questions use `answer_v1`.

This remains a personal-use app with local header-based identity, not authenticated public hosting. For shared hosting, add real authentication/authorization, admission/rate limits, and a reviewed isolation deployment (optionally `SANDBOX_RUNTIME=runsc` on a configured Linux daemon). The trusted worker needs Docker access; the API does not.

## Existing hand-authored C problems

Bundles remain under `backend/seed/problems/<slug>/`. Use:

```bash
./scripts/new-problem.sh <slug> "<Title>" "<Category>"
make seed
```

See `backend/seed/problems/README.md` for snippet contracts. Seeding loads these bundles into SQLite; it is independent of the generated-content validation/import path.

## Verification

```bash
npm test
cd backend
BITFORGE_SANDBOX_TEST=1 go test ./internal/sandbox -v
```

The ordinary test suite covers comparison semantics, schema validation, oracle disagreement, surviving mutants, snapshot preservation, lease fencing, and explanation privacy. Docker integration tests compile/run all four languages and check filesystem/network restrictions, output caps and timeout cleanup. They are opt-in to keep ordinary tests usable without Docker.

## Configuration and API

See `.env.example`. Main options: `DATABASE_URL`, `HTTP_PORT`, `DEFAULT_USER_HANDLE`, `JUDGE_WORKERS`, `COMPILE_TIMEOUT`, `RUN_TIMEOUT`, `SANDBOX_IMAGE`, optional `SANDBOX_RUNTIME`, and `BACKEND_API_URL`. Compilers are fixed inside the sandbox image. `APP_ENV=production` disables development reset.

- `GET /api/v1/interviews`
- `GET /api/v1/interviews/{slug}` (ordered questions with personal completion)
- `GET /api/v1/problems` (`q`, `difficulty`, `category`, `tag`)
- `GET /api/v1/problems/{slug}`
- `POST /api/v1/submissions`: `{ problemSlug, language, mode, sourceCode }`
- `GET /api/v1/submissions/{id}`
- `GET /api/v1/problems/{slug}/explanation?submissionId=...`
- `GET /health` and `GET /api/v1/health`
- `POST /api/v1/dev/reset`: destructive development-only reset, including generated content

`language` is `c`, `cpp`, `python`, `rust`, or `answer`. Coding `mode` is `run` (visible tests) or `submit` (all tests). The browser sends `X-User-Key` and `X-User-Handle` for stable local identity.
