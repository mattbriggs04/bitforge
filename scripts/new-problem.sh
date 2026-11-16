#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 || $# -gt 3 ]]; then
  echo "Usage: $0 <slug> <title> [category]"
  echo "Example: $0 parse-uart-frame \"Parse UART Frame\" \"Embedded C\""
  exit 1
fi

slug="$1"
title="$2"
category="${3:-Embedded C}"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
base_dir="${repo_root}/backend/seed/problems/${slug}"

json_escape() {
  local value="$1"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  value="${value//$'\n'/\\n}"
  printf "%s" "$value"
}

if [[ ! "$slug" =~ ^[a-z0-9-]+$ ]]; then
  echo "Error: slug must match ^[a-z0-9-]+$"
  exit 1
fi

if [[ -e "$base_dir" ]]; then
  echo "Error: ${base_dir} already exists"
  exit 1
fi

title_json="$(json_escape "$title")"
category_json="$(json_escape "$category")"

mkdir -p "$base_dir/templates/c" "$base_dir/cases/visible" "$base_dir/cases/hidden" "$base_dir/assets"

cat > "$base_dir/statement.md" <<STATEMENT
Implement the required C API for **${title}**.

Describe the low-level behavior clearly and keep hidden validation details in backend test cases.
STATEMENT

cat > "$base_dir/constraints.md" <<'CONSTRAINTS'
- List input assumptions and memory safety constraints.
- Include any timing/space complexity target.
- Clarify edge-case behavior.
CONSTRAINTS

cat > "$base_dir/templates/c/starter.c" <<'STARTER'
#include <stddef.h>

int solve(void) {
    return 0;
}
STARTER

cat > "$base_dir/cases/visible/example_visible.c" <<'VISIBLE'
int got = solve();
int expected = 1;
case_passed = (got == expected);
VISIBLE

cat > "$base_dir/cases/hidden/example_hidden.c" <<'HIDDEN'
int got = solve();
int expected = 2;
case_passed = (got == expected);
HIDDEN

cat > "$base_dir/problem.json" <<JSON
{
  "slug": "${slug}",
  "title": "${title_json}",
  "difficulty": "easy",
  "category": "${category_json}",
  "problemType": "systems",
  "shortDescription": "One-line summary of what the candidate must implement.",
  "statementFile": "statement.md",
  "constraintsFile": "constraints.md",
  "metadata": {
    "estimatedMinutes": 20,
    "focusAreas": ["systems"],
    "source": "seed"
  },
  "tags": ["c"],
  "templates": [
    {
      "language": "c",
      "starterCodeFile": "templates/c/starter.c",
      "notes": "Replace starter API and function signatures for this problem."
    }
  ],
  "assets": [],
  "judge": {
    "runner": "c_assert_harness_v1",
    "config": {
      "c_std": "c11"
    }
  },
  "visibleCases": [
    {
      "name": "example visible",
      "displayInput": "int got = solve();",
      "displayExpect": "got = 1",
      "explanation": "Replace this scaffold with a real call and its concrete expected result.",
      "codeFile": "cases/visible/example_visible.c",
      "weight": 1,
      "sortOrder": 10
    }
  ],
  "hiddenCases": [
    {
      "name": "example hidden",
      "displayInput": "hidden input",
      "displayExpect": "hidden output",
      "explanation": "Hidden case for submit mode",
      "codeFile": "cases/hidden/example_hidden.c",
      "weight": 1,
      "sortOrder": 110
    }
  ]
}
JSON

echo "Created ${base_dir#${repo_root}/}"
echo "Next steps:"
echo "  1) Edit starter code and markdown"
echo "  2) Replace example case snippets"
echo "  3) Run: make seed"
