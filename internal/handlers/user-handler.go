package handlers

import (
	"context"
	"net/http"
	"strings"
)

type TokenKeyType string

const TokenKey TokenKeyType = "token"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if we have Authorization header
		authorization := r.Header.Get("Authorization")
		req := r
		if strings.HasPrefix(authorization, "Bearer ") {
			// extract token
			token := strings.TrimPrefix(authorization, "Bearer ")
			ctx := context.WithValue(r.Context(), TokenKey, token)
			req = r.WithContext(ctx)
		}
		next.ServeHTTP(w, req)
	})
}

func UserHandler() http.Handler {
	userHandler := http.NewServeMux()
	userHandler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Users"))
	})
	userHandler.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		token, ok := r.Context().Value(TokenKey).(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte("User ID: " + id + "\nToken: " + token))
	})

	return AuthMiddleware(userHandler)
}
