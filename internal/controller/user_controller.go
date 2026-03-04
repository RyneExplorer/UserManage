package controller

import (
	"UserManagement/internal/middleware"
	"UserManagement/internal/model"
	"UserManagement/internal/service"
	"UserManagement/internal/view"
	"net/http"
	"strconv"
)

type UserController struct {
	Users    *service.UserService
	Renderer *view.Renderer
}

type usersPageData struct {
	Auth    middleware.AuthInfo
	IsAdmin bool
	Users   []model.User
}

type userEditPageData struct {
	Auth    middleware.AuthInfo
	IsAdmin bool
	User    model.User
	Error   string
}

func (c *UserController) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	auth, _ := middleware.GetAuth(r)
	users, err := c.Users.ListAll(r.Context())
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	_ = c.Renderer.Render(w, "userList.html", usersPageData{
		Auth:    auth,
		IsAdmin: auth.Role == model.RoleAdmin,
		Users:   users,
	})
}

func (c *UserController) Edit(w http.ResponseWriter, r *http.Request) {
	auth, _ := middleware.GetAuth(r)

	switch r.Method {
	case http.MethodGet:
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		u, err := c.Users.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		if u == nil {
			http.NotFound(w, r)
			return
		}
		_ = c.Renderer.Render(w, "userEdit.html", userEditPageData{
			Auth:    auth,
			IsAdmin: true,
			User:    *u,
		})
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		id, _ := strconv.Atoi(r.FormValue("id"))
		status64, _ := strconv.ParseInt(r.FormValue("status"), 10, 8)
		u := model.User{
			ID:       id,
			Username: r.FormValue("username"),
			Role:     r.FormValue("role"),
			Status:   int8(status64),
		}

		if err := c.Users.UpdateUser(r.Context(), u); err != nil {
			_ = c.Renderer.Render(w, "userEdit.html", userEditPageData{
				Auth:    auth,
				IsAdmin: true,
				User:    u,
				Error:   "保存失败",
			})
			return
		}
		http.Redirect(w, r, "/users", http.StatusFound)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (c *UserController) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	if err := c.Users.DeleteUser(r.Context(), id); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/users", http.StatusFound)
}
