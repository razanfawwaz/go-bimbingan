package handlers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/razanfawwaz/bimbingan/util"
)

type S3Service struct {
	s3Client *s3.Client
	bucket   string
}

func NewR2Service() (*S3Service, error) {
	// Replace these values with your Cloudflare R2 Storage credentials
	account := util.GetConfig("ACCOUNT_ID")
	accessKey := util.GetConfig("ACCESS_KEY_ID")
	secretKey := util.GetConfig("SECRET_ACCESS_KEY")
	bucket := util.GetConfig("BUCKET_NAME")

	// Create custom resolver for R2 endpoint
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", account),
		}, nil
	})

	// Load AWS config with custom resolver
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("apac"),
	)
	if err != nil {
		return nil, err
	}

	// Create a new S3 client
	s3Client := s3.NewFromConfig(cfg)

	return &S3Service{
		s3Client: s3Client,
		bucket:   bucket,
	}, nil
}

func (s *S3Service) UploadFileToR2(ctx context.Context, key string, file []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(file),
		ContentType: aws.String(contentType),
	}

	_, err := s.s3Client.PutObject(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
