package domain

import "time"

type FeedItem struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	FeedID      int       `json:"feed_id"`
	FeedName    string    `json:"feed_name"`
	PublishedAt time.Time `json:"published_at"`
	IsNew       bool      `json:"is_new"`
}

func (fi *FeedItem) Validate() error {
	if fi.Title == "" {
		return ErrInvalidFeedItemTitle
	}
	if fi.Link == "" {
		return ErrInvalidFeedItemLink
	}
	if fi.FeedID <= 0 {
		return ErrInvalidFeedID
	}
	return nil
}