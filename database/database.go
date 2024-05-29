package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

type FeedObject struct {
	TxID      string `json:"txID"`
	SubnetID  string `json:"subnetID"`
	ChainID   string `json:"chainID"`
	Address   string `json:"address"`
	Timestamp int64  `json:"timestamp"`
	Fee       uint64 `json:"fee"`
	Content   string `json:"content"` // JSON-encoded content
}

func NewDB(conn *sql.DB) (*DB, error) {
	db := &DB{conn: conn}

	query := `CREATE TABLE IF NOT EXISTS feeds (
		txid TEXT PRIMARY KEY,
		subnetID TEXT,
		chainID TEXT,
		address TEXT,
		timestamp BIGINT,
		fee BIGINT,
		content TEXT
	)`
	_, err := db.conn.Exec(query)
	if err != nil {
		log.Printf("Error creating table: %v", err)
		return nil, err
	}

	log.Println("Database initialized successfully")
	return db, nil
}

func (db *DB) SaveFeed(feed *FeedObject) error {
	log.Printf("Saving feed with TxID: %s", feed.TxID)
	query := `INSERT INTO feeds (txid, subnetID, chainID, address, timestamp, fee, content) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.conn.Exec(query, feed.TxID, feed.SubnetID, feed.ChainID, feed.Address, feed.Timestamp, feed.Fee, feed.Content)
	if err != nil {
		log.Printf("Error saving feed: %v", err)
	}
	return err
}

func (db *DB) GetFeed(txID string) (*FeedObject, error) {
	var feed FeedObject
	query := `SELECT txid, subnetID, chainID, address, timestamp, fee, content FROM feeds WHERE txid = $1`
	row := db.conn.QueryRow(query, txID)
	err := row.Scan(&feed.TxID, &feed.SubnetID, &feed.ChainID, &feed.Address, &feed.Timestamp, &feed.Fee, &feed.Content)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No feed found with TxID: %s", txID)
		} else {
			log.Printf("Error fetching feed: %v", err)
		}
		return nil, err
	}
	return &feed, nil
}

func (db *DB) GetAllFeeds() ([]FeedObject, error) {
	var feeds []FeedObject
	query := `SELECT txid, subnetID, chainID, address, timestamp, fee, content FROM feeds`
	rows, err := db.conn.Query(query)
	if err != nil {
		log.Printf("Error fetching all feeds: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var feed FeedObject
		if err := rows.Scan(&feed.TxID, &feed.SubnetID, &feed.ChainID, &feed.Address, &feed.Timestamp, &feed.Fee, &feed.Content); err != nil {
			log.Printf("Error scanning feed row: %v", err)
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error in rows: %v", err)
		return nil, err
	}

	return feeds, nil
}

func (db *DB) GetFeedsByUser(address string) ([]FeedObject, error) {
	var feeds []FeedObject
	query := `SELECT txid, subnetID, chainID, address, timestamp, fee, content FROM feeds WHERE address = $1`
	rows, err := db.conn.Query(query, address)
	if err != nil {
		log.Printf("Error fetching feeds by user: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var feed FeedObject
		if err := rows.Scan(&feed.TxID, &feed.SubnetID, &feed.ChainID, &feed.Address, &feed.Timestamp, &feed.Fee, &feed.Content); err != nil {
			log.Printf("Error scanning feed row: %v", err)
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error in rows: %v", err)
		return nil, err
	}

	return feeds, nil
}

func (db *DB) GetLastFeeds(limit int) ([]FeedObject, error) {
	var feeds []FeedObject
	query := `SELECT txid, subnetID, chainID, address, timestamp, fee, content FROM feeds ORDER BY timestamp DESC LIMIT $1`
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		log.Printf("Error fetching last feeds: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var feed FeedObject
		if err := rows.Scan(&feed.TxID, &feed.SubnetID, &feed.ChainID, &feed.Address, &feed.Timestamp, &feed.Fee, &feed.Content); err != nil {
			log.Printf("Error scanning feed row: %v", err)
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error in rows: %v", err)
		return nil, err
	}

	return feeds, nil
}

func (db *DB) Close() {
	log.Println("Closing database connection")
	db.conn.Close()
}
