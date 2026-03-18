package service

import (
	"UserManagement/internal/repository"
	"context"
	"errors"
	"strings"

	"UserManagement/internal/model/entity"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("用户名或密码错误")
var ErrUsernameTaken = errors.New("用户名已被占用")

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, username, password string) (int, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return 0, errors.New("用户名和密码不能为空")
	}

	existing, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, ErrUsernameTaken
	}

	role := entity.RoleUser
	count, err := s.repo.Count(ctx)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		// 首个用户自动设为管理员
		role = entity.RoleAdmin
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	return s.repo.Create(ctx, entity.User{
		Username: username,
		Password: string(hash),
		Role:     role,
		Status:   1,
	})
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (*entity.User, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	u, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidCredentials
	}
	if u.Status == 0 {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	_ = s.repo.UpdateLastLogin(ctx, u.ID)
	return u, nil
}

func (s *UserService) ListAll(ctx context.Context) ([]entity.User, error) {
	return s.repo.ListAll(ctx)
}

func (s *UserService) ListByFilter(ctx context.Context, username string, status *int8) ([]entity.User, error) {
	username = strings.TrimSpace(username)
	return s.repo.ListByFilter(ctx, username, status)
}

func (s *UserService) ListByFilterPaged(ctx context.Context, username string, status *int8, page, pageSize int) ([]entity.User, int, error) {
	username = strings.TrimSpace(username)
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return s.repo.ListByFilterPaged(ctx, username, status, offset, pageSize)
}

func (s *UserService) GetByID(ctx context.Context, id int) (*entity.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, username, password, role string, status int8) error {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return errors.New("用户名和密码不能为空")
	}
	if role != entity.RoleAdmin && role != entity.RoleUser {
		return errors.New("无效的角色")
	}

	existing, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.repo.Create(ctx, entity.User{
		Username: username,
		Password: string(hash),
		Role:     role,
		Status:   status,
	})
	return err
}

func (s *UserService) UpdateUser(ctx context.Context, user entity.User, newPassword string) error {
	user.Username = strings.TrimSpace(user.Username)
	if user.ID == 0 || user.Username == "" {
		return errors.New("无效的用户名")
	}
	if user.Role != entity.RoleAdmin && user.Role != entity.RoleUser {
		return errors.New("无效的角色")
	}
	if user.Status != 0 && user.Status != 1 {
		return errors.New("无效的状态")
	}
	if newPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hash)
		return s.repo.UpdateWithPassword(ctx, user)
	}
	return s.repo.Update(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if id == 0 {
		return errors.New("无效的用户id")
	}
	return s.repo.Delete(ctx, id)
}
