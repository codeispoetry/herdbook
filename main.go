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
	Message string     `json:"message"`
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

	// Start HTTPS server
	fmt.Println("Herdbook API server starting on :9001 (HTTPS)...")
	fmt.Println("Open https://tom-rose.de/herdbook/ in your browser")
	log.Fatal(http.ListenAndServeTLS(":9001", "server.crt", "server.key", nil))
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

	rows, err := db.Query("SELECT timestamp, weight FROM weights ORDER BY timestamp DESC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to query weights"})
		return
	}
	defer rows.Close()

	type WeightEntry struct {
		Timestamp string  `json:"timestamp"`
		Weight    float32 `json:"weight"`
	}

	type BloodEntry struct {
		Timestamp string `json:"timestamp"`
		Diastolic int    `json:"diastolic"`
		Systolic  int    `json:"systolic"`
	}

	type ListResponse struct {
		Weight []WeightEntry `json:"weight"`
		Blood  []BloodEntry  `json:"blood"`
	}

	var weightEntries []WeightEntry
	for rows.Next() {
		var entry WeightEntry
		err := rows.Scan(&entry.Timestamp, &entry.Weight)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to read weight entry"})
			return
		}
		weightEntries = append(weightEntries, entry)
	}

	rows, err = db.Query("SELECT timestamp, diastolic, systolic FROM blood ORDER BY timestamp DESC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to query blood pressure"})
		return
	}
	defer rows.Close()

	var bloodEntries []BloodEntry
	for rows.Next() {
		var entry BloodEntry
		err := rows.Scan(&entry.Timestamp, &entry.Diastolic, &entry.Systolic)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to read blood pressure entry"})
			return
		}
		bloodEntries = append(bloodEntries, entry)
	}

	response := ListResponse{
		Weight: weightEntries,
		Blood:  bloodEntries,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
