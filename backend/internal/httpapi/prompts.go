package httpapi

import (
	"io"
	"net/http"

	"github.com/mattbriggs04/bitforge/backend/internal/authoring"
	"github.com/mattbriggs04/bitforge/backend/internal/curriculum"
)

func (s *Server) handlePromptCatalog(w http.ResponseWriter, r *http.Request) {
	defaults := map[string]map[string][]curriculum.Mix{}
	for _, role := range curriculum.Roles {
		defaults[role.ID] = map[string][]curriculum.Mix{}
		for _, preset := range curriculum.Presets {
			defaults[role.ID][preset.ID] = curriculum.Defaults(role.ID, preset.ID)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"defaults": defaults, "roles": curriculum.Roles, "presets": curriculum.Presets, "formats": curriculum.Formats, "focuses": curriculum.Focuses})
}

// Compilation is pure: no LLM call, file writes, execution, or publication.
func (s *Server) handleCompilePrompt(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 32*1024))
	if err != nil {
		writeError(w, http.StatusBadRequest, "request exceeds 32 KiB")
		return
	}
	var config authoring.Config
	if err = authoring.Decode(data, &config); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	compiled, err := authoring.Compile(config)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, compiled)
}
