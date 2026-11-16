CREATE TABLE interviews (
    slug TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE interview_problems (
    interview_slug TEXT NOT NULL REFERENCES interviews(slug) ON DELETE CASCADE,
    problem_id TEXT NOT NULL REFERENCES problems(id) ON DELETE RESTRICT,
    position INTEGER NOT NULL CHECK(position > 0),
    PRIMARY KEY(interview_slug, position),
    UNIQUE(interview_slug, problem_id)
);
CREATE INDEX idx_interview_problems_problem ON interview_problems(problem_id);
