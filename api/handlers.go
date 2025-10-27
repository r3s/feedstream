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
				log.Printf("SMTP Config - Host: %s, Port: %s, Username: %s",
					a.Config.SMTPHost, a.Config.SMTPPort, a.Config.SMTPUsername)
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

	items, err := db.GetFeedItemsForUser(userID)
	if err != nil {
		http.Error(w, "Error getting feed items", http.StatusInternalServerError)
		return
	}

	// Group items by date
	groupedItems := make(map[string][]models.FeedItem)
	for _, item := range items {
		date := item.PublishedAt.Format("2006-01-02")
		groupedItems[date] = append(groupedItems[date], item)
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
	var orderedGroups []DateGroup
	for _, date := range dates {
		orderedGroups = append(orderedGroups, DateGroup{Date: date, Items: groupedItems[date]})
	}

	// Mark items as old after fetching
	db.MarkItemsAsOld(userID)

	tmpl, err := template.ParseFiles("templates/feeds.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, orderedGroups)
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

	fp := gofeed.NewParser()
	for _, feed := range feeds {
		parsedFeed, err := fp.ParseURL(feed.URL)
		if err != nil {
			log.Printf("Error parsing feed %s: %s", feed.Name, err)
			continue
		}

		for _, item := range parsedFeed.Items {
			publishedAt, err := time.Parse(time.RFC1123Z, item.Published)
			if err != nil {
				log.Printf("Error parsing date for item %s: %s", item.Title, err)
				publishedAt = time.Now()
			}

			feedItem := &models.FeedItem{
				Title:       item.Title,
				Description: item.Description,
				Link:        item.Link,
				FeedID:      feed.ID,
				PublishedAt: publishedAt,
			}
			err = db.CreateFeedItem(feedItem)
			if err != nil {
				log.Printf("Error creating feed item: %s", err)
			}
		}
	}

	http.Redirect(w, r, "/feeds", http.StatusFound)
}
