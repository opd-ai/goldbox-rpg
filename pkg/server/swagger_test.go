package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCServer_ServeSwaggerUI(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		checkBody      bool
	}{
		{
			name:           "SuccessfulRender",
			expectedStatus: http.StatusOK,
			checkBody:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &RPCServer{}
			req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
			w := httptest.NewRecorder()

			server.ServeSwaggerUI(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkBody {
				assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
				assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}

func TestRPCServer_ServeOpenAPISpec(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "api", "openapi.yaml")

	tests := []struct {
		name           string
		setupFunc      func()
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name: "SpecFileExists",
			setupFunc: func() {
				require.NoError(t, os.MkdirAll(filepath.Dir(specPath), 0755))
				require.NoError(t, os.WriteFile(specPath, []byte("openapi: 3.0.0\n"), 0644))

				oldWd, err := os.Getwd()
				require.NoError(t, err)
				t.Cleanup(func() { os.Chdir(oldWd) })
				require.NoError(t, os.Chdir(tmpDir))
			},
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			server := &RPCServer{}
			req := httptest.NewRequest(http.MethodGet, "/api/openapi.yaml", nil)
			w := httptest.NewRecorder()

			server.ServeOpenAPISpec(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkHeaders {
				assert.Equal(t, "application/x-yaml", w.Header().Get("Content-Type"))
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}
