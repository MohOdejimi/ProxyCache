package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestShouldCacheResponse(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		status int
		headers map[string]string
		expected bool 
	}{
		{
			name: "GET 200 with no special headers should cache",
			method: http.MethodGet,
			status: http.StatusOK,
			headers: map[string]string{},
			expected: true,
		}, 
		{
			name: "POST should not cache",
			method: http.MethodPost,
			status: http.StatusOK,
			headers: map[string]string{},
			expected: false,
		},
		{
			name: "GET 404 should not cache",
			method: http.MethodGet,
			status: http.StatusNotFound,
			headers: map[string]string{},
			expected: false,
		},
		{
			name: "no cache header should not cache",
			method: http.MethodGet,
			status: http.StatusOK,
			headers: map[string]string{"Cache-Control": "no-cache"},
			expected: false,
		},
		{
			name: "Set-Cookie should not cache",
			method: http.MethodGet,
			status: http.StatusOK,
			headers: map[string]string{"Set-Cookie": "session=abc"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func (t *testing.T) {
			req := httptest.NewRequest(tt.method, "http://example.com", nil)

			resp := &http.Response{
				StatusCode: tt.status,
				Header: http.Header{},
			}
			for k, v := range tt.headers {
				resp.Header.Set(k, v)
			}

			result := shouldCacheResponse(req, resp)
			if result != tt.expected {
				t.Errorf("Expected shouldCacheResponse to be %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParseMaxAge(t *testing.T) {
    tests := []struct {
        name             string
        cacheControl     string
        expectedDuration time.Duration
        expectedFound    bool
    }{
        {
            name:             "simple max-age",
            cacheControl:     "max-age=300",
            expectedDuration: 300 * time.Second,
            expectedFound:    true,
        },
        {
            name:             "max-age with other directives",
            cacheControl:     "public, max-age=600, must-revalidate",
            expectedDuration: 600 * time.Second,
            expectedFound:    true,
        },
        {
            name:             "no max-age present",
            cacheControl:     "no-cache, no-store",
            expectedDuration: 0,
            expectedFound:    false,
        },
        {
            name:             "empty cache-control header",
            cacheControl:     "",
            expectedDuration: 0,
            expectedFound:    false,
        },
        {
            name:             "max-age of zero",
            cacheControl:     "max-age=0",
            expectedDuration: 0,
            expectedFound:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            duration, found := parseMaxAge(tt.cacheControl)

            if found != tt.expectedFound {
                t.Errorf("test %q: expected found=%v, got found=%v", tt.name, tt.expectedFound, found)
            }
            if duration != tt.expectedDuration {
                t.Errorf("test %q: expected duration=%v, got duration=%v", tt.name, tt.expectedDuration, duration)
            }
        })
    }
}