package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
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

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	query := `CREATE TABLE IF NOT EXISTS feeds (
		txid TEXT PRIMARY KEY,
		subnetID TEXT,
		chainID TEXT,
		address TEXT,
		timestamp INTEGER,
		fee INTEGER,
		content TEXT
	)`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil
}

func (db *DB) SaveFeed(feed *FeedObject) error {
	query := `INSERT INTO feeds (txid, subnetID, chainID, address, timestamp, fee, content) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(query, feed.TxID, feed.SubnetID, feed.ChainID, feed.Address, feed.Timestamp, feed.Fee, feed.Content)
	return err
}

func (db *DB) GetFeed(txID string) (*FeedObject, error) {
	var feed FeedObject
	query := `SELECT txid, subnetID, chainID, address, timestamp, fee, content FROM feeds WHERE txid = ?`
	row := db.conn.QueryRow(query, txID)
	err := row.Scan(&feed.TxID, &feed.SubnetID, &feed.ChainID, &feed.Address, &feed.Timestamp, &feed.Fee, &feed.Content)
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

func (db *DB) GetAllFeeds() ([]FeedObject, error) {
	var feeds []FeedObject
	query := `SELECT txid, subnetID, chainID, address, timestamp, fee, content FROM feeds`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var feed FeedObject
		if err := rows.Scan(&feed.TxID, &feed.SubnetID, &feed.ChainID, &feed.Address, &feed.Timestamp, &feed.Fee, &feed.Content); err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	return feeds, nil
}

func (db *DB) GetFeedsByUser(address string) ([]FeedObject, error) {
	var feeds []FeedObject
	query := `SELECT txid, subnetID, chainID, address, timestamp, fee, content FROM feeds WHERE address = ?`
	rows, err := db.conn.Query(query, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var feed FeedObject
		if err := rows.Scan(&feed.TxID, &feed.SubnetID, &feed.ChainID, &feed.Address, &feed.Timestamp, &feed.Fee, &feed.Content); err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	return feeds, nil
}

func (db *DB) GetLastFeeds(limit int) ([]FeedObject, error) {
	var feeds []FeedObject
	query := `SELECT txid, subnetID, chainID, address, timestamp, fee, content FROM feeds ORDER BY timestamp DESC LIMIT ?`
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var feed FeedObject
		if err := rows.Scan(&feed.TxID, &feed.SubnetID, &feed.ChainID, &feed.Address, &feed.Timestamp, &feed.Fee, &feed.Content); err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	return feeds, nil
}

func (db *DB) Close() {
	db.conn.Close()
}
