# Seed Problem Bundles

BitForge seed content is file-based. Each problem lives in its own folder under `backend/seed/problems/<slug>/`.

## Bundle Layout

```text
backend/seed/problems/<slug>/
├── problem.json
├── statement.md
├── constraints.md
├── templates/
│   └── c/
│       └── starter.c
├── cases/
│   ├── visible/
│   │   └── *.c
│   └── hidden/
│       └── *.c
└── assets/
    └── *
```

## How Seeding Works

- `backend/internal/db/seed.go` loads every `problem.json` from this directory.
- File references in `problem.json` are read and stored in Postgres.
- Case snippets are stored as `payload.code` and compiled into the harness at judge time.
- Hidden cases stay backend-side and are not returned by problem detail APIs.

## `problem.json` Essentials

Required top-level fields:

- `slug`, `title`, `difficulty`, `category`, `problemType`
- `shortDescription`
- `statementFile`, `constraintsFile`
- `templates[]` (at least one language)
- `judge.runner` and `judge.config`
- `visibleCases[]`, `hiddenCases[]`

Each case object must include:

- `name`
- `codeFile` (C snippet path)
- `weight`
- `sortOrder`

Display metadata for visible samples:

- `displayInput`
- `displayExpect`
- `explanation`

## Case Snippet Contract

Runner `c_assert_harness_v1` expects each case snippet to set `case_passed`.

Example:

```c
int got = my_func(3);
int expected = 7;
case_passed = (got == expected);
```

The harness injects:

- standard headers: `stdbool.h`, `stddef.h`, `stdint.h`, `stdio.h`, `string.h`, `stdlib.h`
- your submitted source code
- all configured test snippets

## Authoring Workflow

1. Scaffold a new bundle.
2. Fill statement/constraints and starter code.
3. Write visible and hidden case snippets.
4. Update `problem.json` metadata/tags/order.
5. Reseed database.

Scaffold command:

```bash
./scripts/new-problem.sh <slug> "<Title>" "<Category>"
```

Reseed:

```bash
make seed
```
