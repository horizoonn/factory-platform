package swaggerui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler_ServeUI(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/docs/", nil)

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	require.Contains(t, recorder.Body.String(), "Test API")
	require.Contains(t, recorder.Body.String(), `url: "/docs/openapi.json"`)
	require.Contains(t, recorder.Body.String(), "swagger-ui-dist@"+swaggerUIVersion)
}

func TestHandler_ServeSpec(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	require.JSONEq(t, `{"openapi":"3.0.3"}`, recorder.Body.String())
}

func TestHandler_NotFound(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/docs/unknown", nil)

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	handler := newTestHandler()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/docs/", strings.NewReader("body"))

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
	require.Equal(t, "GET, HEAD", recorder.Header().Get("Allow"))
}

func newTestHandler() *Handler {
	return NewHandler(Config{
		Title:           "Test API",
		UIPath:          "/docs/",
		SpecPath:        "/docs/openapi.json",
		Spec:            `{"openapi":"3.0.3"}`,
		SpecContentType: "application/json",
	})
}
