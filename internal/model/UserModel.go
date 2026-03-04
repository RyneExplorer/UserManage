package model

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
	CreateTime time.Time `json:"create_time"`
	LastTime   time.Time `json:"last_time"`
}
