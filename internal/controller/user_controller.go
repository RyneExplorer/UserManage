package controller

import (
	"UserManagement/internal/middleware"
	"UserManagement/internal/model/dto/request"
	"UserManagement/internal/model/dto/response"
	"UserManagement/internal/model/entity"
	"UserManagement/internal/service"
	"UserManagement/internal/view"
	"net/http"
	"strconv"
)

const pageSize = 10

type UserController struct {
	Users    *service.UserService
	Renderer *view.Renderer
}

func NewUserController(users *service.UserService, renderer *view.Renderer) *UserController {
	return &UserController{Users: users, Renderer: renderer}
}

// setFlash writes a one-time flash message via cookie
func setFlash(w http.ResponseWriter, msg string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    msg,
		Path:     "/",
		MaxAge:   5,
		HttpOnly: true,
	})
}

// popFlash reads and clears the flash cookie
func popFlash(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie("flash")
	if err != nil {
		return ""
	}
	http.SetCookie(w, &http.Cookie{Name: "flash", Path: "/", MaxAge: -1})
	return c.Value
}

func totalPages(total, size int) int {
	if size <= 0 {
		return 1
	}
	p := total / size
	if total%size != 0 {
		p++
	}
	return p
}

func (c *UserController) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	auth, _ := middleware.GetAuth(r)
	authView := response.AuthView{UserID: auth.UserID, Role: auth.Role, Username: auth.Username}

	username := r.URL.Query().Get("username")
	statusParam := r.URL.Query().Get("status")
	pageParam, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if pageParam < 1 {
		pageParam = 1
	}

	var statusPtr *int8
	if statusParam != "" {
		if v, err := strconv.ParseInt(statusParam, 10, 8); err == nil {
			s := int8(v)
			statusPtr = &s
		}
	}

	users, total, err := c.Users.ListByFilterPaged(r.Context(), username, statusPtr, pageParam, pageSize)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	flash := popFlash(w, r)

	_ = c.Renderer.Render(w, "userList.html", response.UsersPageData{
		Auth:      authView,
		IsAdmin:   auth.Role == entity.RoleAdmin,
		Users:     users,
		Username:  username,
		Status:    statusParam,
		Flash:     flash,
		Page:      pageParam,
		PageSize:  pageSize,
		Total:     total,
		TotalPage: totalPages(total, pageSize),
	})
}

func (c *UserController) Edit(w http.ResponseWriter, r *http.Request) {
	auth, _ := middleware.GetAuth(r)
	authView := response.AuthView{UserID: auth.UserID, Role: auth.Role, Username: auth.Username}

	renderListWithModal := func(editUser *entity.User, editError string) {
		users, total, _ := c.Users.ListByFilterPaged(r.Context(), "", nil, 1, pageSize)
		_ = c.Renderer.Render(w, "userList.html", response.UsersPageData{
			Auth:      authView,
			IsAdmin:   auth.Role == entity.RoleAdmin,
			Users:     users,
			EditUser:  editUser,
			EditError: editError,
			Page:      1,
			PageSize:  pageSize,
			Total:     total,
			TotalPage: totalPages(total, pageSize),
		})
	}

	switch r.Method {
	case http.MethodGet:
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		if id == 0 {
			renderListWithModal(&entity.User{}, "")
			return
		}
		u, err := c.Users.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		if u == nil {
			http.NotFound(w, r)
			return
		}
		renderListWithModal(u, "")
		return
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		id, _ := strconv.Atoi(r.FormValue("id"))
		status64, _ := strconv.ParseInt(r.FormValue("status"), 10, 8)
		req := request.UserUpdateRequest{
			ID:       id,
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
			Role:     r.FormValue("role"),
			Status:   int8(status64),
		}
		u := entity.User{ID: req.ID, Username: req.Username, Role: req.Role, Status: req.Status}
		if err := c.Users.UpdateUser(r.Context(), u, req.Password); err != nil {
			renderListWithModal(&u, "保存失败: "+err.Error())
			return
		}
		setFlash(w, "edit:用户 "+req.Username+" 编辑成功")
		http.Redirect(w, r, "/users", http.StatusFound)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	auth, _ := middleware.GetAuth(r)
	authView := response.AuthView{UserID: auth.UserID, Role: auth.Role, Username: auth.Username}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	status64, _ := strconv.ParseInt(r.FormValue("status"), 10, 8)
	req := request.UserCreateRequest{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Role:     r.FormValue("role"),
		Status:   int8(status64),
	}

	if err := c.Users.CreateUser(r.Context(), req.Username, req.Password, req.Role, req.Status); err != nil {
		users, total, _ := c.Users.ListByFilterPaged(r.Context(), "", nil, 1, pageSize)
		newUser := &entity.User{Username: req.Username, Role: req.Role, Status: req.Status}
		_ = c.Renderer.Render(w, "userList.html", response.UsersPageData{
			Auth:      authView,
			IsAdmin:   auth.Role == entity.RoleAdmin,
			Users:     users,
			EditUser:  newUser,
			EditError: "创建失败: " + err.Error(),
			Page:      1,
			PageSize:  pageSize,
			Total:     total,
			TotalPage: totalPages(total, pageSize),
		})
		return
	}
	setFlash(w, "create:用户 "+req.Username+" 新建成功")
	http.Redirect(w, r, "/users", http.StatusFound)
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
	username := r.FormValue("username")
	if err := c.Users.DeleteUser(r.Context(), id); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	setFlash(w, "delete:用户 "+username+" 删除成功")
	http.Redirect(w, r, "/users", http.StatusFound)
}
