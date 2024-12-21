package cryptoUtils

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"

	"golang.org/x/exp/rand"
)

// Helper function to join column names or placeholders
func columnsCommaSeparated(columns []string) string {
	return fmt.Sprintf("%s", strings.Join(columns, ", "))
}

func ShuffleRows(dbFile string, seed int) error {
	// Open the SQLite database connection
	dsn := fmt.Sprintf("%s", dbFile) // SQLite file path
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Prepare the query with the custom seed
	query := fmt.Sprintf(`
		WITH randomized_nodes AS (
			SELECT
				rowid,
				"computer_id",
				"ip_address",
				"node_group",
				ROW_NUMBER() OVER (ORDER BY RANDOM() %% %d) AS new_sort_order
			FROM "nodes"
		)
		UPDATE "nodes"
		SET "sort_order" = (
			SELECT new_sort_order
			FROM randomized_nodes
			WHERE randomized_nodes.rowid = "nodes".rowid
		);
		`, seed)

	// Execute the query
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %v", err)
	}

	return nil
}

func AssignGroupNumbers(dbFile string, groupSize int) error {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Begin a transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback() // Ensure rollback in case of an error

	// Create the SQL query with parameterized group size
	query := `
			WITH RankedNodes AS (
				SELECT 
					rowid,  -- Use the rowid to update the rows
					sort_order,
					(ROW_NUMBER() OVER (ORDER BY sort_order) - 1) / ? + 1 AS new_node_group
				FROM nodes
			)
			UPDATE nodes
			SET node_group = (SELECT new_node_group FROM RankedNodes WHERE RankedNodes.rowid = nodes.rowid);
		`

	// Execute the query with the provided group size
	_, err = tx.Exec(query, groupSize)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// ClearNodeGroupColumn clears the node_group column in the nodes table.
func ClearNodeGroupColumn(dbPath string) error {
	log.Println("Clearing node_group column...")

	// Connect to the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Println("Failed to open database:", err)
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Clear the node_group column by setting it to NULL
	_, err = db.Exec("UPDATE nodes SET node_group = NULL;")
	if err != nil {
		log.Println("Failed to clear node_group column:", err)
		return fmt.Errorf("failed to clear node_group column: %w", err)
	}

	log.Println("node_group column cleared successfully.")
	return nil
}

// insertRandomData populates the nodes table with 100,000 rows of random data.
func InsertRandomData(dbFile string, AmountToInsert int) {
	rand.Seed(1234) // Seed the random number generator

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO nodes (sort_order, computer_id, ip_address, node_group) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for i := 0; i < AmountToInsert; i++ {
		sortOrder := i + 1
		ipAddress := generateRandomIPAddress()
		port := rand.Intn(65535-1024) + 1024 // Random port between 1024 and 65535
		ipWithPort := fmt.Sprintf("%s:%d", ipAddress, port)
		computerID := generateSHA256(ipWithPort)
		nodeGroup := rand.Intn(10) + 1 // Random node group between 1 and 10

		_, err = stmt.Exec(sortOrder, computerID, ipWithPort, nodeGroup)
		if err != nil {
			log.Fatalf("Failed to insert data: %v", err)
		}
	}

	log.Printf("Inserted %d rows successfully.", AmountToInsert)
}

// generateRandomIPAddress creates a random IPv4 address.
func generateRandomIPAddress() string {
	return net.IPv4(byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256)), byte(rand.Intn(256))).String()
}

// generateSHA256 hashes a string using SHA-256 and returns the hex-encoded result.
func generateSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
