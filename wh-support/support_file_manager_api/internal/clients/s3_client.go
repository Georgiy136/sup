package clients

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/local_errors"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/models"
	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/support_file_manager_api/internal/utils"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type S3Client struct {
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	config        *S3Config
}

type S3Config struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	TmpBucketName   string `json:"tmp_bucket_name"`
	BucketName      string `json:"bucket_name"`
	UsePathStyle    bool   `json:"use_path_style"`
}

func NewS3Client() *S3Client {
	return &S3Client{}
}

func (c *S3Client) Configure(ctx context.Context, config configs.Config) {
	const s3CfgKey = "s3_configuration"

	if err := jsoniter.Unmarshal(config.GetByServiceKeyRequired(s3CfgKey), &c.config); err != nil {
		logrus.Panicf("error unmarshaling s3 config: %v", err)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(c.config.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			c.config.AccessKeyID,
			c.config.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		logrus.Panicf("unable to load SDK config, %v", err)
	}

	c.s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.config.Endpoint)
		o.UsePathStyle = c.config.UsePathStyle
	})

	c.presignClient = s3.NewPresignClient(c.s3Client)
}

func (c *S3Client) GetPresignedUploadURL(ctx context.Context, input models.GetUploadURLRequest) (*models.GetUploadURLResponse, error) {
	presignedReq, err := c.presignClient.PresignPostObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.config.TmpBucketName),
		Key:    aws.String(input.ObjectKey),
	}, func(options *s3.PresignPostOptions) {
		options.Expires = input.Expiration
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned post upload URL: %w", err)
	}

	presignedReq = c.prepareS3RWBPresignPostObject(presignedReq, c.GetTmpBucketName())
	return &models.GetUploadURLResponse{
		UploadURL:          presignedReq.URL,
		Method:             http.MethodPost,
		RequiredFormFields: presignedReq.Values,
	}, nil
}

func (c *S3Client) prepareS3RWBPresignPostObject(presignedReq *s3.PresignedPostRequest, bucketName string) *s3.PresignedPostRequest {
	presignedReq.Values["bucket"] = bucketName
	return presignedReq
}

func (c *S3Client) GetPresignedDownloadURL(ctx context.Context, input models.GetDownloadURLRequest) (string, error) {
	presignedReq, err := c.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(input.Bucket),
		Key:                        aws.String(input.ObjectKey),
		ResponseContentDisposition: aws.String(fmt.Sprintf("%s; filename*=UTF-8''%s", utils.GetContentDisposition(input.ContentType), input.OriginalFileName)),
	}, s3.WithPresignExpires(input.Expiration))
	if err != nil {
		if isNotFoundError(err) {
			return "", local_errors.ErrFileNotFound
		}
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return presignedReq.URL, nil
}

func (c *S3Client) GetTmpBucketName() string {
	return c.config.TmpBucketName
}

func (c *S3Client) GetBucketName() string {
	return c.config.BucketName
}

func (c *S3Client) CopyObject(ctx context.Context, input models.CopyObjectRequest) error {
	copySource := fmt.Sprintf("%s/%s", input.SourceBucket, input.SourceKey)

	_, err := c.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(input.DestBucket),
		Key:        aws.String(input.DestKey),
		CopySource: aws.String(copySource),
	})
	if err != nil {
		if isNotFoundError(err) {
			return local_errors.ErrFileNotFound
		}
		return fmt.Errorf("failed to copy object from %s to %s/%s: %w", copySource, input.DestBucket, input.DestKey, err)
	}

	return nil
}

func (c *S3Client) DeleteObject(ctx context.Context, bucket, objectKey string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		if isNotFoundError(err) {
			return local_errors.ErrFileNotFound
		}
		return fmt.Errorf("failed to delete object %s/%s: %w", bucket, objectKey, err)
	}

	return nil
}

func (c *S3Client) PutObject(ctx context.Context, input models.PutObjectRequest) error {
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(input.Bucket),
		Key:           aws.String(input.ObjectKey),
		Body:          bytes.NewReader(input.FileData),
		ContentType:   aws.String(input.ContentType),
		ContentLength: aws.Int64(int64(len(input.FileData))),
	})
	if err != nil {
		return fmt.Errorf("failed to put object %s/%s: %w", input.Bucket, input.ObjectKey, err)
	}

	return nil
}

func isNotFoundError(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		return code == "NotFound" || code == "NoSuchKey"
	}
	return false
}
