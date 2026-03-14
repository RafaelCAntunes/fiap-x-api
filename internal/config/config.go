package config

import (
	"os"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	SQSURL     string
	S3Bucket   string
	EnvVariable string
}

func LoadConfig() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "user_fiap"),
		DBPassword: getEnv("DB_PASSWORD", "password_fiap"),
		DBName:     getEnv("DB_NAME", "fiap_x_video_db"),
		SQSURL:     getEnv("SQS_QUEUE_URL", ""),
		S3Bucket:   getEnv("S3_BUCKET_NAME", "fiap-x-videos-content"),
		EnvVariable: 	getEnv("APP_ENV", "local"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}