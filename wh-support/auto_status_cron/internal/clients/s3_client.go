package clients

import (
	"context"
	"fmt"
	"io"

	"gitlab.wildberries.ru/wbwh/wh-core/gocore_service_configs.git/configs"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type S3Client struct {
	s3Client *s3.Client
	config   *S3Config
}

type S3Config struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
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
}

func (c *S3Client) GetObject(ctx context.Context, objectKey string) ([]byte, error) {
	resp, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s/%s: %w", c.config.BucketName, objectKey, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Errorf("failed to close response body: %v", err)
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s/%s: %w", c.config.BucketName, objectKey, err)
	}

	return data, nil
}
