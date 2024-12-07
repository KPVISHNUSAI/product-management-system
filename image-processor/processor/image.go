package processor

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"image"
	"image/jpeg"
	"net/http"
	"time"
)

type ImageProcessor struct {
	s3Client *s3.S3
	bucket   string
}

func NewImageProcessor(s3Client *s3.S3, bucket string) *ImageProcessor {
	return &ImageProcessor{
		s3Client: s3Client,
		bucket:   bucket,
	}
}

func (p *ImageProcessor) ProcessImage(imageURL string) (string, error) {
	// Download image
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Decode image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", err
	}

	// Compress image
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 60}); err != nil {
		return "", err
	}

	// Upload to S3
	key := "compressed/" + generateFileName(imageURL)
	_, err = p.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})

	return "s3://" + p.bucket + "/" + key, err
}

func generateFileName(imageURL string) string {
	// Generate a unique filename using timestamp and random string
	return fmt.Sprintf("%d-%s.jpg", time.Now().UnixNano(), uuid.New().String())
}
