package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	metrics "github.com/jnovack/simple-redirect/internal/metrics"
)

func TestRedirectStatusCodes(t *testing.T) {
	originalTarget := *target
	originalStatus := *status
	defer func() {
		*target = originalTarget
		*status = originalStatus
	}()

	tests := []struct {
		name             string
		statusCode       int
		requestTarget    string
		expectedLocation string
	}{
		{
			name:             "301 permanent redirect preserves path and query",
			statusCode:       http.StatusMovedPermanently,
			requestTarget:    "/docs/start?lang=en",
			expectedLocation: "https://example.com/base/docs/start?lang=en",
		},
		{
			name:             "302 temporary redirect preserves path and query",
			statusCode:       http.StatusFound,
			requestTarget:    "/docs/start?lang=en",
			expectedLocation: "https://example.com/base/docs/start?lang=en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*target = "https://example.com/base"
			*status = tt.statusCode
			metrics.Target = *target
			metrics.Status = *status
			atomic.StoreInt64(&metrics.HTTPRedirects, 0)
			atomic.StoreInt64(&metrics.HTTPSRedirects, 0)

			req := httptest.NewRequest(http.MethodGet, tt.requestTarget, nil)
			recorder := httptest.NewRecorder()

			redirect(recorder, req)

			resp := recorder.Result()
			if resp.StatusCode != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, resp.StatusCode)
			}

			if location := resp.Header.Get("Location"); location != tt.expectedLocation {
				t.Fatalf("expected Location %q, got %q", tt.expectedLocation, location)
			}

			if redirects := atomic.LoadInt64(&metrics.HTTPRedirects); redirects != 1 {
				t.Fatalf("expected 1 HTTP redirect, got %d", redirects)
			}

			if redirects := atomic.LoadInt64(&metrics.HTTPSRedirects); redirects != 0 {
				t.Fatalf("expected 0 HTTPS redirects, got %d", redirects)
			}
		})
	}
}
