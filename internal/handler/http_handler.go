package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"leona-scanner/internal/usecase"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/webhook"
)

type HTTPHandler struct {
	scannerService *usecase.ScannerService
	pdfService     *usecase.PDFService
}

func NewHTTPHandler(scannerService *usecase.ScannerService, pdfService *usecase.PDFService) *HTTPHandler {
	return &HTTPHandler{
		scannerService: scannerService,
		pdfService:     pdfService,
	}
}

// HandleIndex serves the landing page
func (h *HTTPHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
		log.Printf("Template execute error: %v", err)
	}
}

// HandleFreeAudit serves the dedicated free audit landing page
func (h *HTTPHandler) HandleFreeAudit(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/free-audit.html")
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
		log.Printf("Template execute error: %v", err)
	}
}

// HandleScan processes SBOM upload and returns analysis results
func (h *HTTPHandler) HandleScan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		http.Error(w, "Bestand te groot (max 10MB)", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Email is verplicht", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("sbom")
	if err != nil {
		http.Error(w, "Geen SBOM bestand geüpload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sbomData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Bestand lezen fout", http.StatusInternalServerError)
		return
	}

	// Analyze SBOM
	result, scan, err := h.scannerService.AnalyzeSBOM(email, sbomData)
	if err != nil {
		log.Printf("Analysis error: %v", err)
		http.Error(w, fmt.Sprintf("Analyse fout: %v", err), http.StatusInternalServerError)
		return
	}

	// Render results template
	tmpl, err := template.ParseFiles("templates/results.html")
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	data := map[string]interface{}{
		"Result": result,
		"ScanID": scan.ID,
		"Email":  email,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execute error: %v", err)
	}
}

// HandleCheckout creates a Stripe checkout session
func (h *HTTPHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ScanID string `json:"scan_id"`
		Email  string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ongeldige request", http.StatusBadRequest)
		return
	}

	if req.ScanID == "" || req.Email == "" {
		http.Error(w, "ScanID en email zijn verplicht", http.StatusBadRequest)
		return
	}

	// Create Stripe checkout session
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card", "ideal"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("eur"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("CRA Compliance Report"),
						Description: stripe.String("Volledige CRA-compliance analyse en PDF rapport"),
					},
					UnitAmount: stripe.Int64(9900), // €99.00
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:    stripe.String(fmt.Sprintf("%s/success?scan_id=%s", os.Getenv("BASE_URL"), req.ScanID)),
		CancelURL:     stripe.String(fmt.Sprintf("%s/", os.Getenv("BASE_URL"))),
		CustomerEmail: stripe.String(req.Email),
	}
	params.AddMetadata("scan_id", req.ScanID)

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Stripe session error: %v", err)
		http.Error(w, "Betaling initialiseren fout", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"url": sess.URL})
}

// HandleWebhook processes Stripe webhook events
func (h *HTTPHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), endpointSecret)
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Handle payment success
	if event.Type == "checkout.session.completed" {
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		scanID := sess.Metadata["scan_id"]
		if scanID == "" {
			log.Printf("No scan_id in session metadata")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Mark scan as paid
		if err := h.scannerService.MarkScanPaid(scanID); err != nil {
			log.Printf("Error marking scan as paid: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Generate and send PDF
		result, scan, err := h.scannerService.GetScanResult(scanID)
		if err != nil {
			log.Printf("Error getting scan result: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		pdfData, err := h.pdfService.GeneratePDF(result, scan.Platform)
		if err != nil {
			log.Printf("Error generating PDF: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		customerEmail := sess.CustomerDetails.Email
		if customerEmail == "" {
			customerEmail = sess.CustomerEmail
		}

		if err := h.pdfService.SendPDF(customerEmail, pdfData, scanID); err != nil {
			log.Printf("Error sending PDF: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully processed payment for scan %s", scanID)
	}

	w.WriteHeader(http.StatusOK)
}

// HandleSuccess shows payment success page
func (h *HTTPHandler) HandleSuccess(w http.ResponseWriter, r *http.Request) {
	scanID := r.URL.Query().Get("scan_id")

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="nl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Betaling Geslaagd - LEONA</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body style="background-color: #0A192F;" class="text-white">
    <div class="min-h-screen flex items-center justify-center px-4">
        <div class="max-w-md w-full text-center">
            <div class="mb-6">
                <svg class="w-20 h-20 mx-auto text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
            </div>
            <h1 class="text-3xl font-bold mb-4">Betaling Geslaagd!</h1>
            <p class="text-gray-300 mb-6">
                Uw CRA compliance rapport wordt gegenereerd en wordt binnen enkele minuten naar uw e-mail verzonden.
            </p>
            <p class="text-sm text-gray-400 mb-8">
                Scan ID: <code class="bg-gray-800 px-2 py-1 rounded">%s</code>
            </p>
            <a href="/" style="background-color: #FF4500;" class="inline-block px-6 py-3 rounded-lg font-semibold hover:bg-orange-600 transition">
                Nieuwe Scan Uitvoeren
            </a>
        </div>
    </div>
</body>
</html>
	`, scanID)
}
