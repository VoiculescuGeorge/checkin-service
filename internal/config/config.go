package config

import (
	"github.com/spf13/viper"
)

// I assumed here that we run the checkinservice in EKS and we set the
// DB connection variables as environment variables for the specific pod

type Config struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	ServerPort string `mapstructure:"SERVER_PORT"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (config Config, err error) {
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "user")
	viper.SetDefault("DB_PASSWORD", "password")
	viper.SetDefault("DB_NAME", "checkin_db")
	viper.SetDefault("SERVER_PORT", "8080")

	// Read in environment variables that match the keys.
	viper.AutomaticEnv()

	err = viper.Unmarshal(&config)
	return
}
