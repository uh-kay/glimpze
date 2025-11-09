package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
	Client        *s3.Client
	PresignClient *s3.PresignClient
	BucketName    string
}

func NewR2Client(ctx context.Context, bucketName, accountID, accessKeyID, accessKeySecret string) (*R2Client, error) {
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKeyID,
				accessKeySecret,
				"",
			)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
	})

	presignClient := NewPresignClient(client)

	return &R2Client{
		Client:        client,
		PresignClient: presignClient,
		BucketName:    bucketName,
	}, nil
}

func NewPresignClient(client *s3.Client) *s3.PresignClient {
	return s3.NewPresignClient(client)
}

func (c *R2Client) SaveToR2(ctx context.Context, file multipart.File, fileExt, filename string) error {
	_, err := c.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.BucketName),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(fileExt),
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *R2Client) GetFromR2(ctx context.Context, filename string) (string, error) {
	presignResult, err := c.PresignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(filename),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}

func (c *R2Client) DeleteFromR2(ctx context.Context, filename string) error {
	_, err := c.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.BucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		return err
	}
	return nil
}
