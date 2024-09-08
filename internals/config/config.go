package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	SMTPPort     int    `yaml:"smtp_port"`
	SMTPHost     string `yaml:"smtp_host"`
	APIPort      int    `yaml:"api_port"`
	DatabaseURL  string `yaml:"database_url"`
	JWTSecret    string
	RateLimit    int    `yaml:"rate_limit"`
	MaxFileSize  int    `yaml:"max_file_size"`
	SMTPUsername string `yaml:"smtp_username"`
	SMTPPassword string
}

func Load(filename string) (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Load sensitive data from environment variables
	config.JWTSecret = os.Getenv("JWT_SECRET")
	config.SMTPPassword = os.Getenv("SMTP_PASSWORD")

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) validate() error {
	if c.SMTPPort == 0 {
		return fmt.Errorf("SMTP port is required")
	}
	if c.SMTPHost == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if c.APIPort == 0 {
		return fmt.Errorf("API port is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("Database URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if c.RateLimit == 0 {
		return fmt.Errorf("Rate limit is required")
	}
	if c.MaxFileSize == 0 {
		return fmt.Errorf("Max file size is required")
	}
	if c.SMTPUsername == "" {
		return fmt.Errorf("SMTP username is required")
	}
	if c.SMTPPassword == "" {
		return fmt.Errorf("SMTP password is required")
	}
	return nil
}