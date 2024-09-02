package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	SMTPPort     int    `yaml:"smtp_port"`
	SMTPHost     string `yaml:"smtp_host"`
	APIPort      int    `yaml:"api_port"`
	DatabaseURL  string `yaml:"database_url"`
	JWTSecret    string `yaml:"jwt_secret"`
	RateLimit    int    `yaml:"rate_limit"`
	MaxFileSize  int    `yaml:"max_file_size"`
	SMTPUsername string `yaml:"smtp_username"`
	SMTPPassword string `yaml:"smtp_password"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}