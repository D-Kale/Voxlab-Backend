package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/voxlab/voxlab-backend/internal/config"
)

type MinioStorage struct {
	Client *minio.Client
	Bucket string
}

var Minio *MinioStorage

func InitMinio(cfg *config.Config) error {
	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("error creando cliente MinIO: %w", err)
	}

	Minio = &MinioStorage{
		Client: client,
		Bucket: cfg.MinIO.Bucket,
	}

	if err := Minio.EnsureBucket(); err != nil {
		return fmt.Errorf("error creando bucket: %w", err)
	}

	log.Println("✅ MinIO inicializado correctamente")
	return nil
}

func GetMinio() *MinioStorage {
	return Minio
}

func (m *MinioStorage) EnsureBucket() error {
	ctx := context.Background()
	exists, err := m.Client.BucketExists(ctx, m.Bucket)
	if err != nil {
		return fmt.Errorf("error verificando bucket: %w", err)
	}

	if !exists {
		err = m.Client.MakeBucket(ctx, m.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("error creando bucket: %w", err)
		}

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, m.Bucket)

		err = m.Client.SetBucketPolicy(ctx, m.Bucket, policy)
		if err != nil {
			return fmt.Errorf("error configurando política del bucket: %w", err)
		}
	}

	return nil
}

func (m *MinioStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	_, err := m.Client.PutObject(ctx, m.Bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("error subiendo archivo: %w", err)
	}

	return fmt.Sprintf("http://%s/%s/%s", m.Client.EndpointURL().Host, m.Bucket, objectName), nil
}

func (m *MinioStorage) GetFileURL(objectName string) string {
	return fmt.Sprintf("http://%s/%s/%s", m.Client.EndpointURL().Host, m.Bucket, objectName)
}

func (m *MinioStorage) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := m.Client.PresignedGetObject(ctx, m.Bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("error generando URL firmada: %w", err)
	}

	return url.String(), nil
}

func (m *MinioStorage) DeleteFile(ctx context.Context, objectName string) error {
	err := m.Client.RemoveObject(ctx, m.Bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("error eliminando archivo: %w", err)
	}

	return nil
}

func (m *MinioStorage) FileExists(ctx context.Context, objectName string) (bool, error) {
	_, err := m.Client.StatObject(ctx, m.Bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
