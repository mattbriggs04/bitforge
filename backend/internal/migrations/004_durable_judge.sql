-- New submissions retain the exact public specification, private tests and runner
-- configuration used at enqueue time. Historical attempts remain unversioned.
CREATE TABLE problem_versions (
    id TEXT PRIMARY KEY,
    problem_id TEXT NOT NULL REFERENCES problems(id) ON DELETE RESTRICT,
    snapshot TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE submissions ADD COLUMN problem_version_id TEXT REFERENCES problem_versions(id);
CREATE TABLE judge_jobs (
    submission_id TEXT PRIMARY KEY REFERENCES submissions(id) ON DELETE CASCADE,
    attempts INTEGER NOT NULL DEFAULT 0,
    lease_token TEXT,
    lease_until INTEGER NOT NULL DEFAULT 0,
    available_at INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_judge_jobs_available ON judge_jobs(available_at, lease_until);
-- Pre-migration queued work can be recovered, but cannot reconstruct old tests.
INSERT INTO judge_jobs(submission_id) SELECT id FROM submissions WHERE status IN ('queued','running');
