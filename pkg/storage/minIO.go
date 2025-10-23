package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	minioSDK "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIO интерфейс для взаимодействия с MinIO
type MinIO interface {
	Client() *minioSDK.Client
	UploadUserPhoto(ctx context.Context, id uuid.UUID, file multipart.File, header *multipart.FileHeader) (string, error)
	DeleteUserPhoto(ctx context.Context, photoURL string) error
	GetUserPhotoURL(ctx context.Context, objectName string, expires time.Duration) (string, error)
	Close() error
}

// Config конфигурация подключения к MinIO
type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	BucketName string
	BaseURL    string
}

// minioClient — реализация интерфейса MinIO
type minioClient struct {
	client     *minioSDK.Client
	bucketName string
	baseURL    string
}

// New создаёт новый экземпляр MinIO клиента
func New(ctx context.Context, cfg *Config) (MinIO, error) {
	client, err := minioSDK.New(cfg.Endpoint, &minioSDK.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Проверяем, существует ли bucket
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		if err = client.MakeBucket(ctx, cfg.BucketName, minioSDK.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &minioClient{
		client:     client,
		bucketName: cfg.BucketName,
		baseURL:    strings.TrimSuffix(cfg.BaseURL, "/"),
	}, nil
}

// Client возвращает minio клиент
func (m *minioClient) Client() *minioSDK.Client {
	return m.client
}

// UploadUserPhoto загружает фото пользователя и возвращает публичную ссылку
func (m *minioClient) UploadUserPhoto(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	objectName := fmt.Sprintf("%s_%d%s", userID.String(), time.Now().Unix(), ext)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	_, err := m.client.PutObject(ctx, m.bucketName, objectName, bytes.NewReader(buf.Bytes()), int64(buf.Len()), minioSDK.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload photo: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", m.baseURL, m.bucketName, objectName), nil
}

// DeleteUserPhoto удаляет фото по URL
func (m *minioClient) DeleteUserPhoto(ctx context.Context, photoURL string) error {
	if photoURL == "" {
		return nil
	}

	objectName := strings.TrimPrefix(photoURL, fmt.Sprintf("%s/%s/", m.baseURL, m.bucketName))
	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minioSDK.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}
	return nil
}

// GetUserPhotoURL возвращает presigned URL (временная ссылка на скачивание)
func (m *minioClient) GetUserPhotoURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

// Close — для совместимости, ничего не делает (MinIO клиент не требует закрытия)
func (m *minioClient) Close() error {
	return nil
}
