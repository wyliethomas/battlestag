package middleware

import (
	"net/http"

	"agent-gateway/models"
)

// AuthMiddleware provides API key authentication
type AuthMiddleware struct {
	apiKey string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(apiKey string) *AuthMiddleware {
	return &AuthMiddleware{
		apiKey: apiKey,
	}
}

// Authenticate checks if the request has a valid API key
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get API key from header
		apiKey := r.Header.Get("X-API-Key")

		// Check if API key is valid
		if apiKey == "" {
			models.WriteError(w, http.StatusUnauthorized, "Missing API key")
			return
		}

		if apiKey != m.apiKey {
			models.WriteError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}

		// API key is valid, proceed to next handler
		next(w, r)
	}
}
