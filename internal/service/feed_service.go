package service

import (
	"fmt"
	"log"
	"rss-reader/internal/domain"
	"rss-reader/internal/repository"
	"rss-reader/pkg/datetime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type FeedService struct {
	feedRepo      repository.FeedRepository
	feedItemRepo  repository.FeedItemRepository
	dateFormatter *datetime.Formatter
	lastCleanup   time.Time
	cleanupMu     sync.Mutex
}

func NewFeedService(
	feedRepo repository.FeedRepository,
	feedItemRepo repository.FeedItemRepository,
	dateFormatter *datetime.Formatter,
) *FeedService {
	return &FeedService{
		feedRepo:     feedRepo,
		feedItemRepo: feedItemRepo,
		dateFormatter: dateFormatter,
	}
}

func (s *FeedService) CreateFeed(name, url string, userID int) (*domain.Feed, error) {
	feed := &domain.Feed{
		Name:   name,
		URL:    url,
		UserID: userID,
	}
	if err := feed.Validate(); err != nil {
		return nil, err
	}

	exists, err := s.feedRepo.ExistsByURL(userID, url)
	if err != nil {
		return nil, fmt.Errorf("failed to check feed existence: %w", err)
	}
	if exists {
		return nil, domain.ErrFeedAlreadyExists
	}

	createdFeed, err := s.feedRepo.Create(name, url, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create feed: %w", err)
	}

	return createdFeed, nil
}

func (s *FeedService) GetFeedsByUserID(userID int) ([]domain.Feed, error) {
	feeds, err := s.feedRepo.GetAllByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feeds: %w", err)
	}
	return feeds, nil
}

func (s *FeedService) GetFeedByID(feedID, userID int) (*domain.Feed, error) {
	feed, err := s.feedRepo.GetByID(feedID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feed: %w", err)
	}
	return feed, nil
}

func (s *FeedService) UpdateFeed(feedID int, name, url string, userID int) error {
	feed := &domain.Feed{
		ID:     feedID,
		Name:   name,
		URL:    url,
		UserID: userID,
	}
	if err := feed.Validate(); err != nil {
		return err
	}

	if err := s.feedRepo.Update(feedID, name, url, userID); err != nil {
		return fmt.Errorf("failed to update feed: %w", err)
	}

	return nil
}

func (s *FeedService) DeleteFeed(feedID, userID int) error {
	if err := s.feedRepo.Delete(feedID, userID); err != nil {
		return fmt.Errorf("failed to delete feed: %w", err)
	}
	return nil
}

func (s *FeedService) RefreshFeeds(userID int) (int, int, error) {
	s.cleanupMu.Lock()
	shouldCleanup := time.Since(s.lastCleanup) > 24*time.Hour
	if shouldCleanup {
		s.lastCleanup = time.Now()
	}
	s.cleanupMu.Unlock()

	if shouldCleanup {
		deleted, err := s.feedItemRepo.DeleteOlderThan(90)
		if err != nil {
			log.Printf("Warning: cleanup failed: %v", err)
		} else if deleted > 0 {
			log.Printf("Cleaned up %d old feed items (90+ days)", deleted)
		}
	}

	feeds, err := s.feedRepo.GetAllByUserID(userID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get feeds: %w", err)
	}

	log.Printf("Refreshing %d feeds for user %d", len(feeds), userID)

	parser := gofeed.NewParser()
	totalItems := 0
	newItems := 0

	for _, feed := range feeds {
		log.Printf("Processing feed: %s (%s)", feed.Name, feed.URL)

		parsedFeed, err := parser.ParseURL(feed.URL)
		if err != nil {
			log.Printf("Error parsing feed %s (%s): %v", feed.Name, feed.URL, err)
			continue
		}

		log.Printf("Feed %s has %d items", feed.Name, len(parsedFeed.Items))

		for _, item := range parsedFeed.Items {
			totalItems++

			publishedAt, _ := s.dateFormatter.ParseRSSDate(item.Published)

			description := stripHTMLTags(item.Description)
			if len(description) > 1000 {
				description = description[:1000] + "..."
			}

			feedItem := &domain.FeedItem{
				Title:       item.Title,
				Description: description,
				Link:        item.Link,
				FeedID:      feed.ID,
				PublishedAt: s.dateFormatter.NormalizeToUTC(publishedAt),
			}

			if err := feedItem.Validate(); err != nil {
				log.Printf("Invalid feed item '%s': %v", item.Title, err)
				continue
			}

			err = s.feedItemRepo.Create(feedItem)
			if err != nil {
				log.Printf("Error creating feed item '%s': %v", item.Title, err)
			} else {
				newItems++
			}
		}
	}

	log.Printf("Feed refresh complete: processed %d items, %d new/updated", totalItems, newItems)
	return totalItems, newItems, nil
}

type FeedItemGroup struct {
	Date  string
	Items []domain.FeedItem
}

func (s *FeedService) GetFeedItemsGroupedByDate(userID int, daysOffset int) ([]FeedItemGroup, bool, []string, error) {
	items, err := s.feedItemRepo.GetByUserIDPaginated(userID, daysOffset)
	if err != nil {
		return nil, false, nil, fmt.Errorf("failed to get feed items: %w", err)
	}

	hasMore, err := s.feedItemRepo.HasMoreItems(userID, daysOffset)
	if err != nil {
		log.Printf("Error checking for more items: %v", err)
		hasMore = false
	}

	feedNamesMap := make(map[string]bool)
	for _, item := range items {
		feedNamesMap[item.FeedName] = true
	}

	var feedNames []string
	for feedName := range feedNamesMap {
		feedNames = append(feedNames, feedName)
	}
	sort.Strings(feedNames)

	groupedItems := make(map[string][]domain.FeedItem)
	dateDisplayMap := make(map[string]string)

	for i, item := range items {
		localTime := item.PublishedAt.Local()
		items[i].PublishedAt = localTime

		groupKey := s.dateFormatter.FormatForGrouping(localTime)
		displayDate := s.dateFormatter.FormatForDisplay(localTime)

		groupedItems[groupKey] = append(groupedItems[groupKey], items[i])
		dateDisplayMap[groupKey] = displayDate
	}

	var dates []string
	for date := range groupedItems {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	var orderedGroups []FeedItemGroup
	for _, date := range dates {
		displayDate := dateDisplayMap[date]
		orderedGroups = append(orderedGroups, FeedItemGroup{
			Date:  displayDate,
			Items: groupedItems[date],
		})
	}

	if err := s.feedItemRepo.MarkAllAsOld(userID); err != nil {
		log.Printf("Warning: failed to mark items as old: %v", err)
	}

	return orderedGroups, hasMore, feedNames, nil
}

func (s *FeedService) ImportFeeds(userID int, feeds []struct{ Name, URL string }) (int, []string) {
	successCount := 0
	var errors []string

	for _, feedData := range feeds {
		if feedData.Name == "" || feedData.URL == "" {
			errors = append(errors, "Feed missing name or URL")
			continue
		}

		exists, err := s.feedRepo.ExistsByURL(userID, feedData.URL)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error checking feed %s: %v", feedData.Name, err))
			continue
		}

		if exists {
			errors = append(errors, fmt.Sprintf("Feed already exists: %s", feedData.Name))
			continue
		}

		_, err = s.feedRepo.Create(feedData.Name, feedData.URL, userID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Error creating feed %s: %v", feedData.Name, err))
			continue
		}

		successCount++
	}

	return successCount, errors
}

func (s *FeedService) ExportFeeds(userID int) ([]domain.Feed, error) {
	feeds, err := s.feedRepo.GetAllByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to export feeds: %w", err)
	}
	return feeds, nil
}

func stripHTMLTags(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}

	text := doc.Text()
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}