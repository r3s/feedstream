package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"rss-reader/internal/middleware"
	"rss-reader/internal/service"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

type FeedHandler struct {
	feedService         *service.FeedService
	authMiddleware      *middleware.AuthMiddleware
	feedsTemplate       *template.Template
	addFeedTemplate     *template.Template
	manageFeedsTemplate *template.Template
	editFeedTemplate    *template.Template
}

func NewFeedHandler(feedService *service.FeedService, authMiddleware *middleware.AuthMiddleware) *FeedHandler {
	feedsTemplate, err := template.ParseFiles("templates/feeds.html")
	if err != nil {
		log.Fatalf("Failed to parse feeds template: %v", err)
	}

	addFeedTemplate, err := template.ParseFiles("templates/add_feed.html")
	if err != nil {
		log.Fatalf("Failed to parse add_feed template: %v", err)
	}

	manageFeedsTemplate, err := template.ParseFiles("templates/manage_feeds.html")
	if err != nil {
		log.Fatalf("Failed to parse manage_feeds template: %v", err)
	}

	editFeedTemplate, err := template.ParseFiles("templates/edit_feed.html")
	if err != nil {
		log.Fatalf("Failed to parse edit_feed template: %v", err)
	}

	return &FeedHandler{
		feedService:         feedService,
		authMiddleware:      authMiddleware,
		feedsTemplate:       feedsTemplate,
		addFeedTemplate:     addFeedTemplate,
		manageFeedsTemplate: manageFeedsTemplate,
		editFeedTemplate:    editFeedTemplate,
	}
}

func (h *FeedHandler) ViewFeeds(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authMiddleware.GetUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	daysOffset := 0
	if offsetStr := r.URL.Query().Get("days"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			daysOffset = parsed
		}
	}

	// Refresh feeds when viewing the page
	if daysOffset == 0 {
		totalItems, newItems, err := h.feedService.RefreshFeeds(userID)
		if err != nil {
			log.Printf("Error refreshing feeds: %v", err)
		} else {
			log.Printf("Auto-refreshed feeds for user %d: %d total, %d new", userID, totalItems, newItems)
		}
	}

	dateGroups, hasMore, feedNames, err := h.feedService.GetFeedItemsGroupedByDate(userID, daysOffset)
	if err != nil {
		log.Printf("Error getting feed items: %v", err)
		http.Error(w, "Error getting feed items", http.StatusInternalServerError)
		return
	}

	pageData := struct {
		DateGroups  []service.FeedItemGroup
		HasMore     bool
		NextOffset  int
		CurrentDays int
		FeedNames   []string
	}{
		DateGroups:  dateGroups,
		HasMore:     hasMore,
		NextOffset:  daysOffset + 60,
		CurrentDays: daysOffset,
		FeedNames:   feedNames,
	}

	h.feedsTemplate.Execute(w, pageData)
}

func (h *FeedHandler) AddFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.showAddFeedPage(w, r)
		return
	}

	if r.Method == "POST" {
		h.handleAddFeedPost(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *FeedHandler) showAddFeedPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"csrfField": csrf.TemplateField(r),
	}
	
	h.addFeedTemplate.Execute(w, data)
}

func (h *FeedHandler) handleAddFeedPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authMiddleware.GetUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	url := r.FormValue("url")

	_, err := h.feedService.CreateFeed(name, url, userID)
	if err != nil {
		log.Printf("Error creating feed: %v", err)
		http.Error(w, "Error creating feed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/feeds", http.StatusFound)
}

func (h *FeedHandler) RefreshFeeds(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authMiddleware.GetUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	totalItems, newItems, err := h.feedService.RefreshFeeds(userID)
	if err != nil {
		log.Printf("Error refreshing feeds: %v", err)
		http.Error(w, "Error refreshing feeds", http.StatusInternalServerError)
		return
	}

	log.Printf("Refreshed feeds for user %d: %d total, %d new", userID, totalItems, newItems)
	http.Redirect(w, r, "/feeds", http.StatusFound)
}

func (h *FeedHandler) ManageFeeds(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authMiddleware.GetUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	feeds, err := h.feedService.GetFeedsByUserID(userID)
	if err != nil {
		log.Printf("Error getting feeds: %v", err)
		http.Error(w, "Error getting feeds", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Feeds": feeds,
	}
	
	if csrfToken := csrf.Token(r); csrfToken != "" {
		data["csrfField"] = csrf.TemplateField(r)
	}
	
	if err := h.manageFeedsTemplate.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

func (h *FeedHandler) EditFeed(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authMiddleware.GetUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	feedID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid feed ID", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		h.showEditFeedPage(w, r, feedID, userID)
		return
	}

	if r.Method == "POST" {
		h.handleEditFeedPost(w, r, feedID, userID)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *FeedHandler) showEditFeedPage(w http.ResponseWriter, r *http.Request, feedID, userID int) {
	feed, err := h.feedService.GetFeedByID(feedID, userID)
	if err != nil {
		log.Printf("Error getting feed: %v", err)
		http.Error(w, "Feed not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"ID":        feed.ID,
		"Name":      feed.Name,
		"URL":       feed.URL,
		"csrfField": csrf.TemplateField(r),
	}

	h.editFeedTemplate.Execute(w, data)
}

func (h *FeedHandler) handleEditFeedPost(w http.ResponseWriter, r *http.Request, feedID, userID int) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	url := r.FormValue("url")

	err := h.feedService.UpdateFeed(feedID, name, url, userID)
	if err != nil {
		log.Printf("Error updating feed: %v", err)
		http.Error(w, "Error updating feed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/feeds/manage", http.StatusFound)
}

func (h *FeedHandler) DeleteFeed(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authMiddleware.GetUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	vars := mux.Vars(r)
	feedID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid feed ID", http.StatusBadRequest)
		return
	}

	err = h.feedService.DeleteFeed(feedID, userID)
	if err != nil {
		log.Printf("Error deleting feed %d: %v", feedID, err)
		http.Error(w, "Error deleting feed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/feeds/manage", http.StatusFound)
}

func (h *FeedHandler) ExportFeeds(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authMiddleware.GetUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	feeds, err := h.feedService.ExportFeeds(userID)
	if err != nil {
		log.Printf("Error exporting feeds: %v", err)
		http.Error(w, "Error exporting feeds", http.StatusInternalServerError)
		return
	}

	exportData := struct {
		Feeds []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"feeds"`
	}{}

	for _, feed := range feeds {
		exportData.Feeds = append(exportData.Feeds, struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}{
			Name: feed.Name,
			URL:  feed.URL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=feedstream-feeds.json")

	if err := json.NewEncoder(w).Encode(exportData); err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
}

func (h *FeedHandler) ImportFeeds(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authMiddleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var importData struct {
		Feeds []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"feeds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&importData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON format",
		})
		return
	}

	feeds := make([]struct{ Name, URL string }, len(importData.Feeds))
	for i, f := range importData.Feeds {
		feeds[i] = struct{ Name, URL string }{Name: f.Name, URL: f.URL}
	}

	successCount, errors := h.feedService.ImportFeeds(userID, feeds)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":      successCount > 0,
		"imported":     successCount,
		"errors":       len(errors),
		"errorDetails": errors,
	}

	if successCount > 0 && len(errors) > 0 {
		response["message"] = fmt.Sprintf("Imported %d feeds with %d errors", successCount, len(errors))
	} else if successCount > 0 {
		response["message"] = fmt.Sprintf("Successfully imported %d feeds", successCount)
	} else {
		response["error"] = "No feeds were imported"
	}

	json.NewEncoder(w).Encode(response)
}

func (h *FeedHandler) Debug(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authMiddleware.GetUserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Debug Information for User %d\n", userID)
	fmt.Fprintf(w, "=================================\n\n")

	for i := 0; i < 50; i += 10 {
		endDate := time.Now().AddDate(0, 0, -i)
		startDate := endDate.AddDate(0, 0, -10)

		dateGroups, hasMore, _, err := h.feedService.GetFeedItemsGroupedByDate(userID, i)
		if err != nil {
			fmt.Fprintf(w, "Error for offset %d: %v\n", i, err)
			continue
		}

		totalItems := 0
		for _, group := range dateGroups {
			totalItems += len(group.Items)
		}

		fmt.Fprintf(w, "Offset %d days: %s to %s\n", i,
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		fmt.Fprintf(w, "  Items found: %d\n", totalItems)
		fmt.Fprintf(w, "  Has more: %t\n", hasMore)
		fmt.Fprintf(w, "\n")

		if totalItems == 0 && !hasMore {
			break
		}
	}
}