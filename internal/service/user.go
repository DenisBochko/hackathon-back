package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"hackathon-back/internal/model"
)

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
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
