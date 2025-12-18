package config

import (
	"github.com/spf13/viper"
)

// I assumed here that we run the checkinservice in EKS and we set the
// DB connection variables as environment variables for the specific pod
// AWS config and SQS_QUEUE_URL will be handled the same

type Config struct {
	DBHost           string `mapstructure:"DB_HOST"`
	DBPort           string `mapstructure:"DB_PORT"`
	DBUser           string `mapstructure:"DB_USER"`
	DBPassword       string `mapstructure:"DB_PASSWORD"`
	DBName           string `mapstructure:"DB_NAME"`
	ServerPort       string `mapstructure:"SERVER_PORT"`
	AWSRegion        string `mapstructure:"AWS_REGION"`
	LaborSQSQueueURL string `mapstructure:"LABOR_SQS_QUEUE_URL"`
	EmailSQSQueueURL string `mapstructure:"EMAIL_SQS_QUEUE_URL"`
	AWSEndpoint      string `mapstructure:"AWS_ENDPOINT"`
	LegacyAPIURL     string `mapstructure:"LEGACY_API_URL"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (config Config, err error) {
	viper.SetDefault("DB_HOST", "db")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "user")
	viper.SetDefault("DB_PASSWORD", "password")
	viper.SetDefault("DB_NAME", "checkin_db")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("AWS_REGION", "us-east-1") // Default region for AWS services
	viper.SetDefault("LABOR_SQS_QUEUE_URL", "http://localstack:4566/000000000000/labor-queue")
	viper.SetDefault("EMAIL_SQS_QUEUE_URL", "http://localstack:4566/000000000000/email-queue")
	viper.SetDefault("AWS_ENDPOINT", "http://localstack:4566")
	viper.SetDefault("LEGACY_API_URL", "http://localhost:8081/")

	// Read in environment variables that match the keys.
	viper.AutomaticEnv()

	err = viper.Unmarshal(&config)
	return
}
