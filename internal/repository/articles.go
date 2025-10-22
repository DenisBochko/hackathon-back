package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hackathon-back/internal/apperrors"
	"hackathon-back/internal/model"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
)

const indexName = "articles"

type ElasticRepo struct {
	es *elasticsearch.Client
}

func NewElasticRepository(es *elasticsearch.Client) *ElasticRepo {
	return &ElasticRepo{
		es: es,
	}
}

func (r *ElasticRepo) EnsureIndex(ctx context.Context) (err error) {
	exists, err := r.es.Indices.Exists([]string{indexName}, r.es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("check index existence: %w", err)
	}

	defer func() {
		if cErr := exists.Body.Close(); cErr != nil {
			err = fmt.Errorf("%w, failed to close response body: %w", err, cErr)
		}
	}()

	if exists.StatusCode == http.StatusOK {
		return nil
	}

	mapping := `{
		"settings": {
			"analysis": {
				"analyzer": {
					"ru_analyzer": {"type": "standard", "stopwords": "_russian_"},
					"en_analyzer": {"type": "standard", "stopwords": "_english_"}
				}
			}
		},
		"mappings": {
			"properties": {
				"title_ru": {"type": "text", "analyzer": "ru_analyzer"},
				"title_en": {"type": "text", "analyzer": "en_analyzer"},
				"content_ru": {"type": "text", "analyzer": "ru_analyzer"},
				"content_en": {"type": "text", "analyzer": "en_analyzer"},
				"created_at": {"type": "date"},
				"updated_at": {"type": "date"}
			}
		}
	}`

	res, err := r.es.Indices.Create(indexName, r.es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	defer func() {
		if cErr := exists.Body.Close(); cErr != nil {
			err = fmt.Errorf("%w, failed to close response body: %w", err, cErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("index creation failed: %s", res.String())
	}

	return nil
}

func (r *ElasticRepo) Create(ctx context.Context, article *model.Article) (err error) {
	data, _ := json.Marshal(article)
	res, err := r.es.Index(indexName, bytes.NewReader(data), r.es.Index.WithDocumentID(article.ID.String()), r.es.Index.WithContext(ctx))
	if err != nil {
		return err
	}

	defer func() {
		if cErr := res.Body.Close(); cErr != nil {
			err = fmt.Errorf("%w, failed to close response body: %w", err, cErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("failed to create article: %s", res.String())
	}

	return nil
}

func (r *ElasticRepo) Get(ctx context.Context, id string) (*model.Article, error) {
	res, err := r.es.Get(indexName, id, r.es.Get.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	defer func() {
		if cErr := res.Body.Close(); cErr != nil {
			err = fmt.Errorf("%w, failed to close response body: %w", err, cErr)
		}
	}()

	if res.IsError() {
		return nil, apperrors.ErrArticleDoesNotExist
	}

	var doc struct {
		Source model.Article `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &doc.Source, nil
}

func (r *ElasticRepo) Delete(ctx context.Context, id string) error {
	res, err := r.es.Delete(indexName, id, r.es.Delete.WithContext(ctx))
	if err != nil {
		return err
	}

	defer func() {
		if cErr := res.Body.Close(); cErr != nil {
			err = fmt.Errorf("%w, failed to close response body: %w", err, cErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("failed to delete article: %s", res.String())
	}

	return nil
}

func (r *ElasticRepo) Patch(ctx context.Context, id string, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()

	doc := map[string]interface{}{"doc": fields}
	data, _ := json.Marshal(doc)

	res, err := r.es.Update(indexName, id, bytes.NewReader(data), r.es.Update.WithContext(ctx))
	if err != nil {
		return err
	}

	defer func() {
		if cErr := res.Body.Close(); cErr != nil {
			err = fmt.Errorf("%w, failed to close response body: %w", err, cErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("failed to patch article: %s", res.String())
	}

	return nil
}

func (r *ElasticRepo) Search(ctx context.Context, query string) ([]model.SearchResult, error) {
	body := fmt.Sprintf(`{
		"query": {
			"multi_match": {
				"query": "%s",
				"fields": ["title_ru", "title_en", "content_ru", "content_en"]
			}
		},
		"highlight": {
			"pre_tags": ["<em>"],
			"post_tags": ["</em>"],
			"fields": {
				"title_ru": {}, "title_en": {}, "content_ru": {}, "content_en": {}
			}
		}
	}`, query)

	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex(indexName),
		r.es.Search.WithBody(strings.NewReader(body)),
		r.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cErr := res.Body.Close(); cErr != nil {
			err = fmt.Errorf("%w, failed to close response body: %w", err, cErr)
		}
	}()

	var result struct {
		Hits struct {
			Hits []struct {
				Source    model.Article       `json:"_source"`
				Highlight map[string][]string `json:"highlight,omitempty"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	out := make([]model.SearchResult, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		out = append(out, model.SearchResult{
			Article:   hit.Source,
			Highlight: hit.Highlight,
		})
	}

	return out, nil
}
