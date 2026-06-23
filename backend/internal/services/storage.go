package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageService struct {
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	publicBaseURL string
}

func NewStorageService() (*StorageService, error) {
	bucket := os.Getenv("STORAGE_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("STORAGE_BUCKET env is required")
	}
	accessKeyID := os.Getenv("STORAGE_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("STORAGE_SECRET_ACCESS_KEY")
	if accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("STORAGE_ACCESS_KEY_ID and STORAGE_SECRET_ACCESS_KEY env are required")
	}

	endpoint := os.Getenv("STORAGE_ENDPOINT")       // E.g., https://<account_id>.r2.cloudflarestorage.com
	publicBaseURL := os.Getenv("STORAGE_PUBLIC_URL") // E.g., https://pub-xxxx.r2.dev or custom domain
	region := os.Getenv("STORAGE_REGION")
	if region == "" {
		region = "auto"
	}

	creds := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(creds),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load storage config: %w", err)
	}

	var s3Client *s3.Client
	if endpoint != "" {
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			if os.Getenv("STORAGE_USE_PATH_STYLE") == "true" {
				o.UsePathStyle = true
			}
		})
	} else {
		s3Client = s3.NewFromConfig(cfg)
	}

	presignClient := s3.NewPresignClient(s3Client)

	return &StorageService{
		s3Client:      s3Client,
		presignClient: presignClient,
		bucket:        bucket,
		publicBaseURL: publicBaseURL,
	}, nil
}

func (s *StorageService) GeneratePresignedUploadURL(ctx context.Context, key, contentType string, expires time.Duration) (uploadURL, publicURL string, err error) {
	presignedReq, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	publicURL = fmt.Sprintf("%s/%s", s.publicBaseURL, key)
	return presignedReq.URL, publicURL, nil
}
