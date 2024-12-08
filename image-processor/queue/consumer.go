// image-processor/queue/consumer.go
package queue

import (
	"encoding/json"
	"log"
	"time"

	"github.com/KPVISHNUSAI/product-management-system/api/repository/postgres"
	"github.com/KPVISHNUSAI/product-management-system/api/services"
	"github.com/KPVISHNUSAI/product-management-system/image-processor/processor"
	"github.com/lib/pq"
	"github.com/streadway/amqp"
)

type Consumer struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	imageProcessor *processor.ImageProcessor
	productRepo    *postgres.ProductRepository
	productService *services.ProductService
	queueName      string
	dlqName        string
}

type ImageProcessingTask struct {
	ProductID uint     `json:"product_id"`
	Images    []string `json:"images"`
}

func NewConsumer(amqpURL string, imageProcessor *processor.ImageProcessor, productRepo *postgres.ProductRepository, productService *services.ProductService) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Consumer{
		conn:           conn,
		channel:        ch,
		imageProcessor: imageProcessor,
		productRepo:    productRepo,
		productService: productService,
		queueName:      "image_processing",
		dlqName:        "image_processing_dlq",
	}, nil
}

func (c *Consumer) Start() error {
	// Declare main queue
	q, err := c.channel.QueueDeclare(
		c.queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// Declare DLQ
	_, err = c.channel.QueueDeclare(
		c.dlqName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return err
	}

	// Set QoS
	err = c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}

	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			var task ImageProcessingTask
			if err := json.Unmarshal(d.Body, &task); err != nil {
				c.handleProcessingError(task, err)
				d.Nack(false, true)
				continue
			}

			err = c.productRepo.UpdateProcessingStatus(task.ProductID, "processing")
			if err != nil {
				c.handleProcessingError(task, err)
				d.Nack(false, true)
				continue
			}

			var compressedURLs pq.StringArray
			var processingError error

			for _, url := range task.Images {
				for retries := 0; retries < 3; retries++ {
					compressedURL, err := c.imageProcessor.ProcessImage(url)
					if err == nil {
						compressedURLs = append(compressedURLs, compressedURL)
						break
					}
					if retries == 2 {
						processingError = err
					}
					time.Sleep(time.Second * time.Duration(retries+1))
				}
				if processingError != nil {
					break
				}
			}

			if processingError != nil {
				c.handleProcessingError(task, processingError)
				d.Nack(false, false)
				continue
			}

			// Ensure you're passing pq.Array here
			err = c.productRepo.UpdateCompressedImages(task.ProductID, compressedURLs)

			if err != nil {
				c.handleProcessingError(task, err)
				d.Nack(false, true)
				continue
			}

			// Invalidate cache after updating images
			if err := c.productService.InvalidateCache(task.ProductID); err != nil {
				log.Printf("Failed to invalidate cache: %v", err)
			}

			err = c.productRepo.UpdateProcessingStatus(task.ProductID, "completed")
			if err != nil {
				log.Printf("Failed to update status to completed: %v", err)
			}

			d.Ack(false)
		}
	}()

	return nil
}

func (c *Consumer) handleProcessingError(task ImageProcessingTask, err error) {
	log.Printf("Error processing task for product %d: %v", task.ProductID, err)

	// Update product status to failed
	if err := c.productRepo.UpdateProcessingStatus(task.ProductID, "failed"); err != nil {
		log.Printf("Failed to update status to failed: %v", err)
	}

	// Send to dead letter queue
	errMsg, _ := json.Marshal(map[string]interface{}{
		"product_id": task.ProductID,
		"error":      err.Error(),
		"timestamp":  time.Now(),
	})

	err = c.channel.Publish(
		"",        // exchange
		c.dlqName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        errMsg,
		},
	)
	if err != nil {
		log.Printf("Failed to publish to DLQ: %v", err)
	}
}

func (c *Consumer) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}
