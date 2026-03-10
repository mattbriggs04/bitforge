package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
)

type CompetitionsRepository struct {
	db *sql.DB
}

func NewCompetitionsRepository(db *sql.DB) *CompetitionsRepository {
	return &CompetitionsRepository{db: db}
}

func (r *CompetitionsRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM competition_rooms WHERE code = $1)`,
		code,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check room code exists: %w", err)
	}
	return exists, nil
}

func (r *CompetitionsRepository) CreateRoom(ctx context.Context, input model.NewCompetitionRoom) (*model.CompetitionRoom, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create room tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var roomID string
	const insertRoom = `
		INSERT INTO competition_rooms (
			code,
			host_user_id,
			name,
			mode,
			question_count,
			difficulty_policy,
			status,
			metadata,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, now())
		RETURNING id
	`
	if err := tx.QueryRowContext(
		ctx,
		insertRoom,
		input.Code,
		input.HostUserID,
		input.Name,
		input.Mode,
		input.QuestionCount,
		input.DifficultyPolicy,
		input.Status,
		mustJSONText(input.Metadata),
	).Scan(&roomID); err != nil {
		return nil, fmt.Errorf("insert room: %w", err)
	}

	const insertMember = `
		INSERT INTO competition_room_members (room_id, user_id, is_host)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (room_id, user_id)
		DO UPDATE SET is_host = EXCLUDED.is_host
	`
	if _, err := tx.ExecContext(ctx, insertMember, roomID, input.HostUserID); err != nil {
		return nil, fmt.Errorf("insert host room member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create room tx: %w", err)
	}

	return r.GetRoomByCode(ctx, input.Code)
}

func (r *CompetitionsRepository) JoinRoomByCode(ctx context.Context, code, userID string) (*model.CompetitionRoom, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin join room tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var roomID string
	if err := tx.QueryRowContext(
		ctx,
		`SELECT id FROM competition_rooms WHERE code = $1 AND status IN ('open', 'active')`,
		code,
	).Scan(&roomID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("load room by code: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO competition_room_members (room_id, user_id, is_host) VALUES ($1, $2, FALSE) ON CONFLICT (room_id, user_id) DO NOTHING`,
		roomID,
		userID,
	); err != nil {
		return nil, fmt.Errorf("insert room member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit join room tx: %w", err)
	}

	return r.GetRoomByCode(ctx, code)
}

func (r *CompetitionsRepository) GetRoomByCode(ctx context.Context, code string) (*model.CompetitionRoom, error) {
	const query = `
		SELECT r.id, r.code, r.host_user_id, host.handle, r.name, r.mode, r.question_count,
		       r.difficulty_policy, r.status, r.metadata, r.created_at, r.updated_at
		FROM competition_rooms r
		JOIN users host ON host.id = r.host_user_id
		WHERE r.code = $1
	`
	var room model.CompetitionRoom
	var metadataRaw []byte
	if err := r.db.QueryRowContext(ctx, query, code).Scan(
		&room.ID,
		&room.Code,
		&room.HostUserID,
		&room.HostHandle,
		&room.Name,
		&room.Mode,
		&room.QuestionCount,
		&room.DifficultyPolicy,
		&room.Status,
		&metadataRaw,
		&room.CreatedAt,
		&room.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get room by code: %w", err)
	}

	room.Metadata = map[string]any{}
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &room.Metadata); err != nil {
			return nil, fmt.Errorf("decode room metadata: %w", err)
		}
	}

	members, err := r.getMembersByRoomID(ctx, room.ID)
	if err != nil {
		return nil, err
	}
	room.Members = members

	return &room, nil
}

func (r *CompetitionsRepository) DeleteRoomByCodeForHost(ctx context.Context, code, hostUserID string) (bool, error) {
	res, err := r.db.ExecContext(
		ctx,
		`DELETE FROM competition_rooms WHERE code = $1 AND host_user_id = $2`,
		code,
		hostUserID,
	)
	if err != nil {
		return false, fmt.Errorf("delete room by code for host: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected delete room: %w", err)
	}
	return rows > 0, nil
}

func (r *CompetitionsRepository) ListRoomsForUser(ctx context.Context, userID string) ([]model.CompetitionRoom, error) {
	const query = `
		SELECT r.id, r.code, r.host_user_id, host.handle, r.name, r.mode, r.question_count,
		       r.difficulty_policy, r.status, r.metadata, r.created_at, r.updated_at
		FROM competition_rooms r
		JOIN competition_room_members m ON m.room_id = r.id
		JOIN users host ON host.id = r.host_user_id
		WHERE m.user_id = $1
		ORDER BY r.created_at DESC
		LIMIT 100
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list rooms for user: %w", err)
	}
	defer rows.Close()

	items := make([]model.CompetitionRoom, 0)
	for rows.Next() {
		var room model.CompetitionRoom
		var metadataRaw []byte
		if err := rows.Scan(
			&room.ID,
			&room.Code,
			&room.HostUserID,
			&room.HostHandle,
			&room.Name,
			&room.Mode,
			&room.QuestionCount,
			&room.DifficultyPolicy,
			&room.Status,
			&metadataRaw,
			&room.CreatedAt,
			&room.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room summary: %w", err)
		}

		room.Metadata = map[string]any{}
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &room.Metadata); err != nil {
				return nil, fmt.Errorf("decode room summary metadata: %w", err)
			}
		}
		room.Members = []model.CompetitionRoomMember{}
		items = append(items, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room summaries: %w", err)
	}
	return items, nil
}

func (r *CompetitionsRepository) getMembersByRoomID(ctx context.Context, roomID string) ([]model.CompetitionRoomMember, error) {
	const query = `
		SELECT m.user_id, u.handle, m.is_host, m.joined_at
		FROM competition_room_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.room_id = $1
		ORDER BY m.is_host DESC, m.joined_at ASC, u.handle ASC
	`
	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("query room members: %w", err)
	}
	defer rows.Close()

	items := make([]model.CompetitionRoomMember, 0)
	for rows.Next() {
		var item model.CompetitionRoomMember
		if err := rows.Scan(&item.UserID, &item.Handle, &item.IsHost, &item.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan room member: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room members: %w", err)
	}
	return items, nil
}

func mustJSONText(v any) string {
	if v == nil {
		return "{}"
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
