package router

import (
	"UserManagement/internal/config"
	"UserManagement/internal/controller"
	"UserManagement/internal/middleware"
	"UserManagement/internal/service"
	"UserManagement/internal/view"
	"net/http"
)

type Router struct {
	cfg      config.Config
	sessions *middleware.SessionStore
	auth     *controller.AuthController
	users    *controller.UserController
}

func NewRouter(cfg config.Config,
	users *service.UserService,
	sessions *middleware.SessionStore,
	renderer *view.Renderer,
) *Router {
	return &Router{
		cfg:      cfg,
		sessions: sessions,
		auth:     controller.NewAuthController(users, sessions, cfg.SessionCookieName, renderer),
		users:    controller.NewUserController(users, renderer),
	}
}

func (r *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		if _, ok := middleware.GetAuth(req); ok {
			http.Redirect(w, req, "/dashboard", http.StatusFound)
			return
		}
		http.Redirect(w, req, "/login", http.StatusFound)
	})

	mux.Handle("/dashboard", middleware.RequireLogin(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "static/index.html")
	})))

	mux.HandleFunc("/login", r.auth.Login)
	mux.HandleFunc("/register", r.auth.Register)
	mux.HandleFunc("/logout", r.auth.Logout)

	mux.Handle("/users", middleware.RequireLogin(http.HandlerFunc(r.users.List)))
	mux.Handle("/users/edit", middleware.RequireAdmin(http.HandlerFunc(r.users.Edit)))
	mux.Handle("/users/create", middleware.RequireAdmin(http.HandlerFunc(r.users.CreateUser)))
	mux.Handle("/users/delete", middleware.RequireAdmin(http.HandlerFunc(r.users.Delete)))

	return middleware.WithAuth(mux, r.sessions, r.cfg.SessionCookieName)
}
