package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/D-Kale/go-storage/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	appconfig "github.com/voxlab/voxlab-backend/internal/config"
)

type Storage struct {
	driver          *s3.S3Driver
	bucket          string
	publicEndpoint string
}

var Store *Storage

func tryEndpoint(cfg *appconfig.Config, endpoint string) (*s3.S3Driver, error) {
	s3Cfg := s3.S3Config{
		AccessKey:    cfg.MinIO.AccessKey,
		SecretKey:    cfg.MinIO.SecretKey,
		Bucket:       cfg.MinIO.Bucket,
		Endpoint:     fmt.Sprintf("http://%s", endpoint),
		Region:       "us-east-1",
		UsePathStyle: true,
	}

	driver, err := s3.ConnectS3(s3Cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", endpoint, err)
	}

	if err := driver.EnsureBucketExists(context.Background()); err != nil {
		return nil, fmt.Errorf("bucket on %s: %w", endpoint, err)
	}

	return driver, nil
}

func InitStorage(cfg *appconfig.Config) error {
	endpoints := []string{cfg.MinIO.Endpoint}

	// If the configured endpoint uses port 9000, add 9010 as fallback and vice versa
	switch {
	case strings.HasSuffix(cfg.MinIO.Endpoint, ":9000"):
		endpoints = append(endpoints, strings.Replace(cfg.MinIO.Endpoint, ":9000", ":9010", 1))
	case strings.HasSuffix(cfg.MinIO.Endpoint, ":9010"):
		endpoints = append(endpoints, strings.Replace(cfg.MinIO.Endpoint, ":9010", ":9000", 1))
	default:
		endpoints = append(endpoints, "localhost:9010")
	}

	var driver *s3.S3Driver
	var lastErr error

	for _, ep := range endpoints {
		log.Printf("MinIO trying endpoint: %s", ep)
		driver, lastErr = tryEndpoint(cfg, ep)
		if lastErr == nil {
			log.Printf("MinIO connected to: %s", ep)
			break
		}
		log.Printf("MinIO endpoint %s failed: %v", ep, lastErr)
	}

	if lastErr != nil {
		return fmt.Errorf("minio: all endpoints failed, last error: %w", lastErr)
	}

	Store = &Storage{
		driver:          driver,
		bucket:          cfg.MinIO.Bucket,
		publicEndpoint: cfg.MinIO.PublicEndpoint,
	}

	if err := makeBucketPublic(cfg); err != nil {
		log.Printf("Warning: failed to set public bucket policy: %v", err)
	}

	log.Println("MinIO (go-storage) initialized")
	return nil
}

func GetStorage() *Storage {
	return Store
}

func (s *Storage) Upload(ctx context.Context, path string, reader io.Reader) (string, error) {
	uploadedPath, err := s.driver.Upload(ctx, path, reader)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", s.publicEndpoint, s.bucket, uploadedPath), nil
}

func (s *Storage) Delete(ctx context.Context, path string) error {
	return s.driver.Delete(ctx, path)
}

func (s *Storage) Exists(ctx context.Context, path string) (bool, error) {
	return s.driver.Exists(ctx, path)
}

func makeBucketPublic(cfg *appconfig.Config) error {
	creds := credentials.NewStaticCredentialsProvider(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, "")
	awsCfg, err := awscfg.LoadDefaultConfig(context.TODO(),
		awscfg.WithRegion("us-east-1"),
		awscfg.WithCredentialsProvider(creds),
	)
	if err != nil {
		return fmt.Errorf("aws config: %w", err)
	}

	client := awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("http://%s", cfg.MinIO.Endpoint))
		o.UsePathStyle = true
	})

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetObject",
			"Resource": "arn:aws:s3:::%s/*"
		}]
	}`, cfg.MinIO.Bucket)

	_, err = client.PutBucketPolicy(context.TODO(), &awss3.PutBucketPolicyInput{
		Bucket: aws.String(cfg.MinIO.Bucket),
		Policy: aws.String(policy),
	})
	if err != nil {
		return fmt.Errorf("put bucket policy: %w", err)
	}

	log.Printf("Bucket %s is now publicly readable", cfg.MinIO.Bucket)
	return nil
}

func (s *Storage) GetURL(path string) string {
	return s.driver.GetURL(path)
}
