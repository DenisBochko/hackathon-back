package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"hackathon-back/internal/model"
	"hackathon-back/pkg/mailer"
	"hackathon-back/pkg/storage"
	"mime/multipart"
	"time"
)

type UserService struct {
	userRepo UserRepository
	minio    storage.MinIO
	mailer   mailer.Mailer
}

func NewUserService(userRepo UserRepository, minio storage.MinIO, mlr mailer.Mailer) *UserService {
	return &UserService{
		userRepo: userRepo,
		minio:    minio,
		mailer:   mlr,
	}
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.SelectUserByID(ctx, nil, id)
	if err != nil {
		return nil, fmt.Errorf("failed to select user: %w", err)
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, nil, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *UserService) BlockUser(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepo.Block(ctx, nil, id); err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	return nil
}

func (s *UserService) UploadUserPhoto(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()

	// Проверяем, есть ли старое фото
	oldPhotoURL, _ := s.userRepo.GetPhotoURL(ctx, nil, userID)
	if oldPhotoURL != "" {
		_ = s.minio.DeleteUserPhoto(ctx, oldPhotoURL)
	}

	// Загружаем новое фото в MinIO
	url, err := s.minio.UploadUserPhoto(ctx, userID, file, header)
	if err != nil {
		return "", fmt.Errorf("failed to upload photo: %w", err)
	}

	// Обновляем URL в БД
	if err := s.userRepo.UpdatePhotoURL(ctx, nil, userID, url); err != nil {
		return "", fmt.Errorf("failed to save photo URL: %w", err)
	}

	return url, nil
}

// GetUserPhoto — получить ссылку на фото из БД.
func (s *UserService) GetUserPhoto(ctx context.Context, userID uuid.UUID) (string, error) {
	url, err := s.userRepo.GetPhotoURL(ctx, nil, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user photo: %w", err)
	}
	if url == "" {
		return "", fmt.Errorf("user has no photo")
	}
	return url, nil
}

// DeleteUserPhoto — удалить фото из MinIO и очистить ссылку в БД.
func (s *UserService) DeleteUserPhoto(ctx context.Context, userID uuid.UUID) error {
	oldPhotoURL, err := s.userRepo.GetPhotoURL(ctx, nil, userID)
	if err != nil {
		return fmt.Errorf("failed to get user photo: %w", err)
	}

	// Удаляем фото в MinIO, если оно есть
	if oldPhotoURL != "" {
		_ = s.minio.DeleteUserPhoto(ctx, oldPhotoURL)
	}

	// Очищаем ссылку в БД
	if err := s.userRepo.UpdatePhotoURL(ctx, nil, userID, ""); err != nil {
		return fmt.Errorf("failed to clear photo URL: %w", err)
	}

	return nil
}

func (s *UserService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.SelectUserByEmail(ctx, nil, email)
	if err != nil {
		return err
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}

	expiresAt := time.Now().Add(15 * time.Minute)

	if err := s.userRepo.InsertPasswordResetToken(ctx, nil, user.ID, tokenBytes, expiresAt); err != nil {
		return err
	}

	tokenStr := base64.URLEncoding.EncodeToString(tokenBytes)
	resetURL := fmt.Sprintf("https://frontend.example.com/reset-password?token=%s", tokenStr)

	if err := s.mailer.SendHTML(user.Email, "Password Reset", "Click here to reset your password", resetURL); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	tokenBytes, err := base64.URLEncoding.DecodeString(tokenStr)
	if err != nil {
		return err
	}

	user, err := s.userRepo.SelectUserByResetToken(ctx, nil, tokenBytes)
	if err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdateUserPassword(ctx, nil, user.ID, hashed); err != nil {
		return err
	}

	return s.userRepo.DeletePasswordResetToken(ctx, nil, tokenBytes)
}

func (s *UserService) DeleteSelf(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.Delete(ctx, nil, userID)
}
