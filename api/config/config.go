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
		Host string
		Port string
	}
	RabbitMQ struct {
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

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
