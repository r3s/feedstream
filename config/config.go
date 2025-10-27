package config

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	DatabaseURL   string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	AppPort       string
	EmailFrom     string
	SMTPHost      string
	SMTPPort      string
	SMTPUsername  string
	SMTPPassword  string
	SessionSecret string
}

// Load loads the configuration from environment variables
func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	cfg := &Config{
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		AppPort:       getEnv("APP_PORT", "8080"),
		EmailFrom:     getEnv("EMAIL_FROM", ""),
		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnv("SMTP_PORT", ""),
		SMTPUsername:  getEnv("SMTP_USERNAME", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		SessionSecret: getEnv("SESSION_SECRET", "something-very-secret"),
	}

	// Parse DATABASE_URL if provided, otherwise use individual DB vars
	if cfg.DatabaseURL != "" {
		cfg.parseDBURL()
	} else {
		cfg.DBHost = getEnv("DB_HOST", "localhost")
		cfg.DBPort = getEnv("DB_PORT", "5432")
		cfg.DBUser = getEnv("DB_USER", "postgres")
		cfg.DBPassword = getEnv("DB_PASSWORD", "password")
		cfg.DBName = getEnv("DB_NAME", "rss_reader")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// parseDBURL parses DATABASE_URL and sets individual DB config fields
func (c *Config) parseDBURL() {
	u, err := url.Parse(c.DatabaseURL)
	if err != nil {
		log.Printf("Error parsing DATABASE_URL: %v", err)
		return
	}

	c.DBHost = u.Hostname()
	c.DBPort = u.Port()
	if c.DBPort == "" {
		c.DBPort = "5432"
	}

	c.DBUser = u.User.Username()
	if password, ok := u.User.Password(); ok {
		c.DBPassword = password
	}

	c.DBName = strings.TrimPrefix(u.Path, "/")
}
