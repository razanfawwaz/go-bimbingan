package middleware

import "net/http"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Authenticated-User") == "" {
			http.Redirect(w, r, "/login", http.StatusFound)
		}

		next.ServeHTTP(w, r)
	})
}
