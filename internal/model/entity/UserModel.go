package entity

import "time"

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	Password   string    `json:"-"`
	Role       string    `json:"role"`
	Status     int8      `json:"status"`
	CreateTime time.Time `json:"created_at"`
	UpdateTime time.Time `json:"updated_at"`
}
