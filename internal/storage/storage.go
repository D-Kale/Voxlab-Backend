package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/D-Kale/go-storage/s3"
	"github.com/voxlab/voxlab-backend/internal/config"
)

type Storage struct {
	driver *s3.S3Driver
	bucket string
}

var Store *Storage

func InitStorage(cfg *config.Config) error {
	s3Cfg := s3.S3Config{
		AccessKey:    cfg.MinIO.AccessKey,
		SecretKey:    cfg.MinIO.SecretKey,
		Bucket:       cfg.MinIO.Bucket,
		Endpoint:     fmt.Sprintf("http://%s", cfg.MinIO.Endpoint),
		Region:       "us-east-1",
		UsePathStyle: true,
	}

	driver, err := s3.ConnectS3(s3Cfg)
	if err != nil {
		return fmt.Errorf("creating go-storage driver: %w", err)
	}

	Store = &Storage{
		driver: driver,
		bucket: cfg.MinIO.Bucket,
	}

	if err := driver.EnsureBucketExists(context.Background()); err != nil {
		return fmt.Errorf("ensuring bucket: %w", err)
	}

	log.Println("MinIO (go-storage) initialized")
	return nil
}

func GetStorage() *Storage {
	return Store
}

func (s *Storage) Upload(ctx context.Context, path string, reader io.Reader) (string, error) {
	return s.driver.Upload(ctx, path, reader)
}

func (s *Storage) Delete(ctx context.Context, path string) error {
	return s.driver.Delete(ctx, path)
}

func (s *Storage) Exists(ctx context.Context, path string) (bool, error) {
	return s.driver.Exists(ctx, path)
}

func (s *Storage) GetURL(path string) string {
	return s.driver.GetURL(path)
}
