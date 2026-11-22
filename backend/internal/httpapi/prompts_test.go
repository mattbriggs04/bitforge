package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattbriggs04/bitforge/backend/internal/authoring"
)

func TestCompileEndpointIsStrictAndRepeatable(t *testing.T) {
	handler := NewServer(nil, nil, "demo", false, nil).Handler()
	config := `{"schemaVersion":1,"kind":"interview","role":"ai_ml","preset":"daily","category":"random","language":"python","format":"random","difficulty":"medium","targetMinutes":25,"seed":17}`
	var first string
	for i := 0; i < 2; i++ {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/prompts/compile", strings.NewReader(config)))
		if recorder.Code != 200 {
			t.Fatal(recorder.Code, recorder.Body.String())
		}
		var compiled authoring.CompiledPrompt
		if err := json.Unmarshal(recorder.Body.Bytes(), &compiled); err != nil {
			t.Fatal(err)
		}
		if compiled.Config.Role != "ai_ml" || len(compiled.Config.Mix) != 2 {
			t.Fatal(compiled.Config)
		}
		if i == 0 {
			first = recorder.Body.String()
		} else if first != recorder.Body.String() {
			t.Fatal("same request produced different prompt")
		}
	}
	for _, body := range []string{config + ` {}`, strings.Replace(config, `"seed":17`, `"seed":17,"publish":true`, 1), strings.Repeat("x", 33000), strings.Replace(config, `"targetMinutes":25`, `"targetMinutes":0`, 1)} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/prompts/compile", strings.NewReader(body)))
		if recorder.Code != 400 {
			t.Fatalf("invalid config status: %d", recorder.Code)
		}
	}
}
