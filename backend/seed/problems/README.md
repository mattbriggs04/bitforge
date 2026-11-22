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
- File references in `problem.json` are read and stored in SQLite.
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

Optional problem fields:

- `solutionFile` points to a Markdown file that is exposed as `metadata.solutionMarkdown` for the reveal panel.

Each case object must include:

- `name`
- `codeFile` (C snippet path)
- `weight`
- `sortOrder`

Display metadata for visible samples:

- `displayInput`: literal arguments and C setup/call sequence, including buffers, lengths, initial state and any prior operations. Use newlines for readable snippets.
- `displayExpect`: concrete return values and output fields/buffer contents, not a description such as "fields decoded". For addresses, use relationships such as `ret = dst` rather than an invented runtime address.
- `explanation`: why these specific inputs produce that result. Do not repeat the case name.

The UI numbers and separates each visible example, and renders input and expected result as literal, whitespace-preserving code blocks. These display strings must stay consistent with the referenced visible case snippet; hidden case payloads are never used to populate examples.

For instance, a CAN sample should include the complete eight bytes and `parse_can_control(payload, sizeof(payload), &out)`, then the return code and each decoded field. A phrase such as "8-byte payload" is not a reproducible example.

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

## Practice metadata and hints

Each `metadata` object includes `practice: {role, level, focus}`, `targetMinutes`, measurable `objectives`, and `hints: {nudge, approach}`. Role IDs and focus values come from `backend/internal/curriculum/curriculum.go`. The seed loader validates these fields. Keep hints public and spoiler-light: a concept first, then a strategy. Leave scaffold and walkthrough help to an agent. Existing `estimatedMinutes` and `focusAreas` remain for compatibility.
