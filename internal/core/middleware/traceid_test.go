package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestTraceID_GeneratesTraceparent(t *testing.T) {
	handler := TraceID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceparent := r.Header.Get("traceparent")
		if traceparent == "" {
			t.Error("traceparent not set in request")
		}

		// Validate W3C format: 00-{32 hex}-{16 hex}-01
		pattern := `^00-[0-9a-f]{32}-[0-9a-f]{16}-01$`
		matched, err := regexp.MatchString(pattern, traceparent)
		if err != nil {
			t.Fatalf("regex error: %v", err)
		}
		if !matched {
			t.Errorf("traceparent format invalid: %s", traceparent)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestTraceID_PropagatesExistingTraceparent(t *testing.T) {
	existingTrace := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"

	handler := TraceID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceparent := r.Header.Get("traceparent")
		if traceparent != existingTrace {
			t.Errorf("expected propagated traceparent %s, got %s", existingTrace, traceparent)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("traceparent", existingTrace)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestTraceID_GeneratesUniqueTraceIDs(t *testing.T) {
	traces := make(map[string]bool)
	iterations := 100

	handler := TraceID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceparent := r.Header.Get("traceparent")
		traces[traceparent] = true
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < iterations; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	if len(traces) != iterations {
		t.Errorf("expected %d unique trace IDs, got %d", iterations, len(traces))
	}
}
