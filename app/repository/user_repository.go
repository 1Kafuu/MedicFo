package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"medicfo/app/model"
)

// Sentinel error repository — service yang menerjemahkannya menjadi
// error HTTP.
var (
	ErrNotFound  = errors.New("resource not found")
	ErrDuplicate = errors.New("resource already exists")
)

type UserRepository interface {
	WithTx(tx DBTX) UserRepository
	Create(ctx context.Context, u model.User) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByID(ctx context.Context, id int64) (model.User, error)
}

type userPostgresRepository struct{ db DBTX }

func NewUserRepository(db DBTX) UserRepository {
	return &userPostgresRepository{db: db}
}

func (r *userPostgresRepository) WithTx(tx DBTX) UserRepository {
	return &userPostgresRepository{db: tx}
}

func (r *userPostgresRepository) Create(ctx context.Context, u model.User) (model.User, error) {
	var created model.User
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (username, email, password, role, is_active)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, username, email, password, role, is_active, created_at`,
		u.Username, u.Email, u.Password, u.Role, u.IsActive,
	).Scan(&created.ID, &created.Username, &created.Email, &created.Password,
		&created.Role, &created.IsActive, &created.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505 = unique_violation: username atau email sudah dipakai.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("menyimpan user: %w", err)
	}
	return created, nil
}

func (r *userPostgresRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mencari user by email: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int64) (model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, username, email, password, role, is_active, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.IsActive, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mencari user by id: %w", err)
	}
	return u, nil
}
