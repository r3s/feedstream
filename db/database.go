package db

import (
	"database/sql"
	"fmt"
	"log"
	"rss-reader/config"
	"rss-reader/models"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB(cfg *config.Config) {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to the database")

	createTables()
}

// createTables creates the database tables if they don't exist
func createTables() {
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	createFeedTable := `
	CREATE TABLE IF NOT EXISTS feeds (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		user_id INTEGER REFERENCES users(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	createFeedItemTable := `
	CREATE TABLE IF NOT EXISTS feed_items (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		link TEXT NOT NULL,
		feed_id INTEGER REFERENCES feeds(id),
		published_at TIMESTAMP WITH TIME ZONE,
		is_new BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	createOTPTable := `
	CREATE TABLE IF NOT EXISTS otps (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL,
		otp TEXT NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL
	);`

	_, err := DB.Exec(createUserTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(createFeedTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(createFeedItemTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(createOTPTable)
	if err != nil {
		log.Fatal(err)
	}
}

// CreateUser creates a new user in the database
func CreateUser(email string) (*models.User, error) {
	user := &models.User{Email: email}
	err := DB.QueryRow("INSERT INTO users (email) VALUES ($1) RETURNING id, created_at", email).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by their email address
func GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow("SELECT id, email, created_at FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// StoreOTP stores the OTP in the database
func StoreOTP(email, otp string, expiresAt time.Time) error {
	_, err := DB.Exec("INSERT INTO otps (email, otp, expires_at) VALUES ($1, $2, $3)", email, otp, expiresAt)
	return err
}

// VerifyOTP verifies the OTP
func VerifyOTP(email, otp string) (bool, error) {
	var storedOTP string
	var expiresAt time.Time
	err := DB.QueryRow("SELECT otp, expires_at FROM otps WHERE email = $1 ORDER BY expires_at DESC LIMIT 1", email).Scan(&storedOTP, &expiresAt)
	if err != nil {
		return false, err
	}

	if storedOTP == otp && time.Now().Before(expiresAt) {
		// OTP is valid, delete it so it can't be used again
		_, err = DB.Exec("DELETE FROM otps WHERE email = $1", email)
		return true, err
	}

	return false, nil
}

// CreateFeed creates a new feed in the database
func CreateFeed(name, url string, userID int) (*models.Feed, error) {
	feed := &models.Feed{Name: name, URL: url, UserID: userID}
	err := DB.QueryRow("INSERT INTO feeds (name, url, user_id) VALUES ($1, $2, $3) RETURNING id, created_at", name, url, userID).Scan(&feed.ID, &feed.CreatedAt)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

// GetFeedsForUser retrieves all feeds for a user
func GetFeedsForUser(userID int) ([]models.Feed, error) {
	rows, err := DB.Query("SELECT id, name, url, user_id, created_at FROM feeds WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var feed models.Feed
		err := rows.Scan(&feed.ID, &feed.Name, &feed.URL, &feed.UserID, &feed.CreatedAt)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, nil
}

// CreateFeedItem creates a new feed item in the database
func CreateFeedItem(item *models.FeedItem) error {
	_, err := DB.Exec("INSERT INTO feed_items (title, description, link, feed_id, published_at) VALUES ($1, $2, $3, $4, $5)",
		item.Title, item.Description, item.Link, item.FeedID, item.PublishedAt)
	return err
}

// GetFeedItemsForUser retrieves all feed items for a user
func GetFeedItemsForUser(userID int) ([]models.FeedItem, error) {
	rows, err := DB.Query(`
		SELECT i.id, i.title, i.description, i.link, f.name, i.published_at, i.is_new
		FROM feed_items i
		JOIN feeds f ON i.feed_id = f.id
		WHERE f.user_id = $1
		ORDER BY i.published_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.FeedItem
	for rows.Next() {
		var item models.FeedItem
		var feedName string
		err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Link, &feedName, &item.PublishedAt, &item.IsNew)
		if err != nil {
			return nil, err
		}
		item.FeedName = feedName
		items = append(items, item)
	}
	return items, nil
}

// MarkItemsAsOld marks all feed items for a user as old
func MarkItemsAsOld(userID int) error {
	_, err := DB.Exec(`
		UPDATE feed_items
		SET is_new = FALSE
		WHERE feed_id IN (SELECT id FROM feeds WHERE user_id = $1)
	`, userID)
	return err
}
