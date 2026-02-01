package mw

import (
	"net/http"
	"strings"

	jwtutil "github.com/dunooo0ooo/wb-tech-l0/gateway/internal/jwt"
)

func Auth(secret string, roles ...string) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(h, "Bearer ")

			claims, err := jwtutil.Parse(secret, tokenStr)
			if err != nil || claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if !allowed[claims.Role] {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
