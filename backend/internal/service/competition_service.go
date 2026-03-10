package service

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/repository"
)

var allowedCompetitionModes = map[string]struct{}{
	"time_based":         {},
	"questions_complete": {},
	"code_golf":          {},
}

var allowedDifficultyPolicies = map[string]struct{}{
	"easy":        {},
	"medium":      {},
	"hard":        {},
	"random":      {},
	"progressive": {},
}

type CompetitionService struct {
	competitions      *repository.CompetitionsRepository
	users             *repository.UsersRepository
	defaultUserHandle string
}

type CreateCompetitionRoomInput struct {
	UserHandle       string         `json:"userHandle"`
	UserKey          string         `json:"userKey"`
	Name             string         `json:"name"`
	Mode             string         `json:"mode"`
	QuestionCount    int            `json:"questionCount"`
	DifficultyPolicy string         `json:"difficultyPolicy"`
	Metadata         map[string]any `json:"metadata"`
}

type JoinCompetitionRoomInput struct {
	UserHandle string `json:"userHandle"`
	UserKey    string `json:"userKey"`
	Code       string `json:"code"`
}

type DeleteCompetitionRoomInput struct {
	UserHandle string `json:"userHandle"`
	UserKey    string `json:"userKey"`
	Code       string `json:"code"`
}

func NewCompetitionService(
	competitions *repository.CompetitionsRepository,
	users *repository.UsersRepository,
	defaultUserHandle string,
) *CompetitionService {
	return &CompetitionService{
		competitions:      competitions,
		users:             users,
		defaultUserHandle: defaultUserHandle,
	}
}

func (s *CompetitionService) CreateRoom(ctx context.Context, input CreateCompetitionRoomInput) (*model.CompetitionRoom, error) {
	handle := strings.TrimSpace(input.UserHandle)
	if handle == "" {
		handle = s.defaultUserHandle
	}
	userID, err := s.users.EnsureIdentity(ctx, input.UserKey, handle)
	if err != nil {
		if errors.Is(err, repository.ErrHandleTaken) {
			return nil, newConflictError("username is already taken")
		}
		return nil, fmt.Errorf("ensure user: %w", err)
	}

	mode := strings.TrimSpace(strings.ToLower(input.Mode))
	if mode == "" {
		mode = "time_based"
	}
	if _, ok := allowedCompetitionModes[mode]; !ok {
		return nil, newInvalidError("mode must be one of: time_based, questions_complete, code_golf")
	}

	difficultyPolicy := strings.TrimSpace(strings.ToLower(input.DifficultyPolicy))
	if difficultyPolicy == "" {
		difficultyPolicy = "random"
	}
	if _, ok := allowedDifficultyPolicies[difficultyPolicy]; !ok {
		return nil, newInvalidError("difficultyPolicy must be one of: easy, medium, hard, random, progressive")
	}

	questionCount := input.QuestionCount
	if questionCount == 0 {
		questionCount = 5
	}
	if questionCount < 1 || questionCount > 100 {
		return nil, newInvalidError("questionCount must be between 1 and 100")
	}

	code, err := s.generateRoomCode(ctx)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = fmt.Sprintf("Room %s", code)
	}

	metadata := input.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["hostModel"] = "host-led-mvp"

	room, err := s.competitions.CreateRoom(ctx, model.NewCompetitionRoom{
		Code:             code,
		HostUserID:       userID,
		Name:             name,
		Mode:             mode,
		QuestionCount:    questionCount,
		DifficultyPolicy: difficultyPolicy,
		Status:           "open",
		Metadata:         metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}
	return room, nil
}

func (s *CompetitionService) JoinRoom(ctx context.Context, input JoinCompetitionRoomInput) (*model.CompetitionRoom, error) {
	handle := strings.TrimSpace(input.UserHandle)
	if handle == "" {
		handle = s.defaultUserHandle
	}
	userID, err := s.users.EnsureIdentity(ctx, input.UserKey, handle)
	if err != nil {
		if errors.Is(err, repository.ErrHandleTaken) {
			return nil, newConflictError("username is already taken")
		}
		return nil, fmt.Errorf("ensure user: %w", err)
	}

	code := normalizeRoomCode(input.Code)
	if code == "" {
		return nil, newInvalidError("room code is required")
	}

	room, err := s.competitions.JoinRoomByCode(ctx, code, userID)
	if err != nil {
		return nil, fmt.Errorf("join room: %w", err)
	}
	if room == nil {
		return nil, newNotFoundError("room not found")
	}
	return room, nil
}

func (s *CompetitionService) GetRoomByCode(ctx context.Context, code string) (*model.CompetitionRoom, error) {
	normalized := normalizeRoomCode(code)
	if normalized == "" {
		return nil, newInvalidError("room code is required")
	}
	room, err := s.competitions.GetRoomByCode(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("get room by code: %w", err)
	}
	return room, nil
}

func (s *CompetitionService) ListRoomsForUser(ctx context.Context, userKey, userHandle string) ([]model.CompetitionRoom, error) {
	handle := strings.TrimSpace(userHandle)
	if handle == "" {
		handle = s.defaultUserHandle
	}
	userID, err := s.users.EnsureIdentity(ctx, userKey, handle)
	if err != nil {
		if errors.Is(err, repository.ErrHandleTaken) {
			return nil, newConflictError("username is already taken")
		}
		return nil, fmt.Errorf("ensure user: %w", err)
	}
	rooms, err := s.competitions.ListRoomsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list rooms for user: %w", err)
	}
	return rooms, nil
}

func (s *CompetitionService) DeleteRoom(ctx context.Context, input DeleteCompetitionRoomInput) error {
	handle := strings.TrimSpace(input.UserHandle)
	if handle == "" {
		handle = s.defaultUserHandle
	}
	userID, err := s.users.EnsureIdentity(ctx, input.UserKey, handle)
	if err != nil {
		if errors.Is(err, repository.ErrHandleTaken) {
			return newConflictError("username is already taken")
		}
		return fmt.Errorf("ensure user: %w", err)
	}

	code := normalizeRoomCode(input.Code)
	if code == "" {
		return newInvalidError("room code is required")
	}

	room, err := s.competitions.GetRoomByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("get room by code: %w", err)
	}
	if room == nil {
		return newNotFoundError("room not found")
	}
	if room.HostUserID != userID {
		return newForbiddenError("only the host can delete this room")
	}

	deleted, err := s.competitions.DeleteRoomByCodeForHost(ctx, code, userID)
	if err != nil {
		return fmt.Errorf("delete room: %w", err)
	}
	if !deleted {
		return newNotFoundError("room not found")
	}
	return nil
}

func (s *CompetitionService) generateRoomCode(ctx context.Context) (string, error) {
	const attempts = 12
	for i := 0; i < attempts; i++ {
		code, err := randomHexCode(5)
		if err != nil {
			return "", fmt.Errorf("generate room code: %w", err)
		}
		exists, err := s.competitions.CodeExists(ctx, code)
		if err != nil {
			return "", fmt.Errorf("check room code availability: %w", err)
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("unable to allocate room code")
}

func randomHexCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid code length")
	}
	byteLen := (length + 1) / 2
	buf := make([]byte, byteLen)
	if _, err := crand.Read(buf); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(buf))[:length], nil
}

func normalizeRoomCode(input string) string {
	code := strings.TrimSpace(strings.ToUpper(input))
	if code == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(code))
	for _, r := range code {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
