// image-processor/config/config.go
package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
	}
	Redis struct {
		Host     string
		Port     string
		Password string
	}
	RabbitMQ struct {
		URL string
	}
	AWS struct {
		Region    string
		Bucket    string
		AccessKey string
		SecretKey string
	}
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", "5431")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_DB", "product_management")
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("AWS_REGION", "ap-southeast-2")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	config.Database.Host = viper.GetString("POSTGRES_HOST")
	config.Database.Port = viper.GetString("POSTGRES_PORT")
	config.Database.User = viper.GetString("POSTGRES_USER")
	config.Database.Password = viper.GetString("POSTGRES_PASSWORD")
	config.Database.DBName = viper.GetString("POSTGRES_DB")
	config.RabbitMQ.URL = viper.GetString("RABBITMQ_URL")
	config.AWS.Region = viper.GetString("AWS_REGION")
	config.AWS.Bucket = viper.GetString("AWS_BUCKET_NAME")
	config.AWS.AccessKey = viper.GetString("AWS_ACCESS_KEY")
	config.AWS.SecretKey = viper.GetString("AWS_SECRET_KEY")
	config.Redis.Host = viper.GetString("REDIS_HOST")
	config.Redis.Port = viper.GetString("REDIS_PORT")
	config.Redis.Password = viper.GetString("REDIS_PASSWORD")

	return &config, nil
}
