# BitForge current implementation

Updated 2026-11-14. This supersedes the former Postgres/Redis/competitions runbook.

## Architecture

- Next.js/TypeScript frontend; Go API; SQLite database.
- Separate `backend/cmd/judge-worker`, durable SQLite jobs with bounded concurrency, leases, retries and fenced completion.
- Additive migration `004_durable_judge.sql` adds snapshots and jobs. Existing data is preserved. Historical attempts cannot reconstruct old grading versions.
- Docker-only compile/run for C11, C++20, Python 3/NumPy and Rust 2021. No host execution fallback.
- Legacy C harness retained inside Docker; new generated code uses standalone stdin/stdout and external comparison.

## Generation

- Edit `generation/request.json`, then `make generation-prompt`.
- Give `generation/next-question.txt` to an agent in this repository.
- Seed 0 chooses a fresh seed; nonzero replays configuration selection.
- Contract: `backend/internal/authoring/bundle.go`.
- `backend/cmd/problems` has `prompt`, `validate`, and `import` commands.
- Import reruns validation, checks reference/independent agreement, freezes generated inputs/outputs, and rejects known wrong solutions.
- Examples: ML log-sum-exp (all languages), ADC numerical error budget, pthread multiple choice.
- Private answers, explanations, verifiers and rationales are not public metadata. Explanations require a completed local user's submission.

## Run

```bash
make sandbox-image
npm run dev
```

Docker Desktop must be running. `make dev` starts API, frontend and worker. Database is `backend/bitforge.db`.

## Verify

```bash
npm test
cd backend
BITFORGE_SANDBOX_TEST=1 go test ./internal/sandbox -v
```

Generated imports pin the locally resolved Docker image ID; retain that image. Full setup, commands, contracts and limitations are in README.md.

## Scope and remaining work

This delivers the first generation/practice milestone. The catalog still needs pagination and separate facets; curriculum diversity, adaptive review scheduling, math rendering and Postgres deployment remain future work. Generated conceptual content requires editorial judgment. Existing header identity and the legacy verdict harness are not suitable for authenticated public hosting.

## Workspace

The session began with extensive uncommitted SQLite/UI/content changes and removed legacy deployment files. Preserve unrelated changes; do not reset the worktree. No commit was requested.


## Mixed interviews (2026-11-15)

- `005_interviews.sql` adds ordered many-to-many interview membership.
- Manifest contract and confined file loader: `backend/internal/authoring/interview.go`.
- CLI `validate-interview` / `import-interview` validates every member, then imports atomically.
- `/interviews` lists collections; `/interviews/[slug]` shows question order, resume link and personal solved progress. Question pages retain interview navigation via `?interview=<slug>`.
- First set: `generation/interviews/binary-classifiers/interview.json` (3 single-choice, 1 Python coding, ~45 min). Validation report is alongside it.
- Questions retain independent grading. No separate timed interview attempts or aggregate exam scores are implemented.


## Math rendering

`frontend/components/markdown.tsx` centralizes remark-math + rehype-katex rendering. Use `$...$` inline and `$$` display blocks in problem prose, options (inline only), sample notes and explanations. KaTeX CSS/fonts are imported locally in the root layout. JSON requires escaped backslashes; generator instructions and README document this. Grader inputs, output fixtures and option answer IDs remain plain. The binary-classifier interview now demonstrates LaTeX formatting.
