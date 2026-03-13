package repository

import (
	"UserManagement/internal/model"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) (int, error)
	Count(ctx context.Context) (int, error)
	FindByID(ctx context.Context, id int) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	ListAll(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, user model.User) error
	UpdateLastLogin(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
}
