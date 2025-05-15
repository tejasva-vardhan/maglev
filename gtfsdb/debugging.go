package gtfsdb

import (
	"fmt"
	"github.com/jamespfennell/gtfs"
	"log"
)

func (c *Client) staticDataCounts(staticData *gtfs.Static) map[string]int {
	return map[string]int{
		"routes":    len(staticData.Routes),
		"services":  len(staticData.Services),
		"stops":     len(staticData.Stops),
		"agencies":  len(staticData.Agencies),
		"transfers": len(staticData.Transfers),
		"trips":     len(staticData.Trips),
		"calendar":  len(staticData.Services),
		"shapes":    len(staticData.Shapes),
	}
}

func (c *Client) TableCounts() (map[string]int, error) {
	rows, err := c.DB.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		log.Fatalf("Failed to query table names: %v", err)
	}
	defer rows.Close() // nolint:errcheck
	var tables []string

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("Failed to scan table name: %v", err)
		}
		tables = append(tables, tableName)
	}

	counts := make(map[string]int)

	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err := c.DB.QueryRow(query).Scan(&count)
		if err != nil {
			return nil, err
		}
		counts[table] = count
	}

	return counts, nil
}
