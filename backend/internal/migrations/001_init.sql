CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY DEFAULT (
        lower(hex(randomblob(4))) || '-' ||
        lower(hex(randomblob(2))) || '-4' ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        substr('89ab', abs(random()) % 4 + 1, 1) ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        lower(hex(randomblob(6)))
    ),
    handle TEXT NOT NULL UNIQUE,
    email TEXT UNIQUE,
    client_key TEXT UNIQUE,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS problems (
    id TEXT PRIMARY KEY DEFAULT (
        lower(hex(randomblob(4))) || '-' ||
        lower(hex(randomblob(2))) || '-4' ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        substr('89ab', abs(random()) % 4 + 1, 1) ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        lower(hex(randomblob(6)))
    ),
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    difficulty TEXT NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
    category TEXT NOT NULL,
    problem_type TEXT NOT NULL,
    short_description TEXT NOT NULL,
    statement_md TEXT NOT NULL,
    constraints_md TEXT NOT NULL DEFAULT '',
    metadata TEXT NOT NULL DEFAULT '{}',
    is_published INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS problem_tags (
    problem_id TEXT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    PRIMARY KEY (problem_id, tag)
);

CREATE TABLE IF NOT EXISTS problem_language_templates (
    id TEXT PRIMARY KEY DEFAULT (
        lower(hex(randomblob(4))) || '-' ||
        lower(hex(randomblob(2))) || '-4' ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        substr('89ab', abs(random()) % 4 + 1, 1) ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        lower(hex(randomblob(6)))
    ),
    problem_id TEXT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    language TEXT NOT NULL,
    starter_code TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    UNIQUE (problem_id, language)
);

CREATE TABLE IF NOT EXISTS problem_assets (
    id TEXT PRIMARY KEY DEFAULT (
        lower(hex(randomblob(4))) || '-' ||
        lower(hex(randomblob(2))) || '-4' ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        substr('89ab', abs(random()) % 4 + 1, 1) ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        lower(hex(randomblob(6)))
    ),
    problem_id TEXT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    asset_type TEXT NOT NULL,
    name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    content_text TEXT NOT NULL,
    is_hidden INTEGER NOT NULL DEFAULT 0,
    metadata TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS problem_judge_configs (
    problem_id TEXT PRIMARY KEY REFERENCES problems(id) ON DELETE CASCADE,
    runner TEXT NOT NULL,
    config TEXT NOT NULL DEFAULT '{}',
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS problem_test_cases (
    id TEXT PRIMARY KEY DEFAULT (
        lower(hex(randomblob(4))) || '-' ||
        lower(hex(randomblob(2))) || '-4' ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        substr('89ab', abs(random()) % 4 + 1, 1) ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        lower(hex(randomblob(6)))
    ),
    problem_id TEXT NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    display_input TEXT NOT NULL DEFAULT '',
    display_expected TEXT NOT NULL DEFAULT '',
    explanation TEXT NOT NULL DEFAULT '',
    payload TEXT NOT NULL DEFAULT '{}',
    is_hidden INTEGER NOT NULL DEFAULT 0,
    weight INTEGER NOT NULL DEFAULT 1 CHECK (weight > 0),
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_problem_test_cases_problem_order
    ON problem_test_cases (problem_id, sort_order);

CREATE TABLE IF NOT EXISTS submissions (
    id TEXT PRIMARY KEY DEFAULT (
        lower(hex(randomblob(4))) || '-' ||
        lower(hex(randomblob(2))) || '-4' ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        substr('89ab', abs(random()) % 4 + 1, 1) ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        lower(hex(randomblob(6)))
    ),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    problem_id TEXT NOT NULL REFERENCES problems(id) ON DELETE RESTRICT,
    language TEXT NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('run', 'submit')),
    source_code TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed')),
    verdict TEXT NOT NULL CHECK (verdict IN ('pending', 'accepted', 'wrong_answer', 'compile_error', 'runtime_error', 'system_error')),
    score INTEGER NOT NULL DEFAULT 0,
    total_tests INTEGER NOT NULL DEFAULT 0,
    passed_tests INTEGER NOT NULL DEFAULT 0,
    compile_output TEXT NOT NULL DEFAULT '',
    runtime_output TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    queued_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_submissions_problem ON submissions (problem_id, queued_at DESC);
CREATE INDEX IF NOT EXISTS idx_submissions_user ON submissions (user_id, queued_at DESC);

CREATE TABLE IF NOT EXISTS submission_test_results (
    id TEXT PRIMARY KEY DEFAULT (
        lower(hex(randomblob(4))) || '-' ||
        lower(hex(randomblob(2))) || '-4' ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        substr('89ab', abs(random()) % 4 + 1, 1) ||
        substr(lower(hex(randomblob(2))), 2) || '-' ||
        lower(hex(randomblob(6)))
    ),
    submission_id TEXT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    test_case_id TEXT REFERENCES problem_test_cases(id) ON DELETE SET NULL,
    case_name TEXT NOT NULL,
    is_hidden INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL CHECK (status IN ('passed', 'failed', 'error', 'skipped')),
    message TEXT NOT NULL DEFAULT '',
    execution_ms INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_submission_results_submission_order
    ON submission_test_results (submission_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_problems_category_difficulty ON problems (category, difficulty);
CREATE INDEX IF NOT EXISTS idx_problems_slug ON problems (slug);
