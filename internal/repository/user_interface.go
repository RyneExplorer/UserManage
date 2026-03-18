package repository

import (
	"UserManagement/internal/model/entity"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user entity.User) (int, error)
	Count(ctx context.Context) (int, error)
	FindByID(ctx context.Context, id int) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	ListAll(ctx context.Context) ([]entity.User, error)
	ListByFilter(ctx context.Context, username string, status *int8) ([]entity.User, error)
	ListByFilterPaged(ctx context.Context, username string, status *int8, offset, limit int) ([]entity.User, int, error)
	Update(ctx context.Context, user entity.User) error
	UpdateWithPassword(ctx context.Context, user entity.User) error
	UpdateLastLogin(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
}
