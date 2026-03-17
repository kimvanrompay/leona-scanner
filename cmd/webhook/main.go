package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"leona-scanner/internal/database"

	"github.com/VictorAvelar/mollie-api-go/v3/mollie"
	"github.com/google/uuid"
)

var (
	db           *database.SupabaseClient
	mollieClient *mollie.Client
)

func init() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")
	mollieAPIKey := os.Getenv("MOLLIE_API_KEY")

	if supabaseURL == "" || supabaseKey == "" || mollieAPIKey == "" {
		log.Fatal("Missing required environment variables: SUPABASE_URL, SUPABASE_SERVICE_KEY, MOLLIE_API_KEY")
	}

	db = database.NewSupabaseClient(supabaseURL, supabaseKey)

	// Initialize Mollie client
	config := mollie.NewConfig(true, mollie.APITokenEnv)
	client, err := mollie.NewClient(nil, config)
	if err != nil {
		log.Fatalf("Failed to initialize Mollie client: %v", err)
	}
	mollieClient = client
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	paymentID := r.FormValue("id")
	if paymentID == "" {
		log.Printf("Missing payment ID in webhook")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Fetch payment status from Mollie
	ctx := context.Background()
	_, payment, err := mollieClient.Payments.Get(ctx, paymentID, nil)
	if err != nil {
		log.Printf("Failed to fetch payment from Mollie: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	log.Printf("Webhook received for payment %s, status: %s", paymentID, payment.Status)

	// Update payment status in database
	var paidAt *time.Time
	if payment.Status == "paid" && payment.PaidAt != nil {
		paidAt = payment.PaidAt
	}

	if err := db.UpdatePaymentStatus(ctx, paymentID, string(payment.Status), paidAt); err != nil {
		log.Printf("Failed to update payment status: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// If payment is successful, unlock the scan
	if payment.Status == "paid" {
		unlockScanFromPayment(ctx, payment)
	}

	w.WriteHeader(http.StatusOK)
}

func unlockScanFromPayment(ctx context.Context, payment *mollie.Payment) {
	// Extract scan_id from metadata
	metadata, ok := payment.Metadata.(map[string]interface{})
	if !ok {
		return
	}

	scanIDStr, ok := metadata["scan_id"].(string)
	if !ok {
		return
	}

	scanID, err := uuid.Parse(scanIDStr)
	if err != nil {
		log.Printf("Invalid scan_id in payment metadata: %v", err)
		return
	}

	if err := db.UnlockScan(ctx, scanID); err != nil {
		log.Printf("Failed to unlock scan: %v", err)
		return
	}

	log.Printf("Successfully unlocked scan %s", scanID)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/webhook/mollie", webhookHandler)
	http.HandleFunc("/health", healthHandler)

	log.Printf("Webhook server starting on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
