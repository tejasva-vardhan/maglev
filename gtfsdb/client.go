package gtfsdb

import (
	"context"
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Client is the main entry point for the library
type Client struct {
	config        Config
	DB            *sql.DB
	Queries       *Queries
	importRuntime time.Duration
}

// NewClient creates a new Client with the provided configuration
func NewClient(config Config) *Client {
	db, err := createDB(config)
	if err != nil {
		log.Fatal("Unable to create DB", err)
	} else if config.verbose {
		log.Println("Successfully created tables")
	}

	queries := New(db)

	client := &Client{
		config:  config,
		DB:      db,
		Queries: queries,
	}
	return client
}

func (c *Client) Close() error {
	return c.DB.Close()
}

// DownloadAndStore downloads GTFS data from the given URL and stores it in the database
func (c *Client) DownloadAndStore(ctx context.Context, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = c.processAndStoreGTFSData(b)

	return err
}

// ImportFromFile imports GTFS data from a local zip file into the database
func (c *Client) ImportFromFile(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = c.processAndStoreGTFSData(data)

	return err
}
