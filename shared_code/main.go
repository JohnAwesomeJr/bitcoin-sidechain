package main

import (
	"bitcoin-sidechain/cryptoUtils"
	"bitcoin-sidechain/networkUtils"
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	// Load configuration from file
	config, err := LoadConfig("./config.txt")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}
	port := config["PORT"]

	// Front End Pages
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/sendTransaction", sendTransactionHandler)
	http.HandleFunc("/keysGen", keyGenHandler)
	http.HandleFunc("/balance", walletBalance)

	// API Endpoints ------
	http.HandleFunc("/addNodeRequest", addNodeRequest)
	http.HandleFunc("/ping", pingHandler)

	// Work In Progress
	http.HandleFunc("/walletbalance", checkWalletBalance)
	http.HandleFunc("/verifysignature", VerifySignatureHandler)
	http.HandleFunc("/makewallet", insertNewWallet)
	http.HandleFunc("/talkToOtherServer", TalkToOtherServers)
	http.HandleFunc("/database", serveDatabaseHandler("nodes.db"))
	http.HandleFunc("/downloadData", fileDownloadHandler)
	http.HandleFunc("/syncNodeList", syncNodeList)
	http.HandleFunc("/queData", queData)

	// internal node functions (not for use as an API endpoint)
	http.HandleFunc("/shuffleDatabase", shuffleDatabase)
	http.HandleFunc("/dummy", addDummyNodes)
	http.HandleFunc("/hashData", hashDatabaseHandler)

	ip := "0.0.0.0"
	address := fmt.Sprintf("%s:%s", ip, port)

	if err := http.ListenAndServe(address, nil); err != nil {
		panic(err)
	}

	time.Sleep(2 * time.Second)
	fmt.Println("____________Hi All!___________")

}

// Copy to a function where you want to simulate a delay.
// Generate a random duration between 500 ms and 1 second
// duration := time.Duration(rand.Intn(501)+2000) * time.Millisecond
// fmt.Printf("Simulating latency of %v...\n", duration)
// time.Sleep(duration)
// fmt.Printf("Completed after simulating latency of %v\n", duration)

// LoadConfig reads a configuration file and returns a map of key-value pairs
func LoadConfig(filename string) (map[string]string, error) {
	config := make(map[string]string)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		config[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func FetchJSON(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// Front End Pages
func rootHandler(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "pages/index.html")
}

func sendTransactionHandler(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "pages/signTransaction.html")
}

func keyGenHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "pages/keyGen.html")
}

func walletBalance(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "pages/walletBalance.html")
}

// API Endpoints
func VerifySignatureHandler(w http.ResponseWriter, r *http.Request) {
	type Transaction struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount string `json:"amount"`
		Nonce  string `json:"nonce"`
	}

	type KeySignRequest struct {
		Signature   string      `json:"signature"`
		Transaction Transaction `json:"transaction"`
	}

	// Decode the incoming JSON request
	var req KeySignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Print error to the console
		fmt.Println("Invalid request payload:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // Close the request body

	// Convert the Transaction back to a JSON string
	transactionJSON, err := json.Marshal(req.Transaction)
	if err != nil {
		// Print error to the console
		fmt.Println("Error converting transaction to JSON:", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Assuming cryptoUtils.VerifySignature is your verification function
	publicKey := req.Transaction.From
	toAddress := req.Transaction.To
	amount := req.Transaction.Amount
	nonce := req.Transaction.Nonce

	signature := req.Signature
	message := string(transactionJSON) // The JSON string of the transaction

	// Check if the signature is valid
	verificationResult, err := cryptoUtils.VerifySignature(signature, publicKey, message)
	if err != nil {
		// Print error to the console
		fmt.Println("Error verifying signature:", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Make wallet if it doesn't exist (convert to use MySQL)
	databaseFile := "node-1-database" // Update to use MySQL (use the DSN from your Docker config)
	nonceValidation, nonceError := cryptoUtils.CheckNonce(databaseFile, nonce)
	if nonceError != nil {
		// Print error to the console
		fmt.Println("Error checking nonce:", nonceError)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Convert to use MySQL for NewWallet
	_, walletError := cryptoUtils.NewWallet(toAddress, databaseFile)
	if walletError != nil {
		// Print error to the console
		fmt.Println("Error creating wallet:", walletError)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Convert to use MySQL for MoveSats
	moveError := cryptoUtils.MoveSats(publicKey, toAddress, amount, databaseFile)
	if moveError != nil {
		// Print error to the console
		fmt.Println("Error moving sats:", moveError)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Create the response object
	response := map[string]string{}
	if verificationResult == "valid" && !nonceValidation {
		response["message"] = "Valid"
	} else {
		// Log for debugging
		fmt.Println("Verification Result:", verificationResult)
		fmt.Println("Nonce already used:", nonceValidation, "nonceError?", nonceError)
		response["message"] = "Invalid"
	}

	// Set the content type and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Print error to the console
		fmt.Println("Error encoding response:", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}
}

func checkWalletBalance(w http.ResponseWriter, r *http.Request) {
	// WalletBalanceRequest is a struct to parse the incoming JSON request
	type WalletBalanceRequest struct {
		Wallet string `json:"wallet"`
	}

	// WalletBalanceResponse is a struct to form the JSON response
	type WalletBalanceResponse struct {
		Status  string  `json:"status"`
		Balance float64 `json:"balance,omitempty"`
		Message string  `json:"message,omitempty"`
	}
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Parse the JSON request
	var req WalletBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", "nodeList1.db")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		log.Println("Database connection error:", err)
		return
	}
	defer db.Close()

	// Query the wallet balance
	var balance float64
	err = db.QueryRow("SELECT balance FROM wallet_balances WHERE wallet = ?", req.Wallet).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			// Wallet not found
			response := WalletBalanceResponse{
				Status:  "error",
				Message: "Wallet not found",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		// Internal error
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		log.Println("Query error:", err)
		return
	}

	// Wallet found, send the balance
	response := WalletBalanceResponse{
		Status:  "success",
		Balance: balance,
	}
	json.NewEncoder(w).Encode(response)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	// PingResponse represents the structure of the JSON response
	type PingResponse struct {
		LocalIP  string `json:"local_ip"`
		GlobalIP string `json:"global_ip"`
	}

	localIP, err := networkUtils.GetLocalIP()
	if err != nil {
		http.Error(w, "Error getting local IP", http.StatusInternalServerError)
		return
	}

	globalIP, err := networkUtils.GetGlobalIP()
	if err != nil {
		http.Error(w, "Error getting global IP", http.StatusInternalServerError)
		return
	}

	response := PingResponse{
		LocalIP:  localIP,
		GlobalIP: globalIP,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Work In Progress
func hashDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	hash := cryptoUtils.ComputeDatabaseHash()
	response := map[string]string{"hash": hash}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func insertNewWallet(w http.ResponseWriter, r *http.Request) {

	databaseFile := "nodeList1.db"
	cryptoUtils.NewWallet("Gasp, can it be!", databaseFile)
}

func TalkToOtherServers(w http.ResponseWriter, r *http.Request) {
	// Wait 5 seconds before fetching JSON
	time.Sleep(5 * time.Second)

	// URL to fetch JSON data from
	url := "http://node-2/ping"
	data, err := FetchJSON(url)
	if err != nil {
		log.Fatalf("Error fetching JSON: %v", err)
	}

	fmt.Println("Received JSON:", data)
}
func serveDatabaseHandler(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if file exists
		if _, err := ioutil.ReadFile(filePath); err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Set headers to initiate file download
		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
		w.Header().Set("Content-Type", "application/octet-stream")

		// Serve the file
		http.ServeFile(w, r, filePath)
	}
}

func downloadFileFromEndpoint(endpoint, directory string) error {
	// Send GET request to the endpoint
	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch file from endpoint: %w", err)
	}
	defer resp.Body.Close()

	// Check if response status is OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	// Get the filename from the "Content-Disposition" header
	contentDisposition := resp.Header.Get("Content-Disposition")
	var filename string
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			filename = params["filename"]
		}
	}
	if filename == "" {
		return fmt.Errorf("could not determine filename from response headers")
	}

	// Create the full path for the file
	filePath := filepath.Join(directory, filename)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("File downloaded successfully to: %s\n", filePath)
	return nil
}

func fileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Example usage
	endpoint := "http://node-1/database" // Replace with the actual endpoint
	directory := "./database_downloads"  // Replace with your desired directory

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	// Download the file from the endpoint
	if err := downloadFileFromEndpoint(endpoint, directory); err != nil {
		fmt.Println("Error downloading file:", err)
	}
}

// internal node functions (not for use as an API endpoint)
func shuffleDatabase(w http.ResponseWriter, r *http.Request) {
	groupSize := 2
	shuffleSeed := 8574848843759384334

	data, _ := cryptoUtils.GetDataFromDatabase()
	shuffledData := cryptoUtils.ShuffleResults(data, int64(shuffleSeed))
	orderedData := cryptoUtils.AssignNewOrderBy(shuffledData)
	groupedData := cryptoUtils.AssignNodeGroups(orderedData, groupSize)
	cryptoUtils.UpdateNodesTable(groupedData)

	// Respond to the client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Shuffle done!")
}

func addDummyNodes(w http.ResponseWriter, r *http.Request) {
	cryptoUtils.InsertRandomData(2)
	// Respond to the client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "insert Dummy Data Done")
}

func addNodeRequest(w http.ResponseWriter, r *http.Request) {

	type IncomingRequest struct {
		IPAddress string `json:"ipaddress"`
	}

	type Response struct {
		Message       string `json:"message"`
		InNodes       bool   `json:"in_nodes"`
		InNodesQue    bool   `json:"in_nodes_que"`
		InNodesBuffer bool   `json:"in_nodes_buffer"`
	}

	type ErrorResponse struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	// Ensure the method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		errorResponse := ErrorResponse{
			Message: "Only POST requests are allowed",
			Code:    http.StatusMethodNotAllowed,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Parse the incoming JSON request
	var incoming IncomingRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse := ErrorResponse{
			Message: "Failed to read request body",
			Code:    http.StatusBadRequest,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	err = json.Unmarshal(body, &incoming)
	if err != nil {
		errorResponse := ErrorResponse{
			Message: "Invalid JSON format",
			Code:    http.StatusBadRequest,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Validate the IPAddress field
	if incoming.IPAddress == "" {
		errorResponse := ErrorResponse{
			Message: "IP address is required",
			Code:    http.StatusBadRequest,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Construct the /ping endpoint URL
	pingURL := fmt.Sprintf("http://%s/ping", incoming.IPAddress)

	// Perform the GET request to the /ping endpoint
	resp, err := http.Get(pingURL)
	if err != nil {
		errorResponse := ErrorResponse{
			Message: fmt.Sprintf("Failed to reach %s: %v", pingURL, err),
			Code:    http.StatusBadGateway,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	defer resp.Body.Close()

	// Database connection setup
	dsn := "node:test@tcp(node-1-database:3306)/node"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		errorResponse := ErrorResponse{
			Message: "Failed to connect to the database",
			Code:    http.StatusInternalServerError,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	defer db.Close()

	// Check the database connection
	err = db.Ping()
	if err != nil {
		errorResponse := ErrorResponse{
			Message: "Failed to ping the database",
			Code:    http.StatusInternalServerError,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Variables to track IP presence
	var inNodes, inNodesQue, inNodesBuffer bool
	var count int

	// Step 1: Check if the IP is in the nodes table
	query := `SELECT COUNT(*) FROM nodes WHERE ip_address = ?`
	err = db.QueryRow(query, incoming.IPAddress).Scan(&count)
	if err != nil {
		errorResponse := ErrorResponse{
			Message: fmt.Sprintf("Error querying the database: %v", err),
			Code:    http.StatusInternalServerError,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	inNodes = count > 0

	// Step 2: If IP is in nodes, return that information
	if inNodes {
		response := Response{
			Message:       "IP is already in the nodes database",
			InNodes:       true,
			InNodesQue:    false,
			InNodesBuffer: false,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Step 3: Check if the IP is in the nodes_que table
	query = `SELECT COUNT(*) FROM nodes_que WHERE ip_address = ?`
	err = db.QueryRow(query, incoming.IPAddress).Scan(&count)
	if err != nil {
		errorResponse := ErrorResponse{
			Message: fmt.Sprintf("Error querying the database: %v", err),
			Code:    http.StatusInternalServerError,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	inNodesQue = count > 0

	// Step 4: If IP is in nodes_que, return that information
	if inNodesQue {
		response := Response{
			Message:       "IP is already in the nodes queue",
			InNodes:       false,
			InNodesQue:    true,
			InNodesBuffer: false,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Step 5: If IP is not in nodes or nodes_que, add it to the nodes_buffer
	insertQuery := `INSERT INTO nodes_buffer (ip_address) VALUES (?)`
	_, err = db.Exec(insertQuery, incoming.IPAddress)
	if err != nil {
		errorResponse := ErrorResponse{
			Message: fmt.Sprintf("Failed to add IP to the buffer: %v", err),
			Code:    http.StatusInternalServerError,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Verify the IP is now in the buffer
	query = `SELECT COUNT(*) FROM nodes_buffer WHERE ip_address = ?`
	err = db.QueryRow(query, incoming.IPAddress).Scan(&count)
	if err != nil {
		errorResponse := ErrorResponse{
			Message: fmt.Sprintf("Error verifying IP in buffer: %v", err),
			Code:    http.StatusInternalServerError,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	inNodesBuffer = count > 0

	// Prepare the response
	response := Response{
		Message:       "IP successfully added to nodes buffer",
		InNodes:       false,
		InNodesQue:    false,
		InNodesBuffer: inNodesBuffer,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func syncNodeList(w http.ResponseWriter, r *http.Request) {
	// Database connection
	db, err := sql.Open("mysql", "node:test@tcp(node-1-database:3306)/node")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Step 1: Move data from nodes_buffer to nodes_que
	_, err = db.Exec("INSERT INTO nodes_que SELECT * FROM nodes_buffer")
	if err != nil {
		http.Error(w, "Failed to move data from buffer to queue", http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("DELETE FROM nodes_buffer")
	if err != nil {
		http.Error(w, "Failed to clear nodes buffer", http.StatusInternalServerError)
		return
	}

	// Step 2: Fetch all nodes from the nodes table
	nodes, err := db.Query("SELECT ip_address, reachable FROM nodes")
	if err != nil {
		http.Error(w, "Failed to fetch nodes", http.StatusInternalServerError)
		return
	}
	defer nodes.Close()

	// Step 3: Ping each node and update reachable status
	for nodes.Next() {
		var id int
		var ipAddress string
		var reachable bool
		if err := nodes.Scan(&id, &ipAddress, &reachable); err != nil {
			continue
		}

		url := fmt.Sprintf("http://%s/ping", ipAddress)
		client := http.Client{Timeout: 7 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			_, _ = db.Exec("UPDATE nodes SET reachable = 0 WHERE id = ?", id)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			_, _ = db.Exec("UPDATE nodes SET reachable = 1 WHERE id = ?", id)
		}
	}

	// Step 4: Fetch reachable nodes
	reachableNodes, err := db.Query("SELECT ip_address FROM nodes WHERE reachable = 1")
	if err != nil {
		http.Error(w, "Failed to fetch reachable nodes", http.StatusInternalServerError)
		return
	}
	defer reachableNodes.Close()

	// Step 5: Request and compare queue data from reachable nodes
	for reachableNodes.Next() {
		var ipAddress string
		reachableNodes.Scan(&ipAddress)
		url := fmt.Sprintf("http://%s/queData", ipAddress)
		resp, err := http.Get(url)
		if err != nil {
			_, _ = db.Exec("UPDATE nodes SET reachable = 0 WHERE ip_address = ?", ipAddress)
			continue
		}
		var remoteQueue []struct {
			ID   int    `json:"id"`
			Data string `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&remoteQueue); err == nil {
			for _, data := range remoteQueue {
				_, _ = db.Exec("INSERT IGNORE INTO nodes_que (id, data) VALUES (?, ?)", data.ID, data.Data)
			}
		}
		resp.Body.Close()
	}

	// Step 6: Remove duplicate data from nodes_que
	// _, err = db.Exec("DELETE t1 FROM nodes_que t1 INNER JOIN nodes_que t2 WHERE t1.id > t2.id AND t1.data = t2.data")
	// if err != nil {
	// 	http.Error(w, "Failed to remove duplicate data", http.StatusInternalServerError)
	// 	return
	// }

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sync complete"))
}

func queData(w http.ResponseWriter, r *http.Request) {
	// Database connection
	db, err := sql.Open("mysql", "node:test@tcp(node-1-database:3306)/node")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Fetch data from nodes_que table
	rows, err := db.Query("SELECT ip_address FROM nodes_que")
	if err != nil {
		http.Error(w, "Failed to fetch queue data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var queue []struct {
		Data string `json:"ip_address"`
	}
	for rows.Next() {
		var data struct {
			Data string `json:"ip_address"`
		}
		if err := rows.Scan(&data.Data); err == nil {
			queue = append(queue, data)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(queue)
}
