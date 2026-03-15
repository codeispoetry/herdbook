package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ValueRequest struct {
	Category string `json:"category"`
	Message  string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	// Initialize database
	initDB()

	// Generate certificates if they don't exist
	if err := generateCerts(); err != nil {
		log.Fatal("Failed to generate certificates:", err)
	}

	// Setup routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/post", handlePost)
	http.HandleFunc("/list", handleList)

	// Start HTTP server for local development
	fmt.Println("Herdbook API server starting on :9002 (HTTP)...")
	fmt.Println("Open http://localhost:9002/ in your browser")
	log.Fatal(http.ListenAndServe(":9002", nil))
	//log.Fatal(http.ListenAndServeLTS(":9002", nil))

}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Method not allowed. Use POST to echo a value.",
		})
		return
	}

	var req ValueRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Invalid JSON format",
		})
		return
	}

	db, err := sql.Open("sqlite3", "herdbook.db")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Database error"})
		return
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO entries (timestamp, category, message) VALUES (CURRENT_TIMESTAMP, ?, ?)", req.Category, req.Message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to save message"})
		log.Println("Failed to save message:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValueRequest{Category: req.Category, Message: req.Message})
}

func handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Method not allowed. Use GET to list values.",
		})
		return
	}

	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", "herdbook.db")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Database error"})
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM entries ORDER BY timestamp DESC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to query entries"})
		fmt.Println("Failed to query entries:", err)
		return
	}
	defer rows.Close()

	type Entry struct {
		Id        int    `json:"id"`
		Timestamp string `json:"timestamp"`
		Category  string `json:"category"`
		Message   string `json:"message"`
	}

	var entries []Entry
	for rows.Next() {
		var entry Entry
		err := rows.Scan(&entry.Id, &entry.Timestamp, &entry.Category, &entry.Message)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to read entry"})
			fmt.Println("Failed to read entry:", err)
			return
		}
		entries = append(entries, entry)
	}
	defer rows.Close()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entries)
}

func initDB() {
	db, err := sql.Open("sqlite3", "herdbook.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create table if it doesn't exist
	createEntryTableSQL := `CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		category TEXT NOT NULL,
		message TEXT NOT NULL
	);`

	_, err = db.Exec(createEntryTableSQL)
	if err != nil {
		log.Fatal("Failed to create entries table:", err)
	}

	fmt.Println("Database initialized successfully")
}

func generateCerts() error {
	// Check if certificates already exist
	if _, err := os.Stat("server.crt"); err == nil {
		if _, err := os.Stat("server.key"); err == nil {
			fmt.Println("SSL certificates already exist")
			return nil
		}
	}

	fmt.Println("Generating SSL certificates...")

	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Health Tracker"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: nil,
		DNSNames:    []string{"localhost", "tom-rose.de"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// Save certificate
	certOut, err := os.Create("server.crt")
	if err != nil {
		return fmt.Errorf("failed to open cert.pem for writing: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("failed to write certificate: %v", err)
	}

	// Save private key
	keyOut, err := os.Create("server.key")
	if err != nil {
		return fmt.Errorf("failed to open key.pem for writing: %v", err)
	}
	defer keyOut.Close()

	privKeyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privKeyBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %v", err)
	}

	fmt.Println("SSL certificates generated successfully")
	return nil
}
