package api

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"rss-reader/db"
	"rss-reader/models"
	"rss-reader/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

// LoginHandler handles the login page
func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		r.ParseForm()
		email := r.FormValue("email")
		otp := r.FormValue("otp")

		if otp == "" {
			// Step 1: User provides email, send OTP
			_, err := db.GetUserByEmail(email)
			if err != nil {
				// If user not found, create a new user
				if err.Error() == "sql: no rows in result set" {
					_, err = db.CreateUser(email)
					if err != nil {
						http.Error(w, "Error creating user", http.StatusInternalServerError)
						return
					}
				} else {
					http.Error(w, "Error getting user", http.StatusInternalServerError)
					return
				}
			}

			generatedOTP := utils.GenerateOTP()
			err = db.StoreOTP(email, generatedOTP, time.Now().Add(10*time.Minute))
			if err != nil {
				http.Error(w, "Error storing OTP", http.StatusInternalServerError)
				return
			}

			// Send the OTP via email
			subject := "Your OTP for RSS Reader"
			body := fmt.Sprintf("Your OTP is: %s", generatedOTP)
			err = utils.SendEmail(a.Config, email, subject, body)
			if err != nil {
				log.Printf("Error sending OTP email to %s: %v", email, err)
				log.Printf("Resend Config - API Key exists: %t, From: %s",
					a.Config.ResendAPIKey != "", a.Config.EmailFrom)
				http.Error(w, fmt.Sprintf("Error sending OTP email: %v", err), http.StatusInternalServerError)
				return
			}

			tmpl, err := template.ParseFiles("templates/login.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, map[string]string{"Email": email, "Message": "An OTP has been sent to your email."})

		} else {
			// Step 2: User provides OTP, verify it
			valid, err := db.VerifyOTP(email, otp)
			if err != nil {
				http.Error(w, "Error verifying OTP", http.StatusInternalServerError)
				return
			}

			if valid {
				session, _ := a.Store.Get(r, "session")
				user, err := db.GetUserByEmail(email)
				if err != nil {
					http.Error(w, "Error getting user", http.StatusInternalServerError)
					return
				}
				session.Values["authenticated"] = true
				session.Values["user_id"] = user.ID
				session.Save(r, w)
				http.Redirect(w, r, "/feeds", http.StatusFound)
			} else {
				tmpl, err := template.ParseFiles("templates/login.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				tmpl.Execute(w, map[string]string{"Email": email, "Error": "Invalid OTP"})
			}
		}
	}
}

// LogoutHandler handles the logout functionality
func (a *App) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "session")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// FeedsHandler handles the feeds page
func (a *App) FeedsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "session")
	userID := session.Values["user_id"].(int)

	// Get pagination offset from query parameter
	daysOffset := 0
	if offsetStr := r.URL.Query().Get("days"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			daysOffset = parsed
		}
	}

	items, err := db.GetFeedItemsForUserPaginated(userID, daysOffset)
	if err != nil {
		http.Error(w, "Error getting feed items", http.StatusInternalServerError)
		return
	}

	// Check if there are more items to load
	hasMore, err := db.HasMoreFeedItems(userID, daysOffset)
	if err != nil {
		log.Printf("Error checking for more items: %v", err)
		hasMore = false
	}

	// Group items by date using improved date formatting
	groupedItems := make(map[string][]models.FeedItem)
	dateDisplayMap := make(map[string]string) // Maps grouping key to display string

	for i, item := range items {
		// Convert UTC time to local time for display
		localTime := item.PublishedAt.Local()

		// Debug first few items to see what's happening with times
		if i < 3 {
			log.Printf("DEBUG: Item %d - UTC: %s, Local: %s",
				i, item.PublishedAt.Format("2006-01-02 15:04:05 UTC"),
				localTime.Format("2006-01-02 15:04:05 MST"))
		}

		// Update the item with local time for template use
		items[i].PublishedAt = localTime

		groupKey := utils.FormatDateForGrouping(localTime)
		displayDate := utils.FormatDateForDisplay(localTime)

		groupedItems[groupKey] = append(groupedItems[groupKey], items[i])
		dateDisplayMap[groupKey] = displayDate
	}

	// Extract and sort dates in descending order
	var dates []string
	for date := range groupedItems {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	// Create an ordered slice of date-grouped items for the template
	type DateGroup struct {
		Date  string
		Items []models.FeedItem
	}
	type FeedsPageData struct {
		DateGroups  []DateGroup
		HasMore     bool
		NextOffset  int
		CurrentDays int
	}

	var orderedGroups []DateGroup
	for _, date := range dates {
		displayDate := dateDisplayMap[date]
		orderedGroups = append(orderedGroups, DateGroup{Date: displayDate, Items: groupedItems[date]})
	}

	pageData := FeedsPageData{
		DateGroups:  orderedGroups,
		HasMore:     hasMore,
		NextOffset:  daysOffset + 10,
		CurrentDays: daysOffset,
	}

	// Mark items as old after fetching
	db.MarkItemsAsOld(userID)

	tmpl, err := template.ParseFiles("templates/feeds.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, pageData)
}

// AddFeedHandler handles adding a new feed
func (a *App) AddFeedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("templates/add_feed.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		r.ParseForm()
		name := r.FormValue("name")
		url := r.FormValue("url")

		session, _ := a.Store.Get(r, "session")
		userID := session.Values["user_id"].(int)

		_, err := db.CreateFeed(name, url, userID)
		if err != nil {
			http.Error(w, "Error creating feed", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/feeds", http.StatusFound)
	}
}

// RefreshFeedsHandler handles refreshing the feeds
func (a *App) RefreshFeedsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "session")
	userID := session.Values["user_id"].(int)

	feeds, err := db.GetFeedsForUser(userID)
	if err != nil {
		http.Error(w, "Error getting feeds", http.StatusInternalServerError)
		return
	}

	log.Printf("Refreshing %d feeds for user %d", len(feeds), userID)

	fp := gofeed.NewParser()
	totalItems := 0
	newItems := 0

	for _, feed := range feeds {
		log.Printf("Processing feed: %s (%s)", feed.Name, feed.URL)

		parsedFeed, err := fp.ParseURL(feed.URL)
		if err != nil {
			log.Printf("Error parsing feed %s (%s): %v", feed.Name, feed.URL, err)
			continue
		}

		log.Printf("Feed %s has %d items", feed.Name, len(parsedFeed.Items))

		for _, item := range parsedFeed.Items {
			totalItems++

			// Debug RSS date parsing
			log.Printf("DEBUG RSS: Item '%s' raw date: '%s'", item.Title, item.Published)

			// Use improved date parsing (ParseRSSDate handles fallback internally)
			publishedAt, _ := utils.ParseRSSDate(item.Published)

			log.Printf("DEBUG RSS: Parsed as UTC: %s, Local: %s",
				publishedAt.Format("2006-01-02 15:04:05 UTC"),
				publishedAt.Local().Format("2006-01-02 15:04:05 MST"))

			// Clean up description
			description := item.Description
			if len(description) > 1000 {
				description = description[:1000] + "..."
			}

			feedItem := &models.FeedItem{
				Title:       item.Title,
				Description: description,
				Link:        item.Link,
				FeedID:      feed.ID,
				PublishedAt: utils.NormalizeToUTC(publishedAt),
			}

			err = db.CreateFeedItem(feedItem)
			if err != nil {
				// Check if it's a duplicate (which is expected and OK)
				if !isDuplicateError(err) {
					log.Printf("Error creating feed item '%s': %v", item.Title, err)
				}
			} else {
				newItems++
			}
		}
	}

	log.Printf("Feed refresh complete: processed %d items, %d new/updated", totalItems, newItems)
	http.Redirect(w, r, "/feeds", http.StatusFound)
}

// isDuplicateError checks if the error is due to a duplicate entry
func isDuplicateError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "UNIQUE"))
}

// DebugHandler shows debug information about feed items and pagination
func (a *App) DebugHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := a.Store.Get(r, "session")
	userID := session.Values["user_id"].(int)

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Debug Information for User %d\n", userID)
	fmt.Fprintf(w, "=================================\n\n")

	// Show pagination ranges
	for i := 0; i < 50; i += 10 {
		endDate := time.Now().AddDate(0, 0, -i)
		startDate := endDate.AddDate(0, 0, -10)

		items, err := db.GetFeedItemsForUserPaginated(userID, i)
		if err != nil {
			fmt.Fprintf(w, "Error for offset %d: %v\n", i, err)
			continue
		}

		hasMore, _ := db.HasMoreFeedItems(userID, i)

		fmt.Fprintf(w, "Offset %d days: %s to %s\n", i,
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		fmt.Fprintf(w, "  Items found: %d\n", len(items))
		fmt.Fprintf(w, "  Has more: %t\n", hasMore)

		if len(items) > 0 {
			fmt.Fprintf(w, "  Newest: %s\n", items[0].PublishedAt.Format("2006-01-02 15:04"))
			fmt.Fprintf(w, "  Oldest: %s\n", items[len(items)-1].PublishedAt.Format("2006-01-02 15:04"))
		}
		fmt.Fprintf(w, "\n")

		if len(items) == 0 && !hasMore {
			break
		}
	}
}
