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
	problemService    *service.ProblemService
	submissionService *service.SubmissionService
	defaultUserHandle string
}

func NewServer(problemService *service.ProblemService, submissionService *service.SubmissionService, defaultUserHandle string) *Server {
	return &Server{
		problemService:    problemService,
		submissionService: submissionService,
		defaultUserHandle: defaultUserHandle,
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

	userHandle := strings.TrimSpace(r.Header.Get("X-User-Handle"))
	if userHandle == "" {
		userHandle = s.defaultUserHandle
	}

	result, err := s.submissionService.Create(r.Context(), service.CreateSubmissionInput{
		ProblemSlug: req.ProblemSlug,
		Language:    req.Language,
		Mode:        req.Mode,
		SourceCode:  req.SourceCode,
		UserHandle:  userHandle,
	})
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "must be") || strings.Contains(err.Error(), "supported") {
			writeError(w, http.StatusBadRequest, err.Error())
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

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Handle")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
