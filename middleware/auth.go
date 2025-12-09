package middleware

import (
	"net/http"
	"os"
)

// APIKeyMiddleware checks for a specific header
func APIKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the key from the header
		// Standard is "Authorization: Bearer <key>", but we will use "X-API-Key" for simplicity
		key := r.Header.Get("X-API-Key")

		// 2. Get the real key from .env
		realKey := os.Getenv("API_KEY")

		// 3. Compare
		if realKey == "" {
			// Safety: If no key is set in .env, block everything
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		if key != realKey {
			http.Error(w, "Unauthorized: Invalid API Key", http.StatusUnauthorized)
			return
		}

		// 4. Pass to the next handler
		next.ServeHTTP(w, r)
	})
}
