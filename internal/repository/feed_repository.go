package repository

import (
	"database/sql"
	"fmt"
	"rss-reader/internal/domain"
)

type FeedRepository interface {
	Create(name, url string, userID int) (*domain.Feed, error)
	GetByID(feedID, userID int) (*domain.Feed, error)
	GetAllByUserID(userID int) ([]domain.Feed, error)
	Update(feedID int, name, url string, userID int) error
	Delete(feedID, userID int) error
	ExistsByURL(userID int, url string) (bool, error)
}

type feedRepository struct {
	db *sql.DB
}

func NewFeedRepository(db *sql.DB) FeedRepository {
	return &feedRepository{db: db}
}

func (r *feedRepository) Create(name, url string, userID int) (*domain.Feed, error) {
	feed := &domain.Feed{
		Name:   name,
		URL:    url,
		UserID: userID,
	}

	err := r.db.QueryRow(
		"INSERT INTO feeds (name, url, user_id) VALUES ($1, $2, $3) RETURNING id, created_at",
		name, url, userID,
	).Scan(&feed.ID, &feed.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create feed: %w", err)
	}

	return feed, nil
}

func (r *feedRepository) GetByID(feedID, userID int) (*domain.Feed, error) {
	feed := &domain.Feed{}

	err := r.db.QueryRow(
		"SELECT id, name, url, user_id, created_at FROM feeds WHERE id = $1 AND user_id = $2",
		feedID, userID,
	).Scan(&feed.ID, &feed.Name, &feed.URL, &feed.UserID, &feed.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrFeedNotFound
		}
		return nil, fmt.Errorf("failed to get feed: %w", err)
	}

	return feed, nil
}

func (r *feedRepository) GetAllByUserID(userID int) ([]domain.Feed, error) {
	rows, err := r.db.Query(
		"SELECT id, name, url, user_id, created_at FROM feeds WHERE user_id = $1 ORDER BY name",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get feeds: %w", err)
	}
	defer rows.Close()

	var feeds []domain.Feed
	for rows.Next() {
		var feed domain.Feed
		err := rows.Scan(&feed.ID, &feed.Name, &feed.URL, &feed.UserID, &feed.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed: %w", err)
		}
		feeds = append(feeds, feed)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feeds: %w", err)
	}

	return feeds, nil
}

func (r *feedRepository) Update(feedID int, name, url string, userID int) error {
	result, err := r.db.Exec(
		"UPDATE feeds SET name = $1, url = $2 WHERE id = $3 AND user_id = $4",
		name, url, feedID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update feed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrFeedNotFound
	}

	return nil
}

func (r *feedRepository) Delete(feedID, userID int) error {
	result, err := r.db.Exec(
		"DELETE FROM feeds WHERE id = $1 AND user_id = $2",
		feedID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete feed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrFeedNotFound
	}

	return nil
}

func (r *feedRepository) ExistsByURL(userID int, url string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM feeds WHERE user_id = $1 AND url = $2",
		userID, url,
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check feed existence: %w", err)
	}

	return count > 0, nil
}