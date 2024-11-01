package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func hashing() {
	// Check if a filename is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run db_to_csv.go <database_file.db>")
		return
	}

	// Get the database filename from command-line arguments
	dbFilename := os.Args[1]
	csvFilename := "output.csv"

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbFilename)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// Query all table names
	tables, err := getTableNames(db)
	if err != nil {
		fmt.Println("Error fetching table names:", err)
		return
	}

	// Create the CSV file
	csvFile, err := os.Create(csvFilename)
	if err != nil {
		fmt.Println("Error creating CSV file:", err)
		return
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)

	// Write data from each table to CSV
	for _, table := range tables {
		if err := writeTableToCSV(db, writer, table); err != nil {
			fmt.Println("Error writing table to CSV:", err)
			return
		}
	}
	writer.Flush()

	// Compute the SHA-256 hash of the CSV file
	hash, err := computeFileHash(csvFilename)
	if err != nil {
		fmt.Println("Error computing file hash:", err)
		return
	}
	fmt.Printf("SHA-256 hash of %s: %s\n", csvFilename, hash)

	// Delete the CSV file
	if err := os.Remove(csvFilename); err != nil {
		fmt.Println("Error deleting CSV file:", err)
		return
	}
	fmt.Printf("Deleted file: %s\n", csvFilename)
}

// getTableNames returns a list of table names in the database
func getTableNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

// writeTableToCSV writes the contents of a table to a CSV file
func writeTableToCSV(db *sql.DB, writer *csv.Writer, tableName string) error {
	// Query table rows
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	writer.Write(columns)

	// Write rows to CSV
	for rows.Next() {
		columns := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return err
		}

		rowData := make([]string, len(columns))
		for i, col := range columns {
			if col != nil {
				rowData[i] = fmt.Sprintf("%v", col)
			}
		}
		writer.Write(rowData)
	}

	return nil
}

// computeFileHash computes the SHA-256 hash of a file
func computeFileHash(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}