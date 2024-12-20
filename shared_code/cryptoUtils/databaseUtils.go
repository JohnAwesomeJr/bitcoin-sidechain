package cryptoUtils

import (
	"database/sql"
	"fmt"
	"log"
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

	// Retrieve all rows from the table, ordered by the current sort_order
	query := fmt.Sprintf("SELECT rowid, * FROM %s ORDER BY sort_order ASC", tableName)
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

	// Seed the random number generator for deterministic shuffling
	rand.Seed(uint64(seed))

	// Shuffle the rows deterministically
	for i := len(allRows) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		allRows[i], allRows[j] = allRows[j], allRows[i]
	}

	// Begin a transaction to update the sort_order column with unique values
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	// Prepare an update statement to modify sort_order
	updateQuery := fmt.Sprintf("UPDATE %s SET sort_order = ? WHERE rowid = ?", tableName)
	stmt, err := tx.Prepare(updateQuery)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer stmt.Close()

	// Assign new sort_order values, starting from 1 and incrementing
	for i, row := range allRows {
		// We are using the index `i` to assign sort_order, starting from 1
		rowID := row["rowid"]
		if rowID == nil {
			tx.Rollback()
			return nil, fmt.Errorf("row does not contain 'rowid'")
		}

		// Update the sort_order with the new value (starting from 1)
		_, err := stmt.Exec(i+1, rowID) // i+1 to start sort_order from 1
		if err != nil {
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

func AssignGroupNumbersToNodes(dbPath string, groupSize int) error {
	// Validate groupSize
	if groupSize <= 0 {
		return fmt.Errorf("group size must be greater than 0")
	}

	// Connect to the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create a temporary table with row numbers based on sort_order
	_, err = db.Exec(`
		CREATE TEMP TABLE NumberedRows AS
		SELECT 
			ROW_NUMBER() OVER (ORDER BY sort_order) AS row_num, 
			computer_id
		FROM nodes;
	`)
	if err != nil {
		return fmt.Errorf("failed to create temporary table: %w", err)
	}

	// Update the node_group column using a prepared statement
	stmt, err := db.Prepare(`
		UPDATE nodes
		SET node_group = (
			SELECT ((row_num - 1) / ?) + 1
			FROM NumberedRows
			WHERE NumberedRows.computer_id = nodes.computer_id
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(groupSize)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	// Drop the temporary table
	_, err = db.Exec("DROP TABLE NumberedRows;")
	if err != nil {
		return fmt.Errorf("failed to drop temporary table: %w", err)
	}

	log.Println("Group assignment completed successfully.")
	return nil
}
