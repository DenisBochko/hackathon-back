package model

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken
// @Description Токен для восстановления пароля.
type PasswordResetToken struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"userID"`
	Token     []byte    `db:"token" json:"token"`
	ExpiresAt time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
} // @Name PasswordResetToken
