package repository

import (
	"database/sql"
	"fmt"
	"rss-reader/internal/domain"
	"strings"
	"time"
)

type FeedItemRepository interface {
	Create(item *domain.FeedItem) error
	GetByUserIDPaginated(userID int, daysOffset int) ([]domain.FeedItem, error)
	HasMoreItems(userID int, daysOffset int) (bool, error)
	MarkAllAsOld(userID int) error
}

type feedItemRepository struct {
	db *sql.DB
}

func NewFeedItemRepository(db *sql.DB) FeedItemRepository {
	return &feedItemRepository{db: db}
}

func (r *feedItemRepository) Create(item *domain.FeedItem) error {
	_, err := r.db.Exec(`
		INSERT INTO feed_items (title, description, link, feed_id, published_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (link, feed_id) DO UPDATE SET
		title = EXCLUDED.title,
		description = EXCLUDED.description,
		published_at = EXCLUDED.published_at,
		is_new = CASE
			WHEN feed_items.title != EXCLUDED.title OR
				 feed_items.description != EXCLUDED.description THEN TRUE
			ELSE feed_items.is_new
		END`,
		item.Title, item.Description, item.Link, item.FeedID, item.PublishedAt)

	if err != nil {
		if isDuplicateError(err) {
			return nil
		}
		return fmt.Errorf("failed to create feed item: %w", err)
	}

	return nil
}

func (r *feedItemRepository) GetByUserIDPaginated(userID int, daysOffset int) ([]domain.FeedItem, error) {
	endDate := time.Now().AddDate(0, 0, -daysOffset)
	startDate := endDate.AddDate(0, 0, -60)

	rows, err := r.db.Query(`
		SELECT i.id, i.title, i.description, i.link, f.name,
			   i.published_at AT TIME ZONE 'UTC' as published_at, i.is_new
		FROM feed_items i
		JOIN feeds f ON i.feed_id = f.id
		WHERE f.user_id = $1
		AND i.published_at <= $2
		AND i.published_at >= $3
		ORDER BY i.published_at DESC
	`, userID, endDate, startDate)

	if err != nil {
		return nil, fmt.Errorf("failed to get feed items: %w", err)
	}
	defer rows.Close()

	var items []domain.FeedItem
	for rows.Next() {
		var item domain.FeedItem
		var feedName string
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Link,
			&feedName,
			&item.PublishedAt,
			&item.IsNew,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed item: %w", err)
		}
		item.FeedName = feedName
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feed items: %w", err)
	}

	return items, nil
}

func (r *feedItemRepository) HasMoreItems(userID int, daysOffset int) (bool, error) {
	checkDate := time.Now().AddDate(0, 0, -(daysOffset + 60))

	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM feed_items i
		JOIN feeds f ON i.feed_id = f.id
		WHERE f.user_id = $1 AND i.published_at < $2
	`, userID, checkDate).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check for more items: %w", err)
	}

	return count > 0, nil
}

func (r *feedItemRepository) MarkAllAsOld(userID int) error {
	_, err := r.db.Exec(`
		UPDATE feed_items
		SET is_new = FALSE
		WHERE feed_id IN (SELECT id FROM feeds WHERE user_id = $1)
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to mark items as old: %w", err)
	}

	return nil
}

func isDuplicateError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "UNIQUE"))
}