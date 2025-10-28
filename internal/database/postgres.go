package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Manager struct {
	DB *sql.DB
}

type Config struct {
	ConnectionString string
	Host             string
	Port             string
	User             string
	Password         string
	DBName           string
}

func NewManager(cfg Config) (*Manager, error) {
	var connectionString string

	if cfg.ConnectionString != "" {
		connectionString = cfg.ConnectionString
	} else {
		connectionString = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
		)
	}

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to the database")

	manager := &Manager{DB: db}

	if err := manager.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return manager, nil
}

func (m *Manager) runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS feeds (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			url TEXT NOT NULL,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS feed_items (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			link TEXT NOT NULL,
			feed_id INTEGER REFERENCES feeds(id) ON DELETE CASCADE,
			published_at TIMESTAMP WITH TIME ZONE,
			is_new BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(link, feed_id)
		)`,
		`CREATE TABLE IF NOT EXISTS otps (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL,
			otp TEXT NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_feed_items_feed_id ON feed_items(feed_id)`,
		`CREATE INDEX IF NOT EXISTS idx_feed_items_published_at ON feed_items(published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_feeds_user_id ON feeds(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_otps_email ON otps(email, expires_at DESC)`,
	}

	for i, migration := range migrations {
		if _, err := m.DB.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func (m *Manager) Close() error {
	if m.DB != nil {
		return m.DB.Close()
	}
	return nil
}

func (m *Manager) GetDB() *sql.DB {
	return m.DB
}