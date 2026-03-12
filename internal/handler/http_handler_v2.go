package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"leona-scanner/internal/scanner"
	"leona-scanner/internal/usecase"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/webhook"
)

// GapAnalysisItem represents a single CRA article compliance check
type GapAnalysisItem struct {
	Article     string // e.g. "10.4", "14.1"
	Requirement string // Dutch description of CRA requirement
	Status      string // COMPLIANT, PARTIAL, NON_COMPLIANT
	Finding     string // Specific finding for this scan
}

// CriticalFinding represents a blocker issue
type CriticalFinding struct {
	Title       string
	Description string
	CRAArticle  string
}

// Stats contains summary counts
type Stats struct {
	NonCompliant int
	Partial      int
	Compliant    int
}

type HTTPHandlerV2 struct {
	scannerService *usecase.ScannerService
	pdfService     *usecase.PDFService
}

func NewHTTPHandlerV2(scannerService *usecase.ScannerService, pdfService *usecase.PDFService) *HTTPHandlerV2 {
	return &HTTPHandlerV2{
		scannerService: scannerService,
		pdfService:     pdfService,
	}
}

// HandleIndex serves the landing page
func (h *HTTPHandlerV2) HandleIndex(w http.ResponseWriter, r *http.Request) {
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

// HandleDemo serves the demo request page
func (h *HTTPHandlerV2) HandleDemo(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/demo.html")
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

// HandleServices serves the enterprise services page
func (h *HTTPHandlerV2) HandleServices(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/services.html")
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

// HandleInsights serves the blog/insights index page
func (h *HTTPHandlerV2) HandleInsights(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/insights.html")
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

// HandleKennisbank serves the knowledge base index page
func (h *HTTPHandlerV2) HandleKennisbank(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/kennisbank.html")
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

// HandleFreeReport serves the free CRA readiness report landing page
func (h *HTTPHandlerV2) HandleFreeReport(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/free-report.html")
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
func (h *HTTPHandlerV2) HandleFreeAudit(w http.ResponseWriter, r *http.Request) {
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

// HandleScan processes SBOM upload and returns HTMX partial with Gap Analysis
func (h *HTTPHandlerV2) HandleScan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		http.Error(w, "Bestand te groot (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("sbom")
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

	// Analyze SBOM (no email required for free audit)
	email := "anonymous@scan.local" // Default for free scans
	result, scan, err := h.scannerService.AnalyzeSBOM(email, sbomData)
	if err != nil {
		log.Printf("Analysis error: %v", err)
		http.Error(w, fmt.Sprintf("Analyse fout: %v", err), http.StatusInternalServerError)
		return
	}

	// Transform analysis into Gap Analysis format
	gapAnalysis := h.buildGapAnalysis(result)
	criticalFindings := h.extractCriticalFindings(result)
	stats := h.calculateStats(gapAnalysis)

	// Detect SBOM format
	sbomFormat := "CycloneDX" // Default
	if len(header.Filename) > 0 {
		if strings.HasSuffix(header.Filename, ".spdx") || strings.Contains(string(sbomData), "spdxVersion") {
			sbomFormat = "SPDX"
		}
	}

	// Render HTMX partial
	tmpl, err := template.ParseFiles("templates/partials/scan-results.html")
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	data := map[string]interface{}{
		"ScanID":           scan.ID,
		"ComplianceScore":  result.OverallScore,
		"SBOMFormat":       sbomFormat,
		"TotalComponents":  result.TotalComponents,
		"KernelVersion":    h.detectKernelVersion(result),
		"GapAnalysis":      gapAnalysis,
		"CriticalFindings": criticalFindings,
		"Stats":            stats,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execute error: %v", err)
	}
}

// buildGapAnalysis transforms AnalysisResult into Gap Analysis items
func (h *HTTPHandlerV2) buildGapAnalysis(result *scanner.AnalysisResult) []GapAnalysisItem {
	var items []GapAnalysisItem

	// Map issues to CRA articles
	// This is a simplified mapping - extend with real logic
	
	// Article 10.4 - Security Updates Support Period
	hasKernelEOL := false
	for _, issue := range result.Issues {
		if issue.Category == "SECURITY" && strings.Contains(issue.Description, "kernel") {
			hasKernelEOL = true
			break
		}
	}
	
	kernelStatus := "COMPLIANT"
	kernelFinding := "Linux kernel is binnen support periode"
	if hasKernelEOL {
		kernelStatus = "NON_COMPLIANT"
		kernelFinding = "Gedetecteerde kernel versie nadert EOL of is EOL - patches vereist gedurende productslevensduur"
	}

	items = append(items, GapAnalysisItem{
		Article:     "10.4",
		Requirement: "Security updates gedurende volledige levenscyclus",
		Status:      kernelStatus,
		Finding:     kernelFinding,
	})

	// Article 14 - SBOM Generation & Documentation
	hasTraceabilityIssues := false
	for _, issue := range result.Issues {
		if issue.Category == "TRACEABILITY" {
			hasTraceabilityIssues = true
			break
		}
	}

	sbomStatus := "PARTIAL"
	sbomFinding := fmt.Sprintf("SBOM aanwezig met %d componenten, maar %d componenten missen CPE/PURL traceability", 
		result.TotalComponents, result.HighCount)
	if !hasTraceabilityIssues {
		sbomStatus = "COMPLIANT"
		sbomFinding = "Alle componenten hebben volledige traceability (CPE/PURL)"
	}

	items = append(items, GapAnalysisItem{
		Article:     "14.1",
		Requirement: "Software Bill of Materials met component traceability",
		Status:      sbomStatus,
		Finding:     sbomFinding,
	})

	// Article 14.2 - License Compliance & IP Risk
	hasLicenseIssues := false
	hasGPLRisk := false
	for _, issue := range result.Issues {
		if issue.Category == "IP_RISK" {
			hasLicenseIssues = true
			if strings.Contains(issue.Description, "GPL") {
				hasGPLRisk = true
			}
		}
	}

	licenseStatus := "COMPLIANT"
	licenseFinding := "Alle licenties gedocumenteerd, geen copyleft risico"
	if hasGPLRisk {
		licenseStatus = "NON_COMPLIANT"
		licenseFinding = "GPL-3.0/AGPL componenten gedetecteerd - copyleft verplichtingen van toepassing"
	} else if hasLicenseIssues {
		licenseStatus = "PARTIAL"
		licenseFinding = "Sommige componenten missen licentie-informatie"
	}

	items = append(items, GapAnalysisItem{
		Article:     "14.2",
		Requirement: "Licentie disclosure & IP compliance",
		Status:      licenseStatus,
		Finding:     licenseFinding,
	})

	// Annex I Part II - Secure by Default
	secureByDefaultStatus := "PARTIAL"
	secureByDefaultFinding := "Analyse vereist of onnodige netwerk services zijn uitgeschakeld (BusyBox telnet/FTP check)"

	items = append(items, GapAnalysisItem{
		Article:     "Annex I.2",
		Requirement: "Secure by default configuratie",
		Status:      secureByDefaultStatus,
		Finding:     secureByDefaultFinding,
	})

	// Annex I Part I - Vulnerability Management
	vulnStatus := "PARTIAL"
	vulnFinding := fmt.Sprintf("%d componenten vereisen CVE analyse tegen NVD database", result.TotalComponents)
	if result.OverallScore >= 90 {
		vulnStatus = "COMPLIANT"
		vulnFinding = "Geen kritieke kwetsbaarheden gedetecteerd"
	}

	items = append(items, GapAnalysisItem{
		Article:     "Annex I.1",
		Requirement: "Geen bekende exploiteerbare kwetsbaarheden",
		Status:      vulnStatus,
		Finding:     vulnFinding,
	})

	return items
}

// extractCriticalFindings pulls out blocking issues
func (h *HTTPHandlerV2) extractCriticalFindings(result *scanner.AnalysisResult) []CriticalFinding {
	var findings []CriticalFinding

	for _, issue := range result.Issues {
		if issue.Severity == "CRITICAL" {
			findings = append(findings, CriticalFinding{
				Title:       issue.Component,
				Description: issue.Description,
				CRAArticle:  "10.4", // Map to relevant article - this is simplified
			})
		}
	}

	return findings
}

// calculateStats counts statuses
func (h *HTTPHandlerV2) calculateStats(items []GapAnalysisItem) Stats {
	stats := Stats{}
	for _, item := range items {
		switch item.Status {
		case "COMPLIANT":
			stats.Compliant++
		case "PARTIAL":
			stats.Partial++
		case "NON_COMPLIANT":
			stats.NonCompliant++
		}
	}
	return stats
}

// detectKernelVersion attempts to extract kernel version from results
func (h *HTTPHandlerV2) detectKernelVersion(result *scanner.AnalysisResult) string {
	for _, issue := range result.Issues {
		if strings.Contains(issue.Component, "linux") || strings.Contains(issue.Component, "kernel") {
			// Try to extract version from component name
			return issue.Component
		}
	}
	return "Unknown (geen Linux kernel gedetecteerd)"
}

// HandleCheckoutTier1 creates Stripe session for €499 Audit Start
func (h *HTTPHandlerV2) HandleCheckoutTier1(w http.ResponseWriter, r *http.Request) {
	h.handleCheckout(w, r, "tier1", 49900, "The Audit Start", "1x PDF Rapport (40+ pagina's) + CRA Annex I mapping")
}

// HandleCheckoutTier2 creates Stripe session for €2,450 Compliance Shield
func (h *HTTPHandlerV2) HandleCheckoutTier2(w http.ResponseWriter, r *http.Request) {
	h.handleCheckout(w, r, "tier2", 245000, "The Compliance Shield", "Unlimited scans + CRAVIT-VERIFIED certificaat + Email support")
}

// HandleCheckoutTier3 creates Stripe session for €4,900 Enterprise Partner
func (h *HTTPHandlerV2) HandleCheckoutTier3(w http.ResponseWriter, r *http.Request) {
	h.handleCheckout(w, r, "tier3", 490000, "The Enterprise Partner", "API toegang + CI/CD integratie + On-site audit + Priority support")
}

// handleCheckout is the unified checkout handler
func (h *HTTPHandlerV2) handleCheckout(w http.ResponseWriter, r *http.Request, tier string, amount int64, name, description string) {
	var req struct {
		ScanID string `json:"scan_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Ongeldige request", http.StatusBadRequest)
		return
	}

	if req.ScanID == "" {
		http.Error(w, "ScanID is verplicht", http.StatusBadRequest)
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
						Name:        stripe.String(name),
						Description: stripe.String(description),
					},
					UnitAmount: stripe.Int64(amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(fmt.Sprintf("%s/success?scan_id=%s&tier=%s", os.Getenv("BASE_URL"), req.ScanID, tier)),
		CancelURL:  stripe.String(fmt.Sprintf("%s/free-audit", os.Getenv("BASE_URL"))),
	}
	params.AddMetadata("scan_id", req.ScanID)
	params.AddMetadata("tier", tier)

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Stripe session error: %v", err)
		http.Error(w, "Betaling initialiseren fout", http.StatusInternalServerError)
		return
	}

	// Return HTMX redirect snippet
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<script>window.location.href = '%s';</script>`, sess.URL)
}

// HandleWebhook processes Stripe webhook events
func (h *HTTPHandlerV2) HandleWebhook(w http.ResponseWriter, r *http.Request) {
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
		tier := sess.Metadata["tier"]
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

		log.Printf("Successfully processed payment for scan %s (tier: %s)", scanID, tier)
	}

	w.WriteHeader(http.StatusOK)
}

// HandleSuccess shows payment success page
func (h *HTTPHandlerV2) HandleSuccess(w http.ResponseWriter, r *http.Request) {
	scanID := r.URL.Query().Get("scan_id")
	tier := r.URL.Query().Get("tier")

	tierName := "Unknown"
	switch tier {
	case "tier1":
		tierName = "The Audit Start (€499)"
	case "tier2":
		tierName = "The Compliance Shield (€2.450)"
	case "tier3":
		tierName = "The Enterprise Partner (€4.900)"
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="nl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Betaling Geslaagd - LEONA & CRAVIT</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://fonts.googleapis.com/css2?family=Funnel+Display:wght@700&display=swap" rel="stylesheet">
</head>
<body style="background-color: #0A192F;" class="text-white font-sans">
    <div class="min-h-screen flex items-center justify-center px-4">
        <div class="max-w-2xl w-full text-center bg-gray-900 p-12 rounded-3xl border-2 border-blue-500 shadow-2xl">
            <div class="mb-8">
                <svg class="w-24 h-24 mx-auto text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                </svg>
            </div>
            <h1 style="font-family: 'Funnel Display', serif;" class="text-4xl md:text-5xl font-bold mb-4">Betaling Geslaagd!</h1>
            <div class="bg-blue-900/30 border border-blue-500 rounded-lg p-6 mb-8">
                <p class="text-xl text-gray-200 mb-2">
                    U heeft succesvol <strong class="text-blue-400">%s</strong> aangeschaft.
                </p>
                <p class="text-gray-400 text-sm">
                    Uw volledige CRA Compliance Rapport wordt nu gegenereerd.
                </p>
            </div>
            <div class="bg-gray-800 p-6 rounded-lg mb-8">
                <p class="text-sm text-gray-400 mb-2">Scan ID</p>
                <code class="bg-gray-950 px-4 py-2 rounded text-green-400 font-mono text-lg">%s</code>
            </div>
            <div class="space-y-4 mb-8">
                <div class="flex items-start gap-4 text-left">
                    <span class="text-2xl">📧</span>
                    <div>
                        <p class="font-bold">Email Delivery</p>
                        <p class="text-sm text-gray-400">PDF rapport wordt binnen 5-10 minuten verzonden</p>
                    </div>
                </div>
                <div class="flex items-start gap-4 text-left">
                    <span class="text-2xl">📄</span>
                    <div>
                        <p class="font-bold">Factuur</p>
                        <p class="text-sm text-gray-400">Direct beschikbaar in uw inbox</p>
                    </div>
                </div>
                <div class="flex items-start gap-4 text-left">
                    <span class="text-2xl">🔒</span>
                    <div>
                        <p class="font-bold">Veilige Opslag</p>
                        <p class="text-sm text-gray-400">Uw scan is veilig opgeslagen in onze database</p>
                    </div>
                </div>
            </div>
            <a href="/free-audit" style="background-color: #4169E1;" class="inline-block px-8 py-4 rounded-lg font-bold text-lg hover:bg-blue-600 transition shadow-lg">
                Nieuwe Scan Uitvoeren →
            </a>
            <div class="mt-8 pt-8 border-t border-gray-700">
                <p class="text-sm text-gray-500">
                    Vragen? Email <a href="mailto:support@craleona.be" class="text-blue-400 hover:underline">support@craleona.be</a>
                </p>
            </div>
        </div>
    </div>
</body>
</html>
	`, tierName, scanID)
}
