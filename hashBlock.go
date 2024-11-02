package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func computeDatabaseHash(dbFilename string) string {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbFilename)
	if err != nil {
		return "error"
	}
	defer db.Close()

	// Query all table names
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return "error"
	}
	defer rows.Close()

	// Buffer to hold CSV data
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	// Iterate through tables and write their data to buffer
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return "error"
		}

		// Query table rows
		tableRows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			return "error"
		}
		defer tableRows.Close()

		// Get column names and write as header
		columns, err := tableRows.Columns()
		if err != nil {
			return "error"
		}
		writer.Write(columns)

		// Write rows to buffer
		for tableRows.Next() {
			columns := make([]interface{}, len(columns))
			columnPointers := make([]interface{}, len(columns))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}

			if err := tableRows.Scan(columnPointers...); err != nil {
				return "error"
			}

			rowData := make([]string, len(columns))
			for i, col := range columns {
				if col != nil {
					rowData[i] = fmt.Sprintf("%v", col)
				}
			}
			writer.Write(rowData)
		}
	}

	writer.Flush()

	// Compute SHA-256 hash of the buffer content
	hash := sha256.Sum256(buffer.Bytes())
	return hex.EncodeToString(hash[:])
}
