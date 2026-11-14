package httpapi

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mattbriggs04/bitforge/backend/internal/model"
	"github.com/mattbriggs04/bitforge/backend/internal/service"
)

type Server struct {
	problemService    *service.ProblemService
	submissionService *service.SubmissionService
	defaultUserHandle string
	devResetEnabled   bool
	resetDatabase     func(context.Context) error
}

func NewServer(
	problemService *service.ProblemService,
	submissionService *service.SubmissionService,
	defaultUserHandle string,
	devResetEnabled bool,
	resetDatabase func(context.Context) error,
) *Server {
	return &Server{
		problemService:    problemService,
		submissionService: submissionService,
		defaultUserHandle: defaultUserHandle,
		devResetEnabled:   devResetEnabled,
		resetDatabase:     resetDatabase,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/problems", s.handleListProblems)
	mux.HandleFunc("GET /api/v1/problems/{slug}", s.handleGetProblem)
	mux.HandleFunc("GET /api/v1/problems/{slug}/explanation", s.handleExplanation)
	mux.HandleFunc("POST /api/v1/submissions", s.handleCreateSubmission)
	mux.HandleFunc("GET /api/v1/submissions/{id}", s.handleGetSubmission)
	mux.HandleFunc("POST /api/v1/dev/reset", s.handleDevReset)

	return s.withMiddleware(mux)
}

func (s *Server) handleDevReset(w http.ResponseWriter, r *http.Request) {
	if !s.devResetEnabled || s.resetDatabase == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err := s.resetDatabase(r.Context()); err != nil {
		log.Printf("dev reset error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to reset database")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "reset"})
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
	s.annotateCompletion(r.Context(), r, items)

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
	completed, err := s.completedProblemIDs(r.Context(), r)
	if err == nil {
		problem.Completed = completed[problem.ID]
	} else {
		log.Printf("load problem completion error: %v", err)
	}

	writeJSON(w, http.StatusOK, problem)
}

func (s *Server) annotateCompletion(ctx context.Context, r *http.Request, items []model.ProblemSummary) {
	completed, err := s.completedProblemIDs(ctx, r)
	if err != nil {
		log.Printf("load problem completion error: %v", err)
		return
	}
	for i := range items {
		items[i].Completed = completed[items[i].ID]
	}
}

func (s *Server) completedProblemIDs(ctx context.Context, r *http.Request) (map[string]bool, error) {
	userHandle, userKey := s.identityFromRequest(r)
	return s.submissionService.CompletedProblemIDs(ctx, userHandle, userKey)
}

type createSubmissionRequest struct {
	ProblemSlug string `json:"problemSlug"`
	Language    string `json:"language"`
	Mode        string `json:"mode"`
	SourceCode  string `json:"sourceCode"`
}

func (s *Server) handleCreateSubmission(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 400000)
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

func (s *Server) handleExplanation(w http.ResponseWriter, r *http.Request) {
	handle, key := s.identityFromRequest(r)
	text, err := s.submissionService.Explanation(r.Context(), r.PathValue("slug"), r.URL.Query().Get("submissionId"), key, handle)
	if err != nil {
		if !s.writeServiceError(w, err) {
			writeError(w, 500, "could not load explanation")
		}
		return
	}
	writeJSON(w, 200, map[string]string{"explanation": text})
}
