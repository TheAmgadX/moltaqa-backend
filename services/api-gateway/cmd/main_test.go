package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiGateway_RootRoute(t *testing.T) {
	server := createServer()
	handler := server.Handler

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	expectedBody := "Hello, World!"
	if body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}

func TestApiGateway_CORSHeaders_AllowedOrigin(t *testing.T) {
	server := createServer()
	handler := server.Handler

	// Origin listed in allowedOrigins map ("http://localhost:3000")
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Since OPTIONS preflight is handled by CORSMiddleware, it should return 204 No Content
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin to be http://localhost:3000, got %q", allowOrigin)
	}

	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods to match allowed list, got %q", allowMethods)
	}
}

func TestApiGateway_CORSHeaders_DisallowedOrigin(t *testing.T) {
	server := createServer()
	handler := server.Handler

	// Origin NOT listed in allowedOrigins map ("http://example.com")
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Chi router returns 405 Method Not Allowed because OPTIONS is not configured on "/" route
	// and CORSMiddleware bypassed it because the origin was disallowed.
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}

	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("expected Access-Control-Allow-Origin to be empty for disallowed origin, got %q", allowOrigin)
	}
}
