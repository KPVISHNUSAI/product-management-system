package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port      string
		JWTSecret string
	}
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
		URL      string
		Host     string
		Port     string
		User     string
		Password string
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

	viper.SetDefault("SERVER_PORT", "9000")
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "redis")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Load Server config
	config.Server.Port = viper.GetString("SERVER_PORT")
	config.Server.JWTSecret = viper.GetString("JWT_SECRET")

	// Load Database config
	config.Database.Host = viper.GetString("POSTGRES_HOST")
	config.Database.Port = viper.GetString("POSTGRES_PORT")
	config.Database.User = viper.GetString("POSTGRES_USER")
	config.Database.Password = viper.GetString("POSTGRES_PASSWORD")
	config.Database.DBName = viper.GetString("POSTGRES_DB")

	// Load Redis config
	config.Redis.Host = viper.GetString("REDIS_HOST")
	config.Redis.Port = viper.GetString("REDIS_PORT")
	config.Redis.Password = viper.GetString("REDIS_PASSWORD")

	// Load RabbitMQ config
	config.RabbitMQ.URL = viper.GetString("RABBITMQ_URL")
	config.RabbitMQ.Host = viper.GetString("RABBITMQ_HOST")
	config.RabbitMQ.Port = viper.GetString("RABBITMQ_PORT")
	config.RabbitMQ.User = viper.GetString("RABBITMQ_USER")
	config.RabbitMQ.Password = viper.GetString("RABBITMQ_PASSWORD")

	return &config, nil
}
