package queue

import (
	"encoding/json"

	"github.com/KPVISHNUSAI/product-management-system/api/repository/postgres"
	"github.com/KPVISHNUSAI/product-management-system/image-processor/processor"
	"github.com/streadway/amqp"
)

type Consumer struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	imageProcessor *processor.ImageProcessor
	productRepo    *postgres.ProductRepository
}

type ImageProcessingTask struct {
	ProductID uint     `json:"product_id"`
	Images    []string `json:"images"`
}

func NewConsumer(amqpURL string, imageProcessor *processor.ImageProcessor, productRepo *postgres.ProductRepository) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &Consumer{
		conn:           conn,
		channel:        ch,
		imageProcessor: imageProcessor,
		productRepo:    productRepo,
	}, nil
}

func (c *Consumer) Start() error {
	q, err := c.channel.QueueDeclare(
		"image_processing",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := c.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			var task ImageProcessingTask
			if err := json.Unmarshal(d.Body, &task); err != nil {
				continue
			}

			var compressedURLs []string
			for _, url := range task.Images {
				compressedURL, err := c.imageProcessor.ProcessImage(url)
				if err != nil {
					continue
				}
				compressedURLs = append(compressedURLs, compressedURL)
			}

			c.productRepo.UpdateCompressedImages(task.ProductID, compressedURLs)
		}
	}()

	return nil
}
