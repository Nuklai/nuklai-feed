package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
)

const (
	feedBucket = "feedBucket"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Initialize BoltDB
func (m *Manager) initDB() error {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default paths")
	}
	// Set default values to the current directory
	defaultDir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory: " + err.Error())
	}

	databaseFolder := getEnv("NUKLAI_FEED_DB_PATH", filepath.Join(defaultDir, ".nuklai-feed/db"))
	if err := os.MkdirAll(databaseFolder, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create database directory '%s': %w", databaseFolder, err)
	}

	dbPath := filepath.Join(databaseFolder, "feeds.db")
	m.db, err = bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return err
	}

	return m.db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(feedBucket))
		return err
	})
}

// Save feed to BoltDB
func (m *Manager) saveFeed(feed *FeedObject) error {
	return m.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(feedBucket))
		data, err := json.Marshal(feed)
		if err != nil {
			return err
		}
		id := feed.TxID.String()
		return b.Put([]byte(id), data)
	})
}

// Get the last n feeds
func (m *Manager) getLastFeeds(n int) ([]*FeedObject, error) {
	var feeds []*FeedObject
	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(feedBucket))
		c := b.Cursor()

		// Iterate from the end to get the last n feeds
		for k, v := c.Last(); k != nil && len(feeds) < n; k, v = c.Prev() {
			var feed FeedObject
			if err := json.Unmarshal(v, &feed); err != nil {
				return err
			}
			feeds = append(feeds, &feed)
		}

		return nil
	})

	return feeds, err
}

// Append new feed and save it
func (m *Manager) appendFeed(feed *FeedObject) {
	// Save to database
	if err := m.saveFeed(feed); err != nil {
		m.log.Error("Failed to save feed", zap.Error(err))
	}
}
