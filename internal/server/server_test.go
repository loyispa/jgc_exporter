package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/loyispa/jgc_exporter/internal/health"
	"github.com/prometheus/client_golang/prometheus"
)

func contains(s, sub string) bool { return strings.Contains(s, sub) }

func newTestServer() *Server {
	reg := prometheus.NewRegistry()
	mon := health.NewMonitor()
	return New(5898, reg, mon)
}

func TestMetricsEndpoint(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDashboardEndpoint(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}

	var dashboard health.Dashboard
	if err := json.NewDecoder(w.Body).Decode(&dashboard); err != nil {
		t.Fatalf("failed to decode dashboard: %v", err)
	}
	if dashboard.OverallStatus != "healthy" {
		t.Fatalf("expected healthy status, got %s", dashboard.OverallStatus)
	}
}

func TestUIEndpoint(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/ui", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !contains(body, "recsSection") || !contains(body, "recsContainer") {
		t.Fatalf("UI must contain recommendations section: recsSection=%v recsContainer=%v",
			contains(body, "recsSection"), contains(body, "recsContainer"))
	}
	if !contains(body, "Tuning Advice") {
		t.Fatal("UI must contain Tuning Advice heading")
	}
	if !contains(body, "gcEventTypeFilter") || !contains(body, ">Total</button>") {
		t.Fatal("UI must contain GC event type filter")
	}
	if !contains(body, "Throughput") || !contains(body, "throughputCanvas") {
		t.Fatal("UI must contain Throughput chart")
	}
	if !contains(body, "gcEventFilterSearch") || !contains(body, "gcDurationFilterSearch") {
		t.Fatal("UI must contain type filter search inputs")
	}
	if !contains(body, "GC Duration") || !contains(body, "gcDurationTypeFilter") {
		t.Fatal("UI must contain GC Duration chart and type filter")
	}
	if !contains(body, "gc-two-col-row") {
		t.Fatal("UI must place GC Events and GC Duration on one row")
	}
	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Fatalf("expected text/html, got %s", ct)
	}
	if w.Body.Len() == 0 {
		t.Fatal("expected non-empty body")
	}
	if !contains(body, "Auto-refresh: 5s") || !contains(body, "setInterval(refresh,5000)") {
		t.Fatal("UI refresh interval should be 5s")
	}
}

func TestRootRedirect(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "/ui" {
		t.Fatalf("expected redirect to /ui, got %s", loc)
	}
}

func TestNotFound(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAddr(t *testing.T) {
	srv := newTestServer()
	if srv.Addr() != ":5898" {
		t.Fatalf("expected :5898, got %s", srv.Addr())
	}
}

func TestHandler(t *testing.T) {
	srv := newTestServer()
	h := srv.Handler()
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestUISlashEndpoint(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest("GET", "/ui/", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestShutdownWithoutStart(t *testing.T) {
	srv := newTestServer()
	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown without start should succeed: %v", err)
	}
}
