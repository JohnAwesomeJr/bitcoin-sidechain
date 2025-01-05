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

	// Work In Progress
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/walletbalance", checkWalletBalance)
	http.HandleFunc("/verifysignature", VerifySignatureHandler)
	http.HandleFunc("/makewallet", insertNewWallet)
	http.HandleFunc("/talkToOtherServer", TalkToOtherServers)
	http.HandleFunc("/database", serveDatabaseHandler("nodes.db"))
	http.HandleFunc("/downloadData", fileDownloadHandler)

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
	cryptoUtils.InsertRandomData(10)
	// Respond to the client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "insert Dummy Data Done")
}
