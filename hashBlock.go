package main

import (
    "bytes"
    "crypto/sha256"
    "database/sql"
    "encoding/csv"
    "encoding/hex"
    "fmt"
    _ "modernc.org/sqlite" // Import the pure Go SQLite driver
)

func computeDatabaseHash(dbFilename string) string {
    // Open the SQLite database
    db, err := sql.Open("sqlite", dbFilename)
    if err != nil {
        return fmt.Sprintf("failed to open database: %v", err)
    }
    defer db.Close()

    // Query all table names
    rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
    if err != nil {
        return fmt.Sprintf("failed to query table names: %v", err)
    }
    defer rows.Close()

    // Buffer to hold CSV data
    var buffer bytes.Buffer
    writer := csv.NewWriter(&buffer)

    // Iterate through tables and write their data to buffer
    for rows.Next() {
        var table string
        if err := rows.Scan(&table); err != nil {
            return fmt.Sprintf("failed to scan table name: %v", err)
        }

        // Query table rows
        tableRows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
        if err != nil {
            return fmt.Sprintf("failed to query rows from table %s: %v", table, err)
        }
        defer tableRows.Close()

        // Get column names and write as header
        columns, err := tableRows.Columns()
        if err != nil {
            return fmt.Sprintf("failed to get columns from table %s: %v", table, err)
        }
        writer.Write(columns)

        // Write rows to buffer
        for tableRows.Next() {
            columnPointers := make([]interface{}, len(columns))
            for i := range columnPointers {
                columnPointers[i] = new(interface{}) // Allocate new interface{}
            }

            if err := tableRows.Scan(columnPointers...); err != nil {
                return fmt.Sprintf("failed to scan row from table %s: %v", table, err)
            }

            rowData := make([]string, len(columns))
            for i, col := range columnPointers {
                if col != nil {
                    rowData[i] = fmt.Sprintf("%v", *(col.(*interface{}))) // Dereference the interface
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
