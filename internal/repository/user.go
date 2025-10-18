package repository

import (
	"context"
	"hackathon-back/internal/apperrors"

	"github.com/google/uuid"
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
