package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type User struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
	Domain   string `yaml:"domain"`
}

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
	Users        []User `yaml:"users"`
	DefaultUser  User   `yaml:"default_user"`
	SSLCertFile  string `yaml:"ssl_cert_file"`
	SSLKeyFile       string `yaml:"ssl_key_file"`
	DevelopmentMode  bool   `yaml:"development_mode"`
	SMTPReadTimeout  int    `yaml:"smtp_read_timeout"`
	SMTPWriteTimeout int    `yaml:"smtp_write_timeout"`
	MaxMessageSize   int    `yaml:"max_message_size"`
	MaxRecipients    int    `yaml:"max_recipients"`
	MaxLineLength    int    `yaml:"max_line_length"`
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
		return fmt.Errorf("database url is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if c.RateLimit == 0 {
		return fmt.Errorf("rate limit is required")
	}
	if c.MaxFileSize == 0 {
		return fmt.Errorf("max file size is required")
	}
	if c.SMTPUsername == "" {
		return fmt.Errorf("smtp username is required")
	}
	if c.SMTPPassword == "" {
		return fmt.Errorf("smtp password is required")
	}
	if c.SMTPReadTimeout == 0 {
		return fmt.Errorf("smtp read timeout is required")
	}
	if c.SMTPWriteTimeout == 0 {
		return fmt.Errorf("smtp write timeout is required")
	}
	if c.MaxMessageSize == 0 {
		return fmt.Errorf("max message size is required")
	}
	if c.MaxRecipients == 0 {
		return fmt.Errorf("max recipients is required")
	}
	if c.MaxLineLength == 0 {
		return fmt.Errorf("max line length is required")
	}
	if !c.DevelopmentMode && (c.SSLCertFile == "" || c.SSLKeyFile == "") {
		return fmt.Errorf("SSL certificate and key file paths must be provided in production mode")
	}

	return nil
}