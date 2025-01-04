package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Bot settings
	TelegramToken string `json:"token"`

	// Redis settings
	RedisAddr     string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`

	// New Relic settings
	NewRelicLicenseKey string `json:"new_relic_license_key"`
	NewRelicAppName    string `json:"new_relic_app_name"`

	// Other settings
	Environment string `json:"environment"`
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	config := &Config{
		TelegramToken:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		RedisAddr:          getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		NewRelicLicenseKey: os.Getenv("NEW_RELIC_LICENSE_KEY"),
		NewRelicAppName:    getEnvOrDefault("NEW_RELIC_APP_NAME", "Where is my bench bot"),
		Environment:        getEnvOrDefault("ENVIRONMENT", "production"),
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil

}

func (c *Config) Validate() error {
	var missingVars []string

	if c.TelegramToken == "" {
		missingVars = append(missingVars, "TELEGRAM_BOT_TOKEN")
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %v", strings.Join(missingVars, ", "))
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
