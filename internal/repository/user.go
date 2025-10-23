package repository

import (
	"context"
	"database/sql"
	"errors"
	"hackathon-back/internal/apperrors"
	"hackathon-back/internal/model"
	"time"

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

func (r *UserRepository) UpdatePhotoURL(ctx context.Context, ext RepoExtension, userID uuid.UUID, photoURL string) error {
	if ext == nil {
		ext = r.db
	}

	const query = `
		UPDATE sso.users
		SET photo_url = $1,
		    updated_at = NOW()
		WHERE id = $2
		  AND deleted = false
		  AND blocked = false;
	`

	res, err := ext.Exec(ctx, query, photoURL, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return apperrors.ErrUserDoesNotExist
	}
	return nil
}

func (r *UserRepository) GetPhotoURL(ctx context.Context, ext RepoExtension, userID uuid.UUID) (string, error) {
	if ext == nil {
		ext = r.db
	}

	const query = `
		SELECT photo_url
		FROM sso.users
		WHERE id = $1
		  AND deleted = false
		  AND blocked = false;
	`

	var photo sql.NullString
	if err := ext.QueryRow(ctx, query, userID).Scan(&photo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.ErrUserDoesNotExist
		}
		return "", err
	}

	if !photo.Valid {
		return "", nil
	}
	return photo.String, nil
}

func (r *UserRepository) DeletePhoto(ctx context.Context, ext RepoExtension, userID uuid.UUID) (string, error) {
	if ext == nil {
		ext = r.db
	}

	const query = `
		UPDATE sso.users
		SET photo_url = NULL,
		    updated_at = NOW()
		WHERE id = $1
		RETURNING photo_url;
	`

	var oldPhoto sql.NullString
	if err := ext.QueryRow(ctx, query, userID).Scan(&oldPhoto); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.ErrUserDoesNotExist
		}
		return "", err
	}

	if !oldPhoto.Valid {
		return "", nil
	}
	return oldPhoto.String, nil
}

func (r *UserRepository) InsertPasswordResetToken(ctx context.Context, ext RepoExtension, userID uuid.UUID, token []byte, expiresAt time.Time) error {
	if ext == nil {
		ext = r.db
	}

	const query = `
		INSERT INTO sso.password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := ext.Exec(ctx, query, userID, token, expiresAt)
	return err
}

func (r *UserRepository) SelectUserByResetToken(ctx context.Context, ext RepoExtension, token []byte) (*model.User, error) {
	if ext == nil {
		ext = r.db
	}

	const query = `
		SELECT u.id, u.username, u.email, u.password, u.confirmed, u.deleted, u.blocked, u.role, u.created_at, u.updated_at
		FROM sso.password_reset_tokens t
		JOIN sso.users u ON t.user_id = u.id
		WHERE t.token = $1 AND t.expires_at > NOW();
	`

	var user model.User
	if err := ext.QueryRow(ctx, query, token).Scan(
		&user.ID, &user.Username, &user.Email, &user.HashedPassword,
		&user.Confirmed, &user.Deleted, &user.Blocked, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserDoesNotExist
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) DeletePasswordResetToken(ctx context.Context, ext RepoExtension, token []byte) error {
	if ext == nil {
		ext = r.db
	}

	const query = `DELETE FROM sso.password_reset_tokens WHERE token = $1`
	_, err := ext.Exec(ctx, query, token)
	return err
}

func (r *UserRepository) UpdateUserPassword(ctx context.Context, ext RepoExtension, userID uuid.UUID, hashedPassword []byte) error {
	if ext == nil {
		ext = r.db
	}

	const query = `
		UPDATE sso.users
		SET password = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := ext.Exec(ctx, query, hashedPassword, userID)
	return err
}
