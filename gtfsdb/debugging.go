package gtfsdb

import (
	"database/sql"
	"fmt"
	"github.com/OneBusAway/go-gtfs"
	"log"
	"log/slog"
	"maglev.onebusaway.org/internal/logging"
	"strings"
)

func PrintSimpleSchema(db *sql.DB) error { // nolint:unused
	// Get all database objects
	rows, err := db.Query(`
		SELECT type, name, sql
		FROM sqlite_master
		WHERE type IN ('table', 'index', 'view', 'trigger')
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY type, name
	`)
	if err != nil {
		return err
	}
	defer logging.SafeCloseWithLogging(rows,
		slog.Default().With(slog.String("component", "debugging")),
		"database_rows")

	log.Println("DATABASE SCHEMA:")
	log.Println("----------------")

	for rows.Next() {
		var objType, objName, objSQL string
		if err := rows.Scan(&objType, &objName, &objSQL); err != nil {
			return err
		}
		log.Printf("%s: %s\n", strings.ToUpper(objType), objName)
		log.Printf("%s\n\n", objSQL)
	}

	return nil
}

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
		return nil, fmt.Errorf("failed to query table names: %w", err)
	}
	defer logging.SafeCloseWithLogging(rows,
		slog.Default().With(slog.String("component", "debugging")),
		"database_rows")
	var tables []string

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
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
