package response

import "UserManagement/internal/model/entity"

type AuthView struct {
	UserID   int
	Role     string
	Username string
}

type UsersPageData struct {
	Auth      AuthView
	IsAdmin   bool
	Users     []entity.User
	Username  string
	Status    string
	EditUser  *entity.User
	EditError string
	Flash     string // success toast message
	// pagination
	Page      int
	PageSize  int
	Total     int
	TotalPage int
}

type UserEditPageData struct {
	Auth    AuthView
	IsAdmin bool
	User    entity.User
	Error   string
}
