package repository

import (
	"context"

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

func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sso.users SET deleted = TRUE, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *UserRepository) SetBlocked(ctx context.Context, id uuid.UUID, block bool) error {
	query := `UPDATE sso.users SET blocked = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, block, id)
	return err
}
