package controller

import (
	"UserManagement/internal/middleware"
	"UserManagement/internal/model/dto/request"
	"UserManagement/internal/service"
	"UserManagement/internal/view"
	"errors"
	"net/http"
)

type AuthController struct {
	Users      *service.UserService
	Sessions   *middleware.SessionStore
	CookieName string
	Renderer   *view.Renderer
}

type authPageData struct {
	Error   string
	Success bool
}

func NewAuthController(users *service.UserService, sessions *middleware.SessionStore, cookieName string, renderer *view.Renderer) *AuthController {
	return &AuthController{
		Users:      users,
		Sessions:   sessions,
		CookieName: cookieName,
		Renderer:   renderer,
	}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		_ = c.Renderer.Render(w, "login.html", authPageData{})
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			_ = c.Renderer.Render(w, "login.html", authPageData{Error: "表单解析失败"})
			return
		}

		req := request.LoginRequest{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}
		u, err := c.Users.Authenticate(r.Context(), req.Username, req.Password)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				_ = c.Renderer.Render(w, "login.html", authPageData{Error: "用户名或密码错误"})
				return
			}
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		sid := c.Sessions.Create(u.ID, u.Role, u.Username)
		http.SetCookie(w, &http.Cookie{
			Name:     c.CookieName,
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		_ = c.Renderer.Render(w, "register.html", authPageData{})
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			_ = c.Renderer.Render(w, "register.html", authPageData{Error: "表单解析失败"})
			return
		}

		req := request.RegisterRequest{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}
		_, err := c.Users.Register(r.Context(), req.Username, req.Password)
		if err != nil {
			if errors.Is(err, service.ErrUsernameTaken) {
				_ = c.Renderer.Render(w, "register.html", authPageData{Error: "用户名已存在"})
				return
			}
			_ = c.Renderer.Render(w, "register.html", authPageData{Error: "注册失败"})
			return
		}
		_ = c.Renderer.Render(w, "register.html", authPageData{Success: true})
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie(c.CookieName)
	if err == nil && cookie.Value != "" {
		c.Sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     c.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}
