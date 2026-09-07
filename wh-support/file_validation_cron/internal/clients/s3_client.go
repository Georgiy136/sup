package clients

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	serviceerrors "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/file_validation_cron/internal/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"
)

const (
	s3CfgKey = "s3_configuration"
)

type s3Config struct {
	BaseEndpoint    string `json:"base_endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	UsePathStyle    bool   `json:"use_path_style"`
}

type storageClient struct {
	s3Client *s3.Client
	config   *s3Config
}

func NewStorageClient() *storageClient {
	return &storageClient{}
}

func (s *storageClient) Configure(ctx context.Context, config configs.Config) {
	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired(s3CfgKey), &s.config); err != nil {
		logrus.Panicf("error unmarshaling s3 config: %v", err)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(s.config.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s.config.AccessKeyID,
			s.config.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		logrus.Panicf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s.config.BaseEndpoint)
		o.UsePathStyle = s.config.UsePathStyle
	})
	s.s3Client = client
}

func (s *storageClient) HeadObject(ctx context.Context, input models.HeadObjectRequest) (*models.HeadObjectResponse, error) {
	response, err := s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(input.Bucket),
		Key:    aws.String(input.Key),
	})
	if err != nil {
		if _, ok := errors.AsType[*types.NotFound](err); ok {
			return nil, serviceerrors.ErrObjectNotFound
		}

		return nil, fmt.Errorf("can't get info object %s from bucket %s: %w", input.Key, input.Bucket, err)
	}

	return &models.HeadObjectResponse{
		MimeType:  response.ContentType,
		SizeBytes: response.ContentLength,
	}, nil
}

func (s *storageClient) GetObject(ctx context.Context, input models.GetObjectRequest) (*models.GetObjectResponse, error) {
	response, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(input.Bucket),
		Key:    aws.String(input.Key),
		Range:  aws.String(input.Range),
	})
	if err != nil {
		return nil, fmt.Errorf("can't get object %s from bucket %s: %w", input.Key, input.Bucket, err)
	}

	return &models.GetObjectResponse{
		Body:         response.Body,
		ContentRange: response.ContentRange,
	}, nil
}

func (s *storageClient) DeleteObject(ctx context.Context, input models.DeleteObjectRequest) error {
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(input.Bucket),
		Key:    &input.Key,
	})
	if err != nil {
		return fmt.Errorf("can't delete object %s from bucket %s: %w", input.Key, input.Bucket, err)
	}

	return nil
}
