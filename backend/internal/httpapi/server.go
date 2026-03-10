package httpapi

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/service"
)

type Server struct {
	problemService     *service.ProblemService
	submissionService  *service.SubmissionService
	competitionService *service.CompetitionService
	defaultUserHandle  string
}

func NewServer(
	problemService *service.ProblemService,
	submissionService *service.SubmissionService,
	competitionService *service.CompetitionService,
	defaultUserHandle string,
) *Server {
	return &Server{
		problemService:     problemService,
		submissionService:  submissionService,
		competitionService: competitionService,
		defaultUserHandle:  defaultUserHandle,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/problems", s.handleListProblems)
	mux.HandleFunc("GET /api/v1/problems/{slug}", s.handleGetProblem)
	mux.HandleFunc("POST /api/v1/submissions", s.handleCreateSubmission)
	mux.HandleFunc("GET /api/v1/submissions/{id}", s.handleGetSubmission)
	mux.HandleFunc("GET /api/v1/competitions/rooms", s.handleListCompetitionRooms)
	mux.HandleFunc("POST /api/v1/competitions/rooms", s.handleCreateCompetitionRoom)
	mux.HandleFunc("POST /api/v1/competitions/rooms/join", s.handleJoinCompetitionRoom)
	mux.HandleFunc("GET /api/v1/competitions/rooms/{code}", s.handleGetCompetitionRoom)
	mux.HandleFunc("DELETE /api/v1/competitions/rooms/{code}", s.handleDeleteCompetitionRoom)
	mux.HandleFunc("POST /api/v1/competitions/rooms/{code}/delete", s.handleDeleteCompetitionRoom)

	return s.withMiddleware(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleListProblems(w http.ResponseWriter, r *http.Request) {
	filter := model.ProblemFilter{
		Query:      strings.TrimSpace(r.URL.Query().Get("q")),
		Difficulty: strings.TrimSpace(r.URL.Query().Get("difficulty")),
		Category:   strings.TrimSpace(r.URL.Query().Get("category")),
		Tag:        strings.TrimSpace(r.URL.Query().Get("tag")),
	}

	items, err := s.problemService.List(r.Context(), filter)
	if err != nil {
		log.Printf("list problems error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list problems")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleGetProblem(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if slug == "" {
		writeError(w, http.StatusBadRequest, "problem slug is required")
		return
	}

	problem, err := s.problemService.GetBySlug(r.Context(), slug)
	if err != nil {
		log.Printf("get problem error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load problem")
		return
	}
	if problem == nil {
		writeError(w, http.StatusNotFound, "problem not found")
		return
	}

	writeJSON(w, http.StatusOK, problem)
}

type createSubmissionRequest struct {
	ProblemSlug string `json:"problemSlug"`
	Language    string `json:"language"`
	Mode        string `json:"mode"`
	SourceCode  string `json:"sourceCode"`
}

func (s *Server) handleCreateSubmission(w http.ResponseWriter, r *http.Request) {
	var req createSubmissionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	userHandle, userKey := s.identityFromRequest(r)

	result, err := s.submissionService.Create(r.Context(), service.CreateSubmissionInput{
		ProblemSlug: req.ProblemSlug,
		Language:    req.Language,
		Mode:        req.Mode,
		SourceCode:  req.SourceCode,
		UserHandle:  userHandle,
		UserKey:     userKey,
	})
	if err != nil {
		if s.writeServiceError(w, err) {
			return
		}
		log.Printf("create submission error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create submission")
		return
	}

	writeJSON(w, http.StatusAccepted, result)
}

func (s *Server) handleGetSubmission(w http.ResponseWriter, r *http.Request) {
	submissionID := strings.TrimSpace(r.PathValue("id"))
	if submissionID == "" {
		writeError(w, http.StatusBadRequest, "submission id is required")
		return
	}

	submission, err := s.submissionService.GetByID(r.Context(), submissionID)
	if err != nil {
		log.Printf("get submission error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load submission")
		return
	}
	if submission == nil {
		writeError(w, http.StatusNotFound, "submission not found")
		return
	}

	writeJSON(w, http.StatusOK, submission)
}

type createCompetitionRoomRequest struct {
	Name             string         `json:"name"`
	Mode             string         `json:"mode"`
	QuestionCount    int            `json:"questionCount"`
	DifficultyPolicy string         `json:"difficultyPolicy"`
	Metadata         map[string]any `json:"metadata"`
}

type joinCompetitionRoomRequest struct {
	Code string `json:"code"`
}

func (s *Server) handleListCompetitionRooms(w http.ResponseWriter, r *http.Request) {
	userHandle, userKey := s.identityFromRequest(r)

	rooms, err := s.competitionService.ListRoomsForUser(r.Context(), userKey, userHandle)
	if err != nil {
		if s.writeServiceError(w, err) {
			return
		}
		log.Printf("list competition rooms error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list competition rooms")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": rooms})
}

func (s *Server) handleCreateCompetitionRoom(w http.ResponseWriter, r *http.Request) {
	var req createCompetitionRoomRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	userHandle, userKey := s.identityFromRequest(r)

	room, err := s.competitionService.CreateRoom(r.Context(), service.CreateCompetitionRoomInput{
		UserHandle:       userHandle,
		UserKey:          userKey,
		Name:             req.Name,
		Mode:             req.Mode,
		QuestionCount:    req.QuestionCount,
		DifficultyPolicy: req.DifficultyPolicy,
		Metadata:         req.Metadata,
	})
	if err != nil {
		if s.writeServiceError(w, err) {
			return
		}
		log.Printf("create competition room error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create competition room")
		return
	}

	writeJSON(w, http.StatusCreated, room)
}

func (s *Server) handleJoinCompetitionRoom(w http.ResponseWriter, r *http.Request) {
	var req joinCompetitionRoomRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	userHandle, userKey := s.identityFromRequest(r)

	room, err := s.competitionService.JoinRoom(r.Context(), service.JoinCompetitionRoomInput{
		UserHandle: userHandle,
		UserKey:    userKey,
		Code:       req.Code,
	})
	if err != nil {
		if s.writeServiceError(w, err) {
			return
		}
		log.Printf("join competition room error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to join competition room")
		return
	}

	writeJSON(w, http.StatusOK, room)
}

func (s *Server) handleGetCompetitionRoom(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.PathValue("code"))
	if code == "" {
		writeError(w, http.StatusBadRequest, "room code is required")
		return
	}

	room, err := s.competitionService.GetRoomByCode(r.Context(), code)
	if err != nil {
		if s.writeServiceError(w, err) {
			return
		}
		log.Printf("get competition room error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load competition room")
		return
	}
	if room == nil {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}

	writeJSON(w, http.StatusOK, room)
}

func (s *Server) handleDeleteCompetitionRoom(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(r.PathValue("code"))
	if code == "" {
		writeError(w, http.StatusBadRequest, "room code is required")
		return
	}

	userHandle, userKey := s.identityFromRequest(r)

	if err := s.competitionService.DeleteRoom(r.Context(), service.DeleteCompetitionRoomInput{
		UserHandle: userHandle,
		UserKey:    userKey,
		Code:       code,
	}); err != nil {
		if s.writeServiceError(w, err) {
			return
		}
		log.Printf("delete competition room error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete competition room")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func (s *Server) writeServiceError(w http.ResponseWriter, err error) bool {
	appErr, ok := service.AsAppError(err)
	if !ok {
		return false
	}

	status := http.StatusBadRequest
	if appErr.Kind == service.ErrorKindNotFound {
		status = http.StatusNotFound
	} else if appErr.Kind == service.ErrorKindForbidden {
		status = http.StatusForbidden
	} else if appErr.Kind == service.ErrorKindConflict {
		status = http.StatusConflict
	}
	writeError(w, status, appErr.Message)
	return true
}

func (s *Server) identityFromRequest(r *http.Request) (string, string) {
	userHandle := strings.TrimSpace(r.Header.Get("X-User-Handle"))
	if userHandle == "" {
		userHandle = s.defaultUserHandle
	}
	userKey := strings.TrimSpace(r.Header.Get("X-User-Key"))
	return userHandle, userKey
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Handle, X-User-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, recorder.status, time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
