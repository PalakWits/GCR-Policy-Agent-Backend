package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseURL        string   `envconfig:"DATABASE_URL" required:"true"`
	RedisURL           string   `envconfig:"REDIS_URL" required:"true"`
	Port               string   `envconfig:"PORT" default:"8080"`
	LogLevel           string   `envconfig:"LOG_LEVEL" default:"info"`
	APIKeyHeader       string   `envconfig:"API_KEY_HEADER" default:"X-API-Key"`
	OtelURL            string   `envconfig:"OTEL_URL"`
	OtelService        string   `envconfig:"OTEL_SERVICE_NAME" default:"gcr-policy-agent-backend"`
	Domains            []string `envconfig:"DOMAINS" default:"ONDC:RET10,ONDC:RET11,ONDC:RET12,ONDC:RET13,ONDC:RET14,ONDC:RET15,ONDC:RET16,ONDC:RET17,ONDC:RET18"`
	RegistryURL        string   `envconfig:"REGISTRY_URL" default:"https://preprod.registry.ondc.org/v2.0/lookup"`
	PrivateKey         string   `envconfig:"PRIVATE_KEY" default:""`
	SubscriberID       string   `envconfig:"SUBSCRIBER_ID" default:""`
	UniqueKeyID        string   `envconfig:"UNIQUE_KEY_ID" default:""`
	MockSellerResponse bool     `envconfig:"MOCK_SELLER_RESPONSE" default:"false"`
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Warning: error loading .env file: %v\n", err)
	}

	config := &Config{}

	err = envconfig.Process("", config)
	if err != nil {
		return nil, fmt.Errorf("error processing envconfig: %w", err)
	}

	return config, nil
}
