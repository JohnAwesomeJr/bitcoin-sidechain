package cryptoUtils

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/exp/rand"
)

// Helper function to join column names or placeholders
func columnsCommaSeparated(columns []string) string {
	return fmt.Sprintf("%s", strings.Join(columns, ", "))
}

func ShuffleRows(dbPath, tableName string, seed int64) ([]map[string]interface{}, error) {

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Retrieve all rows from the table
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Read all rows into a slice of maps
	var allRows []map[string]interface{}
	for rows.Next() {
		// Create a slice of values and a map to hold each row
		values := make([]interface{}, len(columns))
		rowMap := make(map[string]interface{})
		for i := range values {
			values[i] = new(interface{})
		}

		// Scan the row into the values slice
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}

		// Populate rowMap with column names and values
		for i, colName := range columns {
			rowMap[colName] = *(values[i].(*interface{}))
		}

		// Append the row map to allRows
		allRows = append(allRows, rowMap)
	}

	// Check for any error that occurred during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Seed the random number generator
	rand.Seed(uint64(seed))

	// Shuffle the rows deterministically
	for i := len(allRows) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		allRows[i], allRows[j] = allRows[j], allRows[i]
	}

	// Begin a transaction to update the database
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	// Delete all existing rows in the table
	deleteQuery := fmt.Sprintf("DELETE FROM %s", tableName)
	if _, err := tx.Exec(deleteQuery); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Prepare an insert statement with placeholders
	placeholder := make([]string, len(columns))
	for i := range placeholder {
		placeholder[i] = "?"
	}
	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		columnsCommaSeparated(columns),
		columnsCommaSeparated(placeholder),
	)

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer stmt.Close()

	// Insert the shuffled rows back into the database
	for _, row := range allRows {
		values := make([]interface{}, len(columns))
		for i, col := range columns {
			values[i] = row[col]
		}

		if _, err := stmt.Exec(values...); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return allRows, nil
}
