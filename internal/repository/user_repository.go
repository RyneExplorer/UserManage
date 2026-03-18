package repository

import (
	"UserManagement/internal/model/entity"
	"context"
	"database/sql"
	"errors"
	"strings"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user entity.User) (int, error) {
	res, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (username, password, role, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, NOW(), NULL)`,
		user.Username,
		user.Password,
		user.Role,
		user.Status,
	)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id64), nil
}

func (r *userRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM users`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*entity.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, password, role, status, created_at, COALESCE(updated_at, created_at)
		 FROM users WHERE id = ?`,
		id,
	)

	var u entity.User
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreateTime, &u.UpdateTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, password, role, status, created_at, COALESCE(updated_at, created_at)
		 FROM users WHERE username = ?`,
		username,
	)

	var u entity.User
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreateTime, &u.UpdateTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) ListAll(ctx context.Context) ([]entity.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, username, password, role, status, created_at, COALESCE(updated_at, created_at)
		 FROM users ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreateTime, &u.UpdateTime); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *userRepository) ListByFilter(ctx context.Context, username string, status *int8) ([]entity.User, error) {
	query := `SELECT id, username, password, role, status, created_at, COALESCE(updated_at, created_at) FROM users`
	var conditions []string
	var args []any
	if username != "" {
		conditions = append(conditions, "username LIKE ?")
		args = append(args, "%"+username+"%")
	}
	if status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *status)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreateTime, &u.UpdateTime); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *userRepository) Update(ctx context.Context, user entity.User) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE users SET username = ?, role = ?, status = ? WHERE id = ?`,
		user.Username, user.Role, user.Status, user.ID,
	)
	return err
}

func (r *userRepository) UpdateWithPassword(ctx context.Context, user entity.User) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE users SET username = ?, password = ?, role = ?, status = ? WHERE id = ?`,
		user.Username, user.Password, user.Role, user.Status, user.ID,
	)
	return err
}

func (r *userRepository) ListByFilterPaged(ctx context.Context, username string, status *int8, offset, limit int) ([]entity.User, int, error) {
	base := `FROM users`
	var conditions []string
	var args []any
	if username != "" {
		conditions = append(conditions, "username LIKE ?")
		args = append(args, "%"+username+"%")
	}
	if status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *status)
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(1) "+base+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, username, password, role, status, created_at, COALESCE(updated_at, created_at) ` +
		base + where + ` ORDER BY id DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreateTime, &u.UpdateTime); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET updated_at = NOW() WHERE id = ?`, id)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}
