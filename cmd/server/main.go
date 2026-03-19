package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"leona-scanner/internal/handler"
	"leona-scanner/internal/i18n"
	"leona-scanner/internal/middleware"
	"leona-scanner/internal/repository"
	"leona-scanner/internal/services"
	"leona-scanner/internal/usecase"

	mollie "github.com/VictorAvelar/mollie-api-go/v3/mollie"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v74"
)

var (
	Version = "1.0.0"
)

func main() {
	log.Println("")
	log.Println("┌───────────────────────────────────────────────────────────────────┐")
	log.Println("│   🚀 LEONA CRA Scanner v" + Version + " Starting...        │")
	log.Println("│   🔵 Royal Blue (#1428A0) │ 🟠 Davis Orange (#FF6B35)   │")
	log.Println("└───────────────────────────────────────────────────────────────────┘")
	log.Println("")

	// Phase 1: Configuration Loading
	log.Println("📝 [Phase 1/5] Loading Configuration...")
	if err := godotenv.Load(); err != nil {
		log.Println("   ⚠️  .env file not found, using environment variables")
	} else {
		log.Println("   ✅ .env file loaded successfully")
	}
	log.Println("")

	// Phase 2: Payment Provider Initialization
	log.Println("💳 [Phase 2/5] Initializing Payment Providers...")
	mollieKey := os.Getenv("MOLLIE_API_KEY")
	stripeKey := os.Getenv("STRIPE_API_KEY")

	if mollieKey != "" {
		log.Println("   🇳🇱 Mollie API key detected")
		config := mollie.NewConfig(true, mollie.APITokenEnv)
		_, err := mollie.NewClient(nil, config)
		if err != nil {
			log.Printf("   ❌ Mollie client init failed: %v", err)
		} else {
			log.Println("   ✅ Mollie payment provider initialized (Belgian-optimized)")
		}
	} else if stripeKey != "" {
		stripe.Key = stripeKey
		log.Println("   🐳 Stripe API key detected")
		log.Println("   ✅ Stripe payment provider initialized (fallback)")
	} else {
		log.Println("   ⚠️  No payment provider configured (payments disabled)")
	}
	log.Println("")

	// Phase 3: Database Connection
	log.Println("🗄️ [Phase 3/5] Connecting to Database...")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("   ❌ DATABASE_URL environment variable is required")
	}
	log.Println("   🔌 Connecting to SQLite database...")
	repo, err := repository.NewRepository(dbURL)
	if err != nil {
		log.Fatalf("   ❌ Failed to connect to database: %v", err)
	}
	defer repo.Close()
	log.Println("   ✅ Database connection established")
	log.Println("   📊 Schema validated and ready")
	log.Println("")

	// Phase 4: CVE Vulnerability Service
	log.Println("🔒 [Phase 4/5] Initializing CVE Vulnerability Scanner...")
	nvdAPIKey := os.Getenv("NVD_API_KEY")
	if nvdAPIKey != "" {
		log.Printf("   🔑 NVD API key detected: %s...%s", nvdAPIKey[:8], nvdAPIKey[len(nvdAPIKey)-4:])
	}
	cveService := services.NewCVEService(nvdAPIKey)
	if nvdAPIKey != "" {
		log.Println("   ✅ NVD CVE service initialized (50 req/30s with API key)")
		log.Println("   💨 Rate limiter: 50 requests per 30 seconds")
		log.Println("   📋 Cache TTL: 24 hours")
	} else {
		log.Println("   ⚠️  NVD CVE service initialized (5 req/30s - no API key)")
		log.Println("   🐢 Running on free tier - consider getting API key")
	}
	log.Println("")

	// Phase 5: Service Initialization
	log.Println("⚙️  [Phase 5/5] Initializing Core Services...")
	log.Println("   🛠️  Creating scanner service...")
	scannerService := usecase.NewScannerService(repo, cveService)
	log.Println("   ✅ Scanner service ready")
	log.Println("   📝 Creating PDF service...")
	pdfService := usecase.NewPDFService()
	log.Println("   ✅ PDF service ready")

	// Initialize i18n manager
	log.Println("   🌐 Creating i18n manager...")
	i18nManager, err := i18n.New("./locales")
	if err != nil {
		log.Printf("   ⚠️  i18n initialization failed: %v", err)
		log.Println("   📝 Continuing without i18n support")
		i18nManager = nil
	} else {
		log.Printf("   ✅ i18n ready - Default: %s, Supported: %v", i18nManager.DefaultLang(), i18nManager.SupportedLangs())
	}

	// Initialize Mailgun email service
	log.Println("   📧 Creating Mailgun email service...")
	mailgunService := services.NewMailgunService()
	if mailgunService != nil {
		log.Println("   ✅ Mailgun service ready")
	} else {
		log.Println("   ⚠️  Mailgun not configured - emails will not be sent")
	}

	// Initialize PDF handler with dedicated directory
	pdfHandler := handler.NewPDFHandler(scannerService, "./pdf-reports")

	// Initialize HTTP handler v2 (with Gap Analysis & multi-tier checkout)
	h := handler.NewHTTPHandlerV2(scannerService, pdfService, i18nManager, mailgunService)

	// Setup router
	r := mux.NewRouter()

	// Add logging middleware (controlled by LOG_VERBOSE env var)
	r.Use(middleware.LoggingMiddleware)

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Routes - all using base layout via HandlePage (yield technique)
	r.HandleFunc("/", h.HandlePage("index")).Methods("GET")    // Uses templates/pages/index.html
	r.HandleFunc("/demo", h.HandlePage("demo")).Methods("GET") // Demo request page with navbar/footer

	// Page routes using base layout (just add file to templates/pages/)
	r.HandleFunc("/diensten", h.HandlePage("services")).Methods("GET")
	r.HandleFunc("/producten", h.HandlePage("products-simple")).Methods("GET")
	r.HandleFunc("/snapshot", h.HandlePage("snapshot")).Methods("GET")
	r.HandleFunc("/snapshot/success", h.HandlePage("snapshot-success")).Methods("GET")
	r.HandleFunc("/tcf-bundle", h.HandlePage("tcf-bundle")).Methods("GET")
	r.HandleFunc("/insights", h.HandlePage("insights")).Methods("GET")
	r.HandleFunc("/kennisbank", h.HandlePage("kennisbank")).Methods("GET")
	r.HandleFunc("/free-report", h.HandlePage("free-report")).Methods("GET")
	r.HandleFunc("/free-audit", h.HandlePage("free-audit")).Methods("GET")
	r.HandleFunc("/cra", h.HandlePage("cra")).Methods("GET")
	r.HandleFunc("/cra-compliance", h.HandlePage("cra-compliance")).Methods("GET")
	r.HandleFunc("/cra-assessment", h.HandleCRAAssessment).Methods("GET")

	// Navigation pages
	r.HandleFunc("/wie-zijn-we", h.HandlePage("wie-zijn-we")).Methods("GET")
	r.HandleFunc("/hoe-het-werkt", h.HandlePage("hoe-het-werkt")).Methods("GET")
	r.HandleFunc("/kennis", h.HandlePage("kennis")).Methods("GET")
	r.HandleFunc("/contact", h.HandlePage("contact")).Methods("GET")
	r.HandleFunc("/pricing", h.HandlePage("pricing")).Methods("GET")
	r.HandleFunc("/privacy", h.HandlePage("privacy")).Methods("GET")
	r.HandleFunc("/legal-partners", h.HandlePage("legal-partners")).Methods("GET")

	// Contact form submission
	r.HandleFunc("/api/contact/submit", h.HandleContactSubmit).Methods("POST")

	// Demo request submission
	r.HandleFunc("/api/demo/submit", h.HandleDemoSubmit).Methods("POST")

	// Snapshot audit submission
	r.HandleFunc("/api/snapshot/submit", h.HandleSnapshotSubmit).Methods("POST")

	// Mollie webhook for snapshot payments
	r.HandleFunc("/webhook/mollie", h.HandleMollieWebhook).Methods("POST")

	// Easy to add new pages - just create templates/pages/your-page.html
	// r.HandleFunc("/about", h.HandlePage("about")).Methods("GET")
	// r.HandleFunc("/contact", h.HandlePage("contact")).Methods("GET")
	r.HandleFunc("/api/scan", h.HandleScan).Methods("POST")
	r.HandleFunc("/api/checkout/tier1", h.HandleCheckoutTier1).Methods("POST")
	r.HandleFunc("/api/checkout/tier2", h.HandleCheckoutTier2).Methods("POST")
	r.HandleFunc("/api/checkout/tier3", h.HandleCheckoutTier3).Methods("POST")
	r.HandleFunc("/api/webhook", h.HandleWebhook).Methods("POST")
	r.HandleFunc("/api/lead/engineer", h.HandleEngineerLeadMagnet).Methods("POST")
	r.HandleFunc("/api/lead/lawyer", h.HandleLawyerLeadMagnet).Methods("POST")
	r.HandleFunc("/api/lead/checklist", h.HandleChecklistDownload).Methods("POST")
	r.HandleFunc("/api/lead/risk-assessment", h.HandleRiskAssessment).Methods("POST")
	r.HandleFunc("/api/lead/sample-report", h.HandleSampleReportDownload).Methods("POST")
	r.HandleFunc("/api/cra-assessment/submit", h.HandleCRAAssessmentSubmit).Methods("POST")
	r.HandleFunc("/checklists", h.HandleChecklistPage).Methods("GET")
	r.HandleFunc("/success", h.HandleSuccess).Methods("GET")

	// PDF download routes (€499 automated product)
	r.HandleFunc("/api/pdf/download/{scan_id}", pdfHandler.HandleDownloadPDF).Methods("GET")
	r.HandleFunc("/api/pdf/generate/{scan_id}", pdfHandler.HandleGeneratePDF).Methods("POST")
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

	// Server Ready
	log.Println("")
	log.Println("│")
	log.Println("┌───────────────────────────────────────────────────────────────────┐")
	log.Println("│   ✨ SERVER READY - All Systems Operational                     │")
	log.Println("└───────────────────────────────────────────────────────────────────┘")
	log.Println("")
	log.Println("🌐 Server Status:")
	log.Printf("   • Version: v%s\n", Version)
	log.Printf("   • Port: %s\n", port)
	log.Printf("   • Base URL: http://localhost:%s\n", port)
	log.Println("")
	log.Println("🛠️  Active Services:")
	log.Println("   ✓ CRA Compliance Engine")
	log.Println("   ✓ NVD CVE Vulnerability Scanner")
	log.Println("   ✓ SBOM Parser (CycloneDX + SPDX)")
	log.Println("   ✓ PDF Report Generator")
	log.Println("   ✓ SMTP Email Delivery")
	if mollieKey != "" {
		log.Println("   ✓ Mollie Payment Gateway (🇳🇱 Belgian-optimized)")
	} else if stripeKey != "" {
		log.Println("   ✓ Stripe Payment Gateway (Fallback)")
	}
	log.Println("")
	log.Println("🎨 Branding:")
	log.Println("   • 🔵 Royal Blue (#1428A0)")
	log.Println("   • 🟠 Davis Orange (#FF6B35)")
	log.Println("")
	log.Println("📊 Lead Magnets:")
	log.Println("   • Risk Assessment Quiz (Interactive)")
	log.Println("   • Sample TCF Report (42 pages)")
	log.Println("   • SBOM Validator (Real-time CVE)")
	log.Println("")
	log.Println("┌───────────────────────────────────────────────────────────────────┐")
	log.Printf("│   🚀 VISIT: http://localhost:%s                               │\n", port)
	log.Println("└───────────────────────────────────────────────────────────────────┘")
	log.Println("")
	log.Println("📋 Logging Configuration:")
	if os.Getenv("LOG_VERBOSE") == "true" {
		log.Println("   • Mode: VERBOSE (detailed request/response logs)")
		log.Println("   • Shows: IP, User-Agent, Headers, Size, Timing")
	} else {
		log.Println("   • Mode: COMPACT (one-line logs)")
		log.Println("   • Tip: Set LOG_VERBOSE=true for detailed logs")
	}
	log.Println("")
	log.Println("🟢 Server is LIVE - Accepting requests...")
	log.Println("👀 Watching for incoming connections...")
	log.Println("")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
