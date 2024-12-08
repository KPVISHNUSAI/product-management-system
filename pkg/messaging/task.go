package messaging

type ImageProcessingTask struct {
	ProductID uint     `json:"product_id"`
	Images    []string `json:"images"`
}
