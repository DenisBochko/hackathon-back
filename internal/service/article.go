package service

import (
	"context"
	"fmt"
	"hackathon-back/internal/model"
	"time"

	"github.com/google/uuid"
)

type ArticleRepository interface {
	EnsureIndex(ctx context.Context) error
	Create(ctx context.Context, article *model.Article) error
	Get(ctx context.Context, id string) (*model.Article, error)
	Delete(ctx context.Context, id string) error
	Patch(ctx context.Context, id string, fields map[string]interface{}) error
	Search(ctx context.Context, query string) ([]model.SearchResult, error)
}

type ArticleService struct {
	articleRepo ArticleRepository
}

func NewArticleService(articleRepo ArticleRepository) *ArticleService {
	return &ArticleService{
		articleRepo: articleRepo,
	}
}

func (s *ArticleService) CreateArticle(ctx context.Context, articleRequest *model.ArticleCreateRequest) (*model.Article, error) {
	article := &model.Article{
		ID:        uuid.New(),
		TitleRU:   articleRequest.TitleRU,
		TitleEN:   articleRequest.TitleEN,
		ContentRU: articleRequest.ContentRU,
		ContentEN: articleRequest.ContentEN,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.articleRepo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}

	return article, nil
}

func (s *ArticleService) GetArticle(ctx context.Context, id string) (*model.Article, error) {
	article, err := s.articleRepo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	return article, nil
}

func (s *ArticleService) DeleteArticle(ctx context.Context, id string) error {
	if err := s.articleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	return nil
}

func (s *ArticleService) UpdateArticle(ctx context.Context, id string, fields map[string]interface{}) error {
	if err := s.articleRepo.Patch(ctx, id, fields); err != nil {
		return fmt.Errorf("failed to update article: %w", err)
	}

	return nil
}

func (s *ArticleService) SearchArticles(ctx context.Context, query string) ([]model.SearchResult, error) {
	articles, err := s.articleRepo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search articles: %w", err)
	}

	return articles, nil
}
