package db

import (
	"context"
	"database/sql"
	"fmt"
)

func ResetDevelopmentDatabase(ctx context.Context, conn *sql.DB) error {
	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}

	for _, table := range []string{
		"submission_test_results",
		"judge_jobs",
		"submissions",
		"problem_versions",
		"problem_test_cases",
		"problem_judge_configs",
		"problem_assets",
		"problem_language_templates",
		"problem_tags",
		"problems",
		"users",
		"schema_migrations",
	} {
		if _, err := conn.ExecContext(ctx, `DROP TABLE IF EXISTS `+table); err != nil {
			return fmt.Errorf("drop table %s: %w", table, err)
		}
	}

	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}
	if err := RunMigrations(ctx, conn); err != nil {
		return err
	}
	if err := SeedMVP(ctx, conn); err != nil {
		return err
	}
	return nil
}
