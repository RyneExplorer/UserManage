package router

import (
	"UserManagement/internal/config"
	"UserManagement/internal/controller"
	"UserManagement/internal/middleware"
	"net/http"
)

func NewMux(cfg config.Config, auth *controller.AuthController, users *controller.UserController, sessions *middleware.SessionStore) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := middleware.GetAuth(r); ok {
			http.Redirect(w, r, "/users", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	mux.HandleFunc("/login", auth.Login)
	mux.HandleFunc("/register", auth.Register)
	mux.HandleFunc("/logout", auth.Logout)

	mux.Handle("/users", middleware.RequireLogin(http.HandlerFunc(users.List)))
	mux.Handle("/users/edit", middleware.RequireAdmin(http.HandlerFunc(users.Edit)))
	mux.Handle("/users/delete", middleware.RequireAdmin(http.HandlerFunc(users.Delete)))

	return middleware.WithAuth(mux, sessions, cfg.SessionCookieName)
}
