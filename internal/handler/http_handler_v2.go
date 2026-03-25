package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"leona-scanner/internal/database"
	"leona-scanner/internal/i18n"
	"leona-scanner/internal/scanner"
	"leona-scanner/internal/services"
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
	i18nManager    *i18n.I18n
	mailgunService *services.MailgunService
}

// getTemplateFuncs returns the standard template function map
func getTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
}

// HandleNotFound serves custom branded 404 page
func (h *HTTPHandlerV2) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	tmpl, err := template.ParseFiles("templates/pages/404.html")
	if err != nil {
		http.Error(w, "404 - Page not found", http.StatusNotFound)
		log.Printf("404 template parse error: %v", err)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "404 - Page not found", http.StatusNotFound)
		log.Printf("404 template execute error: %v", err)
	}
}

// HandlePage serves any page from templates/pages/ using the base layout
// URL /about → templates/pages/about.html
func (h *HTTPHandlerV2) HandlePage(pageName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create template set with FuncMap before parsing
		tmpl := template.New("").Funcs(getTemplateFuncs())
		tmpl, err := tmpl.ParseFiles(
			"templates/layouts/base.html",
			"templates/components/ai-banner.html",
			"templates/components/navbar.html",
			"templates/components/footer.html",
			"templates/components/hero.html",
			"templates/components/example-report.html",
			"templates/components/faq.html",
			"templates/partials/logo.html",
			"templates/partials/nav_links.html",
			"templates/partials/demo-button.html",
			fmt.Sprintf("templates/pages/%s.html", pageName),
		)
		if err != nil {
			http.Error(w, "Template fout", http.StatusInternalServerError)
			log.Printf("Template parse error for page '%s': %v", pageName, err)
			return
		}

		// Determine language from request
		lang := "nl" // default
		if h.i18nManager != nil {
			lang = h.i18nManager.GetLanguageFromRequest(
				r.Header.Get("Accept-Language"),
				r.URL.Query().Get("lang"),
			)
		}

		// Use shared data with navbar and i18n
		data := NewSharedData(pageName)
		if h.i18nManager != nil {
			data["T"] = h.i18nManager.GetAll(lang)
			data["Lang"] = lang

			// Get navigation translations and render navbar with them
			if navData := h.i18nManager.Get(lang, "navigation"); navData != nil {
				if navMap, ok := navData.(map[string]interface{}); ok {
					// Templ navbar removed
					_ = navMap
				}
			}
		}

		// Check for success message (for contact form)
		if r.URL.Query().Get("success") == "true" {
			data["SuccessMessage"] = true
		}

		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
			log.Printf("Template execute error for page '%s': %v", pageName, err)
		}
	}
}

// NewHTTPHandlerV2 creates a new HTTP handler with all services
func NewHTTPHandlerV2(
	scannerService *usecase.ScannerService,
	pdfService *usecase.PDFService,
	i18nManager *i18n.I18n,
	mailgunService *services.MailgunService,
) *HTTPHandlerV2 {
	return &HTTPHandlerV2{
		scannerService: scannerService,
		pdfService:     pdfService,
		i18nManager:    i18nManager,
		mailgunService: mailgunService,
	}
}

// HandleIndex serves the landing page
func (h *HTTPHandlerV2) HandleIndex(w http.ResponseWriter, r *http.Request) {
	// Recursively find all .html files in templates directory
	var files []string
	err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Template scan error", http.StatusInternalServerError)
		log.Printf("Template walk error: %v", err)
		return
	}

	// Create template with helper functions
	tmpl, err := template.New("index.html").Funcs(getTemplateFuncs()).ParseFiles(files...)
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	// Prepare data for the template (includes dark navbar)
	data := map[string]interface{}{}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
		log.Printf("Template execute error: %v", err)
	}
}

// HandleDemo serves the standalone demo page (no navbar/footer)
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

// HandleCRAAssessment serves the standalone CRA assessment wizard (no navbar/footer)
func (h *HTTPHandlerV2) HandleCRAAssessment(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/pages/cra-assessment.html")
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	// Standalone wizard page - no base layout
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
		log.Printf("Template execute error: %v", err)
	}
}

// HandleCRAApplicability serves the standalone CRA applicability check (no navbar/footer)
func (h *HTTPHandlerV2) HandleCRAApplicability(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/pages/cra-applicability.html")
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	// Standalone applicability page - no base layout
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
		log.Printf("Template execute error: %v", err)
	}
}

// HandleLEONAApplicability serves the standalone LEONA applicability check (no navbar/footer)
func (h *HTTPHandlerV2) HandleLEONAApplicability(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/pages/leona-applicability.html")
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	// Standalone LEONA applicability page - no base layout
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
		log.Printf("Template execute error: %v", err)
	}
}

// HandleShieldMarch serves the Shield March promotional landing page (no navbar/footer)
func (h *HTTPHandlerV2) HandleShieldMarch(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/pages/shield-march.html")
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	// Standalone promotional page - no base layout
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
		log.Printf("Template execute error: %v", err)
	}
}

// ShieldMarchSubmission represents the form data from Shield March campaign
type ShieldMarchSubmission struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Product string `json:"product"`
	Code    string `json:"code"`
}

// HandleShieldMarchSubmit processes the Shield March form submission and sends confirmation email
func (h *HTTPHandlerV2) HandleShieldMarchSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var submission ShieldMarchSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		log.Printf("Shield March submission decode error: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate promo code
	if submission.Code != "SHIELDMARCH5" {
		log.Printf("Shield March invalid code: %s", submission.Code)
		http.Error(w, "Invalid promo code", http.StatusBadRequest)
		return
	}

	// Log submission
	log.Printf(
		"Shield March: %s (%s) - %s / %s",
		submission.Name, submission.Email, submission.Company, submission.Product,
	)

	// Send confirmation email via Mailgun
	if h.mailgunService != nil {
		h.sendShieldMarchEmails(submission)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		log.Printf("Shield March response encode error: %v", err)
	}
}

// sendShieldMarchEmails sends confirmation and notification emails for Shield March submissions
func (h *HTTPHandlerV2) sendShieldMarchEmails(submission ShieldMarchSubmission) {
	emailSubject := "Actie vereist: Uw Binaire Snapshot (Code: SHIELDMARCH5) staat klaar 🛡️"
	emailBody := fmt.Sprintf(`Beste %s,

Bedankt voor uw aanvraag voor een gratis Binaire Snapshot ter waarde van 2.495 euro.
U heeft hiermee de eerste stap gezet om uw organisatie te beschermen tegen de binaire
risico's van de Cyber Resilience Act.

Omdat wij een **24-uurs garantie** hanteren voor onze rapportages, hebben we uw input
direct nodig om de Assessor™ te activeren.

**Uw volgende stappen:**

1. **Beveiligde Upload:** Antwoord op deze e-mail met uw SBOM (CycloneDX/SPDX)
   of een download link naar uw firmware binaries. Alle uploads worden versleuteld
   opgeslagen.

2. **Binaire Analyse:** Zodra de upload is voltooid, start onze adaptive compliance
   engine de Triple-Check op uw machinecode.

3. **Uw Rapport:** Binnen 24 uur ontvangt u uw officiële CRA-statusrapport, inclusief
   de eliminatie van valse meldingen en een overzicht van de binaire reachability.

**Waarom dit vandaag moet gebeuren:**
Onder de nieuwe EU-wetgeving is "we wisten het niet" geen juridisch verweer meer.
Met dit rapport heeft u zwart-op-wit bewijs van uw binaire integriteit voor uw
directie en toezichthouders.

Heeft u vragen over het exportformat van uw SBOM? Antwoord direct op deze e-mail,
ons engineering team staat klaar.

Relentless regards,

LEONA Team
LEONA Compliance - Operationalizing Binary Integrity

---

**Uw gegevens:**
Bedrijf: %s
Product: %s
Actiecode: %s
`, submission.Name, submission.Company, submission.Product, submission.Code)

	// Convert plain text email body to simple HTML
	emailBodyHTML := "<html><body><pre style='font-family: system-ui, sans-serif; " +
		"white-space: pre-wrap;'>" + emailBody + "</pre></body></html>"

	if err := h.mailgunService.SendHTMLEmail(
		submission.Email,
		emailSubject,
		emailBodyHTML,
	); err != nil {
		log.Printf("Shield March email send error: %v", err)
	}

	// Also send notification to LEONA team
	notificationBody := fmt.Sprintf(`🛡️ SHIELD MARCH SUBMISSION

Naam: %s
Email: %s
Bedrijf: %s
Product: %s
Code: %s

Actie: Deze lead verwacht binnen 24 uur instructies voor SBOM/firmware upload.
`, submission.Name, submission.Email, submission.Company, submission.Product, submission.Code)

	notificationBodyHTML := "<html><body><pre style='font-family: system-ui, " +
		"sans-serif;'>" + notificationBody + "</pre></body></html>"

	if err := h.mailgunService.SendHTMLEmail(
		"kim@eliama.agency",
		"🛡️ Shield March Lead: "+submission.Company,
		notificationBodyHTML,
	); err != nil {
		log.Printf("Shield March notification error: %v", err)
	}
}

// HandleServices serves the enterprise services page
func (h *HTTPHandlerV2) HandleServices(w http.ResponseWriter, r *http.Request) {
	// Create template set with FuncMap before parsing
	tmpl := template.New("").Funcs(getTemplateFuncs())
	tmpl, err := tmpl.ParseFiles(
		"templates/layouts/base.html",
		"templates/components/ai-banner.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/partials/logo.html",
		"templates/partials/nav_links.html",
		"templates/pages/services.html",
	)
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	// Use shared data with navbar
	data := NewSharedData("diensten")

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
		log.Printf("Template execute error: %v", err)
	}
}

// HandleProducts serves the products page
func (h *HTTPHandlerV2) HandleProducts(w http.ResponseWriter, r *http.Request) {
	// Create template set with FuncMap before parsing
	tmpl := template.New("").Funcs(getTemplateFuncs())
	tmpl, err := tmpl.ParseFiles(
		"templates/layouts/base.html",
		"templates/components/ai-banner.html",
		"templates/components/navbar.html",
		"templates/components/footer.html",
		"templates/partials/logo.html",
		"templates/partials/nav_links.html",
		"templates/components/cta-demo.html",
		"templates/components/feature-grid.html",
		"templates/pages/products-simple.html",
	)
	if err != nil {
		http.Error(w, "Template fout", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	// Start with shared data
	data := NewSharedData("producten")

	// Add Feature Section 1
	data["Feature1"] = NewFeatureSection(
		"SBOM Scanning",
		"Real-time kwetsbaarheid detectie",
		"Upload uw SBOM en ontvang direct een volledig CRA compliance rapport met CVE scanning.",
		[]map[string]string{
			{"Title": "Automatische CVE scanning", "Description": "Real-time detectie via NVD API 2.0"},
			{"Title": "CycloneDX & SPDX support", "Description": "Ondersteunt beide SBOM formaten"},
			{"Title": "42-pagina TCF rapport", "Description": "Compleet Technical Construction File voor CRA Annex VII"},
		},
		"",
	)

	// Add Feature Section 2
	data["Feature2"] = NewFeatureSection(
		"CI/CD Integratie",
		"Geautomatiseerde compliance",
		"Integreer LEONA direct in uw build pipeline voor continue compliance monitoring.",
		[]map[string]string{
			{"Title": "Jenkins/GitLab/Bitbucket", "Description": "Native integratie met alle major CI/CD platforms"},
			{"Title": "Yocto layer analyse", "Description": "Specifieke support voor embedded Linux builds"},
			{"Title": "API-first design", "Description": "OpenAPI 3.0 spec voor custom integraties"},
		},
		"",
	)

	// Add CTA
	data["CTA"] = NewCTAData(
		"Klaar voor CRA compliance?",
		"Ontdek hoe LEONA uw embedded Linux producten CRA-ready maakt in minder dan een week.",
		"Vraag Demo Aan",
	)

	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
		log.Printf("Template execute error: %v", err)
	}
}

// HandleSnapshot serves the SNAPSHOT product page
func (h *HTTPHandlerV2) HandleSnapshot(w http.ResponseWriter, r *http.Request) {
	// TODO: Convert to base layout - create templates/pages/snapshot.html
	tmpl, err := template.ParseFiles("templates/snapshot.html")
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

// HandleTCFBundle serves the TCF Bundle product page
func (h *HTTPHandlerV2) HandleTCFBundle(w http.ResponseWriter, r *http.Request) {
	// TODO: Convert to base layout - create templates/pages/tcf-bundle.html
	tmpl, err := template.ParseFiles("templates/tcf-bundle.html")
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
	// TODO: Convert to base layout - create templates/pages/insights.html
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
	// TODO: Convert to base layout - create templates/pages/kennisbank.html
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
	h.handleCheckout(
		w, r, "tier2", 245000,
		"The Compliance Shield",
		"Unlimited scans +-VERIFIED certificaat + Email support", //nolint:misspell
	)
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
    <title>Betaling Geslaagd - LEONA</title>
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

// HandleLegalAssessmentSubmit processes legal assessment form submissions
func (h *HTTPHandlerV2) HandleLegalAssessmentSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Parse form data
	name := strings.TrimSpace(r.FormValue("name"))
	lawFirm := strings.TrimSpace(r.FormValue("law-firm"))
	emailAddr := strings.TrimSpace(r.FormValue("email"))
	scoreStr := r.FormValue("score")
	answersJSON := r.FormValue("answers")

	// Validate required fields
	if name == "" || lawFirm == "" || emailAddr == "" || scoreStr == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Parse answers
	var answers []int
	if err := json.Unmarshal([]byte(answersJSON), &answers); err != nil {
		http.Error(w, "Invalid answers format", http.StatusBadRequest)
		return
	}

	// Store in database as lead (if db is available)
	if db != nil {
		notes := fmt.Sprintf("CRA Legal Assessment - Score: %s/100", scoreStr)
		lead := &database.Lead{
			Email:       emailAddr,
			FirstName:   &name,
			CompanyName: &lawFirm,
			Notes:       &notes,
			LeadType:    "legal-assessment",
			Source:      "website",
			Status:      "new",
		}

		if err := db.CreateLead(r.Context(), lead); err != nil {
			log.Printf("Error saving assessment lead: %v\n", err)
			// Continue anyway - don't block user
		}
	}

	// Send emails via Mailgun
	if h.mailgunService != nil {
		// Send notification to admin
		go h.sendLegalAssessmentNotificationViaMailgun(name, lawFirm, emailAddr, scoreStr, answers)

		// Send confirmation to user
		go h.sendLegalAssessmentConfirmationViaMailgun(emailAddr, name, scoreStr, answers)
	} else {
		log.Println("⚠️  Mailgun not configured - legal assessment emails not sent")
	}

	// Return success message
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<div class="rounded-lg bg-green-50 border border-green-200 p-4">
			<p class="text-sm font-semibold text-green-800">✓ Rapport verzonden!</p>
			<p class="text-sm text-green-700 mt-1">U ontvangt uw volledige analyse binnen enkele minuten op ` + emailAddr + `</p>
		</div>
		<script>
			// Reset form after 5 seconds
			setTimeout(() => {
				const form = document.querySelector('form');
				if (form) form.reset();
			}, 5000);
		</script>
	`))
}

// sendLegalAssessmentNotificationViaMailgun sends admin notification via Mailgun
func (h *HTTPHandlerV2) sendLegalAssessmentNotificationViaMailgun(name, lawFirm, email, score string, answers []int) {
	getScoreColor := func(score string) string {
		var scoreInt int
		fmt.Sscanf(score, "%d", &scoreInt)
		if scoreInt <= 40 {
			return "#dc2626"
		}
		if scoreInt <= 80 {
			return "#f59e0b"
		}
		return "#059669"
	}

	getScoreLabel := func(score string) string {
		var scoreInt int
		fmt.Sscanf(score, "%d", &scoreInt)
		if scoreInt <= 40 {
			return "Technisch Blind"
		}
		if scoreInt <= 80 {
			return "Juridisch Sterk, Technisch Kwetsbaar"
		}
		return "CRA-Ready"
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: system-ui, sans-serif; line-height: 1.6; color: #1a1a1a; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(135deg, #1e40af 0%%%%, #3b82f6 100%%%%); color: white; padding: 30px; border-radius: 8px; }
		.score-badge { display: inline-block; background-color: %s; color: white; font-size: 48px; font-weight: bold; width: 120px; height: 120px; border-radius: 60px; line-height: 120px; margin: 20px 0; }
		.info { background: #f9fafb; padding: 15px; margin: 10px 0; border-left: 4px solid #3b82f6; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>🎯 Nieuwe CRA Legal Assessment</h1>
		</div>
		<div style="text-align: center;">
			<div class="score-badge">%s</div>
			<p><strong>%s</strong></p>
		</div>
		<div class="info"><strong>Naam:</strong> %s</div>
		<div class="info"><strong>Advocatenkantoor:</strong> %s</div>
		<div class="info"><strong>Email:</strong> <a href="mailto:%s">%s</a></div>
	</div>
</body>
</html>
`, getScoreColor(score), score, getScoreLabel(score), name, lawFirm, email, email)

	subject := fmt.Sprintf("🎯 Nieuwe CRA Legal Assessment - %s/100 punten", score)
	if err := h.mailgunService.SendHTMLEmail("kim@leonacompliance.be", subject, body); err != nil {
		log.Printf("❌ ERROR: Failed to send legal assessment notification: %v", err)
	} else {
		log.Printf("✅ SUCCESS: Legal assessment notification sent for %s (%s)", name, lawFirm)
	}
}

// sendLegalAssessmentConfirmationViaMailgun sends confirmation email to user
func (h *HTTPHandlerV2) sendLegalAssessmentConfirmationViaMailgun(email, name, score string, answers []int) {
	getScoreColor := func(score string) string {
		var scoreInt int
		fmt.Sscanf(score, "%d", &scoreInt)
		if scoreInt <= 40 {
			return "#dc2626"
		}
		if scoreInt <= 80 {
			return "#f59e0b"
		}
		return "#059669"
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: system-ui, sans-serif; line-height: 1.6; color: #1a1a1a; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(135deg, #1e40af 0%%%%, #3b82f6 100%%%%); color: white; padding: 30px; border-radius: 8px; }
		.score-badge { display: inline-block; background-color: %s; color: white; font-size: 48px; font-weight: bold; width: 120px; height: 120px; border-radius: 60px; line-height: 120px; margin: 20px 0; }
		.cta { display: inline-block; background-color: #2563eb; color: #ffffff; text-decoration: none; padding: 14px 28px; border-radius: 6px; font-weight: 600; margin: 20px 0; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Uw CRA Legal-Tech Gap Assessment</h1>
			<p>Persoonlijk rapport voor %s</p>
		</div>
		<div style="text-align: center;">
			<div class="score-badge">%s</div>
			<p style="font-size: 18px; margin: 20px 0;">Dank u voor het invullen van de assessment. We nemen zo snel mogelijk contact met u op.</p>
			<a href="https://leonacompliance.be/partner-overleg" class="cta">Plan een Partner Overleg</a>
		</div>
	</div>
</body>
</html>
`, getScoreColor(score), name, score)

	subject := "Uw CRA Legal-Tech Gap Assessment Rapport"
	if err := h.mailgunService.SendHTMLEmail(email, subject, body); err != nil {
		log.Printf("⚠️  WARNING: Failed to send confirmation email to %s: %v", email, err)
	} else {
		log.Printf("📧 Confirmation email sent to %s", email)
	}
}
