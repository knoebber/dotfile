package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/feeds"
	"github.com/knoebber/dotfile/db"
)

const feedSize = 50

func createRSSFeed(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		items, err := db.FileFeed(db.Connection, feedSize, nil)
		if err != nil {
			setError(w, err, "failed to create RSS feed", http.StatusInternalServerError)
			return
		}
		url := config.URL(r)

		feed := &feeds.Feed{
			Title:       "Dotfilehub",
			Link:        &feeds.Link{Href: url},
			Description: "Recent files",
			Created:     now,
			Items:       make([]*feeds.Item, len(items)),
		}

		for i, item := range items {
			feed.Items[i] = &feeds.Item{
				Title: item.Alias,
				Link: &feeds.Link{
					Href: fmt.Sprintf("%s/%s/%s", url, item.Username, item.Alias),
				},
				Description: item.Path,
				Author: &feeds.Author{
					Name: item.Username,
				},
				Created: item.UpdatedAt,
			}
		}

		if err := feed.WriteRss(w); err != nil {
			setError(w, err, "failed to write RSS feed", http.StatusInternalServerError)
			return
		}
	}
}
