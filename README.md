# BitForge

BitForge is a systems-oriented interview practice platform focused on firmware, embedded C, low-level networking, and security engineering workflows.

This MVP is intentionally C-first with architecture ready to add C++, Rust, and Python templates later.

## What Is Implemented

- Landing page with systems-focused product identity
- Problem catalog with search/filter by difficulty/category/tag
- Problem detail page with:
  - statement
  - constraints
  - tags/metadata
  - sample cases
  - reference assets
- In-browser C solve experience using Monaco editor
- Submission workflow:
  - `Run Samples` mode (visible tests)
  - `Submit` mode (visible + hidden tests)
- Async judge pipeline using Redis queue + Go worker
- Postgres-backed data model with migrations and seed data
- Seeded systems problems:
  - `bf-strlen`
  - `bf-memcpy`
  - `bf-memmove`
  - `ring-buffer-int`
  - `parse-ipv4-header`
  - `debounce-button-isr`

## Stack

- Frontend: Next.js (App Router) + TypeScript
- Backend: Go (net/http + repository/service architecture)
- Database: Postgres 16
- Queue/coordination: Redis 7
- Local orchestration: Docker Compose

## Monorepo Layout

```text
.
├── backend
│   ├── cmd
│   │   ├── api
│   │   ├── worker
│   │   ├── migrate
│   │   └── seed
│   └── internal
│       ├── config
│       ├── db
│       ├── httpapi
│       ├── judge
│       ├── migrations
│       ├── model
│       ├── queue
│       ├── repository
│       └── service
├── frontend
│   ├── app
│   │   ├── api/backend/[...path]  # proxy route
│   │   ├── problems
│   │   └── page.tsx
│   ├── components
│   └── lib
└── docker-compose.yml
```

## Data Model (Postgres)

Core tables:

- `users`
- `problems`
- `problem_tags`
- `problem_language_templates`
- `problem_assets`
- `problem_judge_configs`
- `problem_test_cases`
- `submissions`
- `submission_test_results`

The schema is designed for richer systems content, not only plain stdin/stdout algorithm prompts. Problem assets, per-language templates, and JSON payload test cases support future labs (headers, blobs, multi-file assets, etc.).

## Adding New Problems

Use `backend/internal/db/seed.go` as the source of truth for MVP content.

1. Add a new `seedProblem` entry in `defaultSeedProblems()`.
2. Set metadata:
   - `slug`, `title`, `difficulty`, `category`, `problem_type`
   - `statement`, `constraints`, `tags`, `metadata`
3. Add at least one language template (currently `c`):
   - `starter_code`
   - `notes`
4. Add visible cases in `VisibleCases`:
   - `display_input`, `display_expected`, `explanation`
   - `payload.code` (C snippet that sets `case_passed`)
5. Add hidden evaluator cases in `HiddenCases`:
   - same payload style, but `Hidden: true`
6. Optionally add `Assets` for diagrams/register maps/protocol notes.
7. Run:
   - `make seed`
   - restart API/worker if already running

### Case Payload Contract

Current runner `c_assert_harness_v1` expects each test case payload to include:

- `payload.code` (required): C snippet that sets `case_passed` to truthy/falsy.

The judge wraps user code + each test snippet into a generated harness and reports per-case results to `submission_test_results`.

### Runner Extensibility

- Problem behavior is selected by `problem_judge_configs.runner`.
- MVP uses one runner (`c_assert_harness_v1`), but runner dispatch can be extended for:
  - register-level hardware simulation
  - packet blob fixtures
  - multi-file builds
  - isolated sandbox executors

## Submission Pipeline

1. Frontend sends code to `POST /api/v1/submissions`
2. Backend stores submission in Postgres (`queued`)
3. Backend pushes submission ID to Redis list queue
4. Worker pops queued submissions and marks them `running`
5. Worker loads problem judge config + tests
6. Worker compiles/runs C harness and persists verdict + per-case outcomes
7. Frontend polls `GET /api/v1/submissions/{id}` for status/results

### Hidden Tests

- Hidden test definitions never leave the backend
- `Run Samples` evaluates only visible sample tests
- `Submit` evaluates visible + hidden tests

## Local Development

### Prerequisites

- Docker + Docker Compose

### Start everything

```bash
make up
```

Services:

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Postgres: localhost:5432
- Redis: localhost:6379

The backend container runs seed on startup (`cmd/seed`) before starting the API, so initial content is available automatically.

### Stop

```bash
make down
```

## Useful Commands

```bash
make backend        # run API locally (without docker)
make worker         # run worker locally
make migrate        # apply migrations
make seed           # seed sample problems
make backend-test   # go test ./...
make frontend       # next dev
make frontend-lint  # eslint
```

## API Surface

- `GET /health`
- `GET /api/v1/health`
- `GET /api/v1/problems`
  - Query params: `q`, `difficulty`, `category`, `tag`
- `GET /api/v1/problems/{slug}`
- `POST /api/v1/submissions`
  - Body: `{ problemSlug, language, mode, sourceCode }`
  - `mode`: `run` or `submit`
- `GET /api/v1/submissions/{id}`

## Environment Variables

See `.env.example` for defaults.

Key values:

- `DATABASE_URL`
- `REDIS_ADDR`
- `SUBMISSION_QUEUE`
- `DEFAULT_USER_HANDLE`
- `C_COMPILER`
- `COMPILE_TIMEOUT`
- `RUN_TIMEOUT`
- `BACKEND_API_URL` (frontend server-side/proxy target)

## MVP Judge Security Note

The current judge executes compiled C directly inside the worker container for practical MVP speed. It is structured so you can replace the execution layer with a hardened sandbox (namespaces/seccomp/firecracker/isolated runners) without changing API contracts.

## Future Ideas
- Compete with friends
    - Start a room that others can join
    - Variety of modes: code golf, time based, number of problems completed * difficulty for scoring
    - Ability to select time, difficulty, number of rounds
- Highly optimized libc implementations
    - Achieving speeds close to actual memcpy using SIMD, word-level, etc. (hard problems)
- Debugging problems
- RTOS problems (FreeRTOS, Zephyr, NuttX)
