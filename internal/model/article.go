package model

import (
	"time"

	"github.com/google/uuid"
)

type ArticleCreateRequest struct {
	TitleRU   string `json:"title_ru,omitempty"`
	TitleEN   string `json:"title_en,omitempty"`
	ContentRU string `json:"content_ru,omitempty"`
	ContentEN string `json:"content_en,omitempty"`
}

type Article struct {
	ID        uuid.UUID `json:"id"`
	TitleRU   string    `json:"title_ru,omitempty"`
	TitleEN   string    `json:"title_en,omitempty"`
	ContentRU string    `json:"content_ru,omitempty"`
	ContentEN string    `json:"content_en,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SearchResult struct {
	Article   Article             `json:"article"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}
