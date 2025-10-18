package service

import (
	"context"

	"github.com/google/uuid"

	"hackathon-back/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *UserService) BlockUser(ctx context.Context, id uuid.UUID, block bool) error {
	return s.repo.SetBlocked(ctx, id, block)
}
