package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-ddd/internal/presentation/authguard"
)

func TestAuthenticationRateLimitIgnoresForwardedAddress(t *testing.T) {
	calls := 0
	h := limitAuth(authguard.New(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++ }))
	for i := 0; i < 21; i++ {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.RemoteAddr = "192.0.2.1:1234"
		r.Header.Set("X-Forwarded-For", string(rune('a'+i)))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if i < 20 && w.Code != http.StatusOK {
			t.Fatalf("attempt %d: %d", i, w.Code)
		}
		if i == 20 && w.Code != http.StatusTooManyRequests {
			t.Fatalf("limit: %d", w.Code)
		}
	}
	if calls != 20 {
		t.Fatalf("handler calls: %d", calls)
	}
}
func TestLimitBodyAndDecode(t *testing.T) {
	h := limitBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v map[string]any
		if decodeJSON(w, r, &v) {
			w.WriteHeader(http.StatusOK)
		}
	}))
	for _, tc := range []struct {
		name, body string
		chunked    bool
		want       int
	}{
		{"valid", `{"name":"value"}`, false, 200},
		{"multiple values", `{} {}`, false, 400},
		{"oversized", `{"name":"` + strings.Repeat("a", maxRequestBody) + `"}`, false, 413},
		{"oversized chunked", `{"name":"` + strings.Repeat("a", maxRequestBody) + `"}`, true, 413},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			if tc.chunked {
				r.ContentLength = -1
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("got %d want %d", w.Code, tc.want)
			}
		})
	}
}
