package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "GET request with CORS headers",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "OPTIONS preflight request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			rr := httptest.NewRecorder()

			handler := CORS(testHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.checkHeaders {
				// Check CORS headers
				if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
					t.Errorf("Access-Control-Allow-Origin header incorrect: got %v want *", origin)
				}

				if methods := rr.Header().Get("Access-Control-Allow-Methods"); methods == "" {
					t.Error("Access-Control-Allow-Methods header not set")
				}

				if headers := rr.Header().Get("Access-Control-Allow-Headers"); headers == "" {
					t.Error("Access-Control-Allow-Headers header not set")
				}
			}
		})
	}
}
