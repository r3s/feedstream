package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	AppPort       string
	EmailFrom     string
	ResendAPIKey  string
	SessionSecret string
	CSRFSecret    string
	Environment   string
	AppURL        string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		if _, exists := os.Stat(".env"); exists == nil {
			log.Println("Warning: .env file exists but couldn't be loaded:", err)
		}
	}

	environment := getEnv("ENVIRONMENT", "development")
	sessionSecret := getEnv("SESSION_SECRET", "")
	csrfSecret := getEnv("CSRF_SECRET", "")

	if sessionSecret == "" {
		sessionSecret = generateRandomSecret("SESSION_SECRET")
	}
	if csrfSecret == "" {
		csrfSecret = generateRandomSecret("CSRF_SECRET")
	}

	appPort := getEnv("APP_PORT", "8080")
	appURL := getEnv("APP_URL", "")
	
	if appURL == "" {
		if environment == "production" {
			log.Println("Warning: APP_URL not set in production, CSRF origin validation may fail")
		} else {
			appURL = "http://localhost:" + appPort
		}
	}

	cfg := &Config{
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		AppPort:       appPort,
		EmailFrom:     getEnv("EMAIL_FROM", ""),
		ResendAPIKey:  getEnv("RESEND_API_KEY", ""),
		SessionSecret: sessionSecret,
		CSRFSecret:    csrfSecret,
		Environment:   environment,
		AppURL:        appURL,
	}

	log.Printf("Configuration loaded:")
	log.Printf("  Environment: %s", cfg.Environment)
	log.Printf("  APP_PORT: %s", cfg.AppPort)
	log.Printf("  APP_URL: %s", cfg.AppURL)

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

func generateRandomSecret(name string) string {
	log.Printf("Warning: %s not set, generating random secret (will not persist across restarts)", name)
	
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate random secret for %s: %v", name, err)
	}
	
	return base64.StdEncoding.EncodeToString(b)
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}