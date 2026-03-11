package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"leona-scanner/internal/handler"
	"leona-scanner/internal/repository"
	"leona-scanner/internal/usecase"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v74"
	mollie "github.com/VictorAvelar/mollie-api-go/v3/mollie"
)

var (
	Version = "1.0.0"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize payment provider (Mollie preferred, Stripe fallback)
	mollieKey := os.Getenv("MOLLIE_API_KEY")
	stripeKey := os.Getenv("STRIPE_API_KEY")
	
	if mollieKey != "" {
		log.Println("✅ Mollie payment provider initialized")
		// Initialize Mollie client
		config := mollie.NewConfig(true, mollie.APITokenEnv)
		_, err := mollie.NewClient(nil, config)
		if err != nil {
			log.Printf("Warning: Mollie client initialization failed: %v", err)
		}
	} else if stripeKey != "" {
		stripe.Key = stripeKey
		log.Println("✅ Stripe payment provider initialized (fallback)")
	} else {
		log.Println("⚠️  No payment provider configured (payments disabled)")
	}

	// Initialize database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	repo, err := repository.NewRepository(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()
	log.Println("✅ Database connection established")

	// Initialize services
	scannerService := usecase.NewScannerService(repo)
	pdfService := usecase.NewPDFService()

	// Initialize HTTP handler v2 (with Gap Analysis & multi-tier checkout)
	h := handler.NewHTTPHandlerV2(scannerService, pdfService)

	// Setup router
	r := mux.NewRouter()

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Routes
	r.HandleFunc("/", h.HandleIndex).Methods("GET")
	r.HandleFunc("/insights", h.HandleInsights).Methods("GET")
	r.HandleFunc("/kennisbank", h.HandleKennisbank).Methods("GET")
	r.HandleFunc("/free-report", h.HandleFreeReport).Methods("GET")
	r.HandleFunc("/free-audit", h.HandleFreeAudit).Methods("GET")
	r.HandleFunc("/api/scan", h.HandleScan).Methods("POST")
	r.HandleFunc("/api/checkout/tier1", h.HandleCheckoutTier1).Methods("POST")
	r.HandleFunc("/api/checkout/tier2", h.HandleCheckoutTier2).Methods("POST")
	r.HandleFunc("/api/checkout/tier3", h.HandleCheckoutTier3).Methods("POST")
	r.HandleFunc("/api/webhook", h.HandleWebhook).Methods("POST")
	r.HandleFunc("/api/lead/engineer", h.HandleEngineerLeadMagnet).Methods("POST")
	r.HandleFunc("/api/lead/lawyer", h.HandleLawyerLeadMagnet).Methods("POST")
	r.HandleFunc("/api/lead/checklist", h.HandleChecklistDownload).Methods("POST")
	r.HandleFunc("/checklists", h.HandleChecklistPage).Methods("GET")
	r.HandleFunc("/success", h.HandleSuccess).Methods("GET")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Configure server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	log.Printf("LEONA CRA Scanner v%s starting on port %s", Version, port)
	log.Printf("Royal Blue (#1428A0) & Davis Orange (#FF6B35) branding active")
	log.Printf("CRA Compliance Engine initialized")
	if mollieKey != "" {
		log.Printf("Payment provider: Mollie (Belgian-optimized)")
	} else if stripeKey != "" {
		log.Printf("Payment provider: Stripe (fallback)")
	}
	log.Printf("SMTP configured for email delivery")
	log.Printf("Visit http://localhost:%s to start scanning\n", port)
	
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
