package repository

import (
	"context"
	"errors"
	"hackathon-back/internal/apperrors"
	"hackathon-back/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "hackathon-back/internal/model"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Pool() *pgxpool.Pool {
	return r.db
}

func (r *UserRepository) InsertUser(ctx context.Context, ext RepoExtension, user *model.User) (*model.User, error) {
	if ext == nil {
		ext = r.db
	}

	const query = `
		INSERT INTO sso.users (id, username, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING confirmed, deleted, blocked, role, created_at, updated_at;
	`

	err := ext.QueryRow(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.HashedPassword,
	).Scan(
		&user.Confirmed,
		&user.Deleted,
		&user.Blocked,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperrors.ErrUserAlreadyExists
		}

		return nil, err
	}

	return user, nil
}

func (r *UserRepository) SelectUserByID(ctx context.Context, ext RepoExtension, id uuid.UUID) (*model.User, error) {
	if ext == nil {
		ext = r.db
	}

	const query = `
		SELECT id, username, email, password, confirmed, deleted, blocked, role, created_at, updated_at
		FROM sso.users
		WHERE id = $1
		  AND deleted = false 
		  AND blocked = false;
	`

	var user model.User

	if err := ext.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.HashedPassword,
		&user.Confirmed,
		&user.Deleted,
		&user.Blocked,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserDoesNotExist
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) SelectUserByEmail(ctx context.Context, ext RepoExtension, email string) (*model.User, error) {
	if ext == nil {
		ext = r.db
	}

	const query = `
		SELECT id, username, email, password, confirmed, deleted, blocked, role, created_at, updated_at
		FROM sso.users
		WHERE email = $1
		  AND deleted = false 
		  AND blocked = false;
	`

	var user model.User

	if err := ext.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.HashedPassword,
		&user.Confirmed,
		&user.Deleted,
		&user.Blocked,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserDoesNotExist
		}

		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Delete(ctx context.Context, ext RepoExtension, id uuid.UUID) error {
	if ext == nil {
		ext = r.db
	}

	const query = `
		UPDATE sso.users
		SET deleted = TRUE, 
			updated_at = NOW() 
		WHERE id = $1
	`

	res, err := ext.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return apperrors.ErrUserDoesNotExist
	}

	return nil
}

func (r *UserRepository) Block(ctx context.Context, ext RepoExtension, id uuid.UUID) error {
	if ext == nil {
		ext = r.db
	}

	const query = `
		UPDATE sso.users 
		SET blocked = true, 
		    updated_at = NOW() 
		WHERE id = $1
	`

	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return apperrors.ErrUserDoesNotExist
	}

	return nil
}
