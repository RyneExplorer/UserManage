package mysql

import (
	"UserManagement/internal/model"
	"UserManagement/internal/repository"
	"context"
	"database/sql"
	"errors"
)

type UserRepositoryMySQL struct {
	db *sql.DB
}

func NewUserRepositoryMySQL(db *sql.DB) repository.UserRepository {
	return &UserRepositoryMySQL{db: db}
}

func (r *UserRepositoryMySQL) Create(ctx context.Context, user model.User) (int, error) {
	res, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (username, password_hash, role, status, create_time, last_time)
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

func (r *UserRepositoryMySQL) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM users`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepositoryMySQL) FindByID(ctx context.Context, id int) (*model.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, role, status, create_time, COALESCE(last_time, create_time)
		 FROM users WHERE id = ?`,
		id,
	)

	var u model.User
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreateTime, &u.LastTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepositoryMySQL) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, role, status, create_time, COALESCE(last_time, create_time)
		 FROM users WHERE username = ?`,
		username,
	)

	var u model.User
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreateTime, &u.LastTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepositoryMySQL) ListAll(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, username, password_hash, role, status, create_time, COALESCE(last_time, create_time)
		 FROM users ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.Role, &u.Status, &u.CreateTime, &u.LastTime); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *UserRepositoryMySQL) Update(ctx context.Context, user model.User) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE users
		 SET username = ?, role = ?, status = ?
		 WHERE id = ?`,
		user.Username,
		user.Role,
		user.Status,
		user.ID,
	)
	return err
}

func (r *UserRepositoryMySQL) UpdateLastLogin(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET last_time = NOW() WHERE id = ?`, id)
	return err
}

func (r *UserRepositoryMySQL) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}
