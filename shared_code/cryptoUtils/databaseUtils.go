package cryptoUtils

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"

	"golang.org/x/exp/rand"
)

// Helper function to join column names or placeholders
func columnsCommaSeparated(columns []string) string {
	return fmt.Sprintf("%s", strings.Join(columns, ", "))
}

func GetDataFromDatabase(dbFile string) ([]map[string]interface{}, error) {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Execute a query to retrieve the data
	rows, err := db.Query("SELECT * FROM nodes ORDER BY computer_id") // Replace with your actual query
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Prepare to collect the result
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interfaces to hold the column values
		values := make([]interface{}, len(columns))
		valuePointers := make([]interface{}, len(columns))

		// Point each pointer to a value
		for i := range values {
			valuePointers[i] = &values[i]
		}

		// Scan the row into the valuePointers
		err := rows.Scan(valuePointers...)
		if err != nil {
			return nil, err
		}

		// Create a map to hold the row data
		rowData := make(map[string]interface{})
		for i, colName := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowData[colName] = string(b)
			} else {
				rowData[colName] = val
			}
		}

		// Append the row data to results
		results = append(results, rowData)
	}

	// Check for errors after the loop
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
func ShuffleResults(results []map[string]interface{}, seed int64) []map[string]interface{} {
	// Create a new random source with the provided seed
	rand.Seed(uint64(seed))

	// Make a copy of the results slice to preserve the original
	shuffled := make([]map[string]interface{}, len(results))
	copy(shuffled, results)

	// Shuffle the slice in place
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled
}
func AssignNewOrderBy(results []map[string]interface{}) []map[string]interface{} {
	// Overwrite the 'order_by' column with new numbers (1, 2, 3, ...)
	for i := range results {
		results[i]["order_by"] = i + 1
	}
	return results
}
func AssignNodeGroups(results []map[string]interface{}, groupSize int) []map[string]interface{} {
	// Assign groups based on group size (e.g., first 10 rows are group 1, next 10 are group 2, etc.)
	for i := range results {
		group := (i / groupSize) + 1 // Calculate the group number
		results[i]["node_group"] = group
	}
	return results
}
func UpdateNodesTable(dbFile string, results []map[string]interface{}) error {
	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	// Clear the existing data in the table
	_, err = db.Exec("DELETE FROM nodes")
	if err != nil {
		return err
	}

	// Prepare the SQL statement for inserting data
	stmt, err := db.Prepare("INSERT INTO nodes (sort_order, computer_id, ip_address, node_group) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Loop through the results and insert each row into the nodes table
	for _, row := range results {
		sortOrder := row["order_by"].(int)
		computerID := row["computer_id"].(string)
		ipAddress := row["ip_address"].(string)
		nodeGroup := row["node_group"].(int)

		// Execute the insert statement for each row
		_, err := stmt.Exec(sortOrder, computerID, ipAddress, nodeGroup)
		if err != nil {
			return err
		}
	}

	return nil
}

func ShuffleRows(dbFile string, seed int64) {
	type Node struct {
		SortOrder  int
		ComputerID string
		IPAddress  string
		NodeGroup  int
	}
	// Connect to SQLite database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Fetch all rows from the 'nodes' table
	rows, err := db.Query("SELECT sort_order, computer_id, ip_address, node_group FROM nodes")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Create a slice to hold the nodes
	var nodes []Node
	for rows.Next() {
		var node Node
		if err := rows.Scan(&node.SortOrder, &node.ComputerID, &node.IPAddress, &node.NodeGroup); err != nil {
			log.Fatal(err)
		}
		nodes = append(nodes, node)
	}

	// Check for any row iteration errors
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Set the seed for deterministic randomness
	rand.Seed(uint64(seed))

	// Custom sorting function based on IP address and seed
	sort.SliceStable(nodes, func(i, j int) bool {
		// Combine the IP address with the seed to get a consistent result
		rand.Seed(uint64(seed + int64(i))) // Use a different seed for each comparison (by index)
		return rand.Intn(2) == 0           // Simulate randomization based on IP
	})

	// Print the sorted nodes for debugging purposes
	fmt.Println("Sorted nodes based on IP address and seed:")
	for _, node := range nodes {
		fmt.Printf("SortOrder: %d, ComputerID: %s, IPAddress: %s, NodeGroup: %d\n",
			node.SortOrder, node.ComputerID, node.IPAddress, node.NodeGroup)
	}

	// Update the order back into the database based on the new sorted order
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	for i, node := range nodes {
		_, err := tx.Exec("UPDATE nodes SET sort_order = ? WHERE computer_id = ?", i+1, node.ComputerID)
		if err != nil {
			log.Fatal(err)
			tx.Rollback()
			return
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	// Confirmation message
	fmt.Println("Database updated with the new sorted order.")
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
