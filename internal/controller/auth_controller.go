package controller

import (
	"UserManagement/internal/middleware"
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
	Error string
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

		u, err := c.Users.Authenticate(r.Context(), r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				_ = c.Renderer.Render(w, "login.html", authPageData{Error: "用户名或密码错误"})
				return
			}
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		sid := c.Sessions.Create(u.ID, u.Role)
		http.SetCookie(w, &http.Cookie{
			Name:     c.CookieName,
			Value:    sid,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/users", http.StatusFound)
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

		_, err := c.Users.Register(r.Context(), r.FormValue("username"), r.FormValue("password"))
		if err != nil {
			if errors.Is(err, service.ErrUsernameTaken) {
				_ = c.Renderer.Render(w, "register.html", authPageData{Error: "用户名已存在"})
				return
			}
			_ = c.Renderer.Render(w, "register.html", authPageData{Error: "注册失败"})
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
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
