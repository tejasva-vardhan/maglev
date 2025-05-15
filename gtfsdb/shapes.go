package gtfsdb

import (
	"database/sql"
	"fmt"
)

// Shape represents points that define a vehicle's path
type Shape struct {
	ID       string  // shape_id
	Lat      float64 // shape_pt_lat
	Lon      float64 // shape_pt_lon
	Sequence int     // shape_pt_sequence
}

// InsertShapes inserts multiple stop times using a transaction for better performance
func InsertShapes(db *sql.DB, shapes []Shape) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO shapes (
			shape_id, lat, lon, shape_pt_sequence
		) VALUES (?, ?, ?, ?);
	`)
	if err != nil {
		tx.Rollback() // nolint:errcheck
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close() // nolint:errcheck

	for _, st := range shapes {
		_, err := stmt.Exec(
			st.ID, st.Lat, st.Lon, st.Sequence,
		)
		if err != nil {
			tx.Rollback() // nolint:errcheck
			return fmt.Errorf("error inserting shapes: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func createShapesTable(tx *sql.Tx) {
	createTable(tx, "shapes", `
		CREATE TABLE IF NOT EXISTS shapes (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    shape_id TEXT NOT NULL,
		    lat REAL NOT NULL,
		    lon REAL NOT NULL,
		    shape_pt_sequence INTEGER NOT NULL
		);
    `)
}
