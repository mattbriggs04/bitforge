package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type UsersRepository struct {
	db *sql.DB
}

var ErrHandleTaken = errors.New("username is already taken")

func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) EnsureByHandle(ctx context.Context, handle string) (string, error) {
	const upsert = `
		INSERT INTO users (handle)
		VALUES ($1)
		ON CONFLICT (handle) DO UPDATE SET handle = EXCLUDED.handle
		RETURNING id
	`
	var id string
	if err := r.db.QueryRowContext(ctx, upsert, handle).Scan(&id); err != nil {
		return "", fmt.Errorf("ensure user by handle: %w", err)
	}
	return id, nil
}

func (r *UsersRepository) EnsureIdentity(ctx context.Context, clientKey, handle string) (string, error) {
	normalizedHandle := strings.TrimSpace(handle)
	if normalizedHandle == "" {
		normalizedHandle = "demo"
	}

	normalizedClientKey := strings.TrimSpace(clientKey)
	if normalizedClientKey == "" {
		return r.EnsureByHandle(ctx, normalizedHandle)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin ensure identity tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	current, err := getUserByClientKeyTx(ctx, tx, normalizedClientKey)
	if err != nil {
		return "", fmt.Errorf("query user by client key: %w", err)
	}
	if current != nil {
		id, assignErr := r.assignHandleToUserTx(ctx, tx, current.ID, normalizedClientKey, normalizedHandle)
		if assignErr != nil {
			return "", assignErr
		}
		if err = tx.Commit(); err != nil {
			return "", fmt.Errorf("commit ensure identity tx: %w", err)
		}
		committed = true
		return id, nil
	}

	byHandle, err := getUserByHandleTx(ctx, tx, normalizedHandle)
	if err != nil {
		return "", fmt.Errorf("query user by handle: %w", err)
	}
	if byHandle != nil {
		if byHandle.ClientKey.Valid && strings.TrimSpace(byHandle.ClientKey.String) != "" && byHandle.ClientKey.String != normalizedClientKey {
			return "", ErrHandleTaken
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE users SET client_key = $2, handle = $3 WHERE id = $1`,
			byHandle.ID,
			normalizedClientKey,
			normalizedHandle,
		); err != nil {
			return "", fmt.Errorf("claim user by handle: %w", err)
		}
		if err = tx.Commit(); err != nil {
			return "", fmt.Errorf("commit ensure identity tx: %w", err)
		}
		committed = true
		return byHandle.ID, nil
	}

	var id string
	if err := tx.QueryRowContext(
		ctx,
		`INSERT INTO users (client_key, handle) VALUES ($1, $2) RETURNING id`,
		normalizedClientKey,
		normalizedHandle,
	).Scan(&id); err != nil {
		if isUniqueHandleError(err) {
			return "", ErrHandleTaken
		}
		return "", fmt.Errorf("insert user identity: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("commit ensure identity tx: %w", err)
	}
	committed = true
	return id, nil
}

type userRecord struct {
	ID        string
	Handle    string
	ClientKey sql.NullString
}

func getUserByClientKeyTx(ctx context.Context, tx *sql.Tx, clientKey string) (*userRecord, error) {
	var row userRecord
	if err := tx.QueryRowContext(
		ctx,
		`SELECT id, handle, client_key FROM users WHERE client_key = $1 FOR UPDATE`,
		clientKey,
	).Scan(&row.ID, &row.Handle, &row.ClientKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func getUserByHandleTx(ctx context.Context, tx *sql.Tx, handle string) (*userRecord, error) {
	var row userRecord
	if err := tx.QueryRowContext(
		ctx,
		`SELECT id, handle, client_key FROM users WHERE handle = $1 FOR UPDATE`,
		handle,
	).Scan(&row.ID, &row.Handle, &row.ClientKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *UsersRepository) assignHandleToUserTx(ctx context.Context, tx *sql.Tx, userID, clientKey, handle string) (string, error) {
	owner, err := getUserByHandleTx(ctx, tx, handle)
	if err != nil {
		return "", fmt.Errorf("query handle owner: %w", err)
	}

	if owner == nil {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE users SET handle = $2, client_key = $3 WHERE id = $1`,
			userID,
			handle,
			clientKey,
		); err != nil {
			return "", fmt.Errorf("update user handle: %w", err)
		}
		return userID, nil
	}

	if owner.ID == userID {
		return userID, nil
	}

	ownerKey := strings.TrimSpace(owner.ClientKey.String)
	if ownerKey != "" && ownerKey != clientKey {
		return "", ErrHandleTaken
	}

	// Legacy row with no client key can be merged into the current identity row.
	if ownerKey == "" {
		if err := mergeUsersTx(ctx, tx, owner.ID, userID); err != nil {
			return "", fmt.Errorf("merge legacy user identity: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE users SET handle = $2, client_key = $3 WHERE id = $1`,
			userID,
			handle,
			clientKey,
		); err != nil {
			return "", fmt.Errorf("update merged user handle: %w", err)
		}
		return userID, nil
	}

	return "", ErrHandleTaken
}

func mergeUsersTx(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	if fromUserID == toUserID {
		return nil
	}

	for _, stmt := range []struct {
		query string
		args  []any
	}{
		{
			query: `UPDATE submissions SET user_id = $1 WHERE user_id = $2`,
			args:  []any{toUserID, fromUserID},
		},
		{
			query: `UPDATE competition_rooms SET host_user_id = $1 WHERE host_user_id = $2`,
			args:  []any{toUserID, fromUserID},
		},
		{
			query: `
				INSERT INTO competition_room_members (room_id, user_id, is_host, joined_at)
				SELECT room_id, $1, is_host, joined_at
				FROM competition_room_members
				WHERE user_id = $2
				ON CONFLICT (room_id, user_id)
				DO UPDATE SET
					is_host = competition_room_members.is_host OR EXCLUDED.is_host,
					joined_at = LEAST(competition_room_members.joined_at, EXCLUDED.joined_at)
			`,
			args: []any{toUserID, fromUserID},
		},
		{
			query: `DELETE FROM competition_room_members WHERE user_id = $1`,
			args:  []any{fromUserID},
		},
		{
			query: `DELETE FROM users WHERE id = $1`,
			args:  []any{fromUserID},
		},
	} {
		if _, err := tx.ExecContext(ctx, stmt.query, stmt.args...); err != nil {
			return err
		}
	}
	return nil
}

func isUniqueHandleError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == "users_handle_key"
	}
	return false
}
