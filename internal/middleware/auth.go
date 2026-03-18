package middleware

import (
	"context"
	"net/http"
)

type ctxKey int

const authKey ctxKey = 1

type AuthInfo struct {
	UserID   int
	Role     string
	Username string
}

func WithAuth(next http.Handler, store *SessionStore, cookieName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieName)
		if err == nil && c.Value != "" {
			if sess, ok := store.Get(c.Value); ok {
				ctx := context.WithValue(r.Context(), authKey, AuthInfo{UserID: sess.UserID, Role: sess.Role, Username: sess.Username})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := GetAuth(r); !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, ok := GetAuth(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if auth.Role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetAuth(r *http.Request) (AuthInfo, bool) {
	v := r.Context().Value(authKey)
	if v == nil {
		return AuthInfo{}, false
	}
	auth, ok := v.(AuthInfo)
	return auth, ok && auth.UserID != 0
}
