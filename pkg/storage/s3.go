package storage

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
)

type S3Client struct {
	client *s3.S3
	bucket string
}

func NewS3Client(s3Client *s3.S3, bucket string) *S3Client {
	return &S3Client{
		client: s3Client,
		bucket: bucket,
	}
}

func (c *S3Client) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err := c.client.PutObject(input)
	if err != nil {
		return "", err
	}

	return "https://" + c.bucket + ".s3.amazonaws.com/" + key, nil
}

func (c *S3Client) DownloadFile(ctx context.Context, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	result, err := c.client.GetObject(input)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func (c *S3Client) DeleteFile(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	_, err := c.client.DeleteObject(input)
	return err
}
