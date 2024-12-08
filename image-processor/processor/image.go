package processor

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
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
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	// Read content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("invalid content type: %s", contentType)
	}

	// Decode image
	img, format, err := image.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Compress image
	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 60})
	case "png":
		err = png.Encode(&buf, img)
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return "", fmt.Errorf("failed to compress image: %w", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d-%s.%s", time.Now().UnixNano(), uuid.New().String(), format)
	key := fmt.Sprintf("compressed/%s", filename)

	// Upload to S3
	_, err = p.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(fmt.Sprintf("image/%s", format)),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return fmt.Sprintf("s3://%s/%s", p.bucket, key), nil
}

func generateFileName(imageURL string) string {
	// Generate a unique filename using timestamp and random string
	return fmt.Sprintf("%d-%s.jpg", time.Now().UnixNano(), uuid.New().String())
}
