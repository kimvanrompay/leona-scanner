package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"leona-scanner/internal/database"

	"github.com/VictorAvelar/mollie-api-go/v3/mollie"
	"gopkg.in/gomail.v2"
)

// SnapshotSubmission contains all the data from the snapshot audit form
type SnapshotSubmission struct {
	// Contact
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Company   string `json:"company"`
	Phone     string `json:"phone"`

	// Build System
	BuildSystem        string `json:"build_system"`
	BuildSystemVersion string `json:"build_system_version"`
	TargetArchitecture string `json:"target_architecture"`
	KernelVersion      string `json:"kernel_version"`
	Libc               string `json:"libc"`

	// Product
	ProductName     string   `json:"product_name"`
	ProductCategory string   `json:"product_category"`
	Connectivity    []string `json:"connectivity"`
	AnnualVolume    string   `json:"annual_volume"`

	// Security
	SecureBoot      string   `json:"secure_boot"`
	TPM             string   `json:"tpm"`
	OTAFeatures     []string `json:"ota_features"`
	UpdateFramework string   `json:"update_framework"`

	// Artifacts
	ArtifactAccess     string   `json:"artifact_access"`
	EstimatedSize      string   `json:"estimated_size"`
	AvailableArtifacts []string `json:"available_artifacts"`

	// Context
	Timeline        string `json:"timeline"`
	Concerns        string `json:"concerns"`
	AdditionalNotes string `json:"additional_notes"`
}

// HandleSnapshotSubmit processes snapshot audit submissions
//
//nolint:funlen,gocognit,gocyclo,cyclop // Complex form processing and payment flow
func (h *HTTPHandlerV2) HandleSnapshotSubmit(w http.ResponseWriter, r *http.Request) {
	log.Printf("📋 Snapshot Audit submission received from %s", r.RemoteAddr)

	if err := r.ParseForm(); err != nil {
		log.Printf("❌ Failed to parse form: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Extract and validate form data
	submission := &SnapshotSubmission{
		FirstName:          strings.TrimSpace(r.FormValue("first-name")),
		LastName:           strings.TrimSpace(r.FormValue("last-name")),
		Email:              strings.TrimSpace(r.FormValue("email")),
		Company:            strings.TrimSpace(r.FormValue("company")),
		Phone:              strings.TrimSpace(r.FormValue("phone")),
		BuildSystem:        r.FormValue("build-system"),
		BuildSystemVersion: strings.TrimSpace(r.FormValue("build-system-version")),
		TargetArchitecture: r.FormValue("target-architecture"),
		KernelVersion:      strings.TrimSpace(r.FormValue("kernel-version")),
		Libc:               r.FormValue("libc"),
		ProductName:        strings.TrimSpace(r.FormValue("product-name")),
		ProductCategory:    r.FormValue("product-category"),
		Connectivity:       r.Form["connectivity"],
		AnnualVolume:       r.FormValue("annual-volume"),
		SecureBoot:         r.FormValue("secure-boot"),
		TPM:                r.FormValue("tpm"),
		OTAFeatures:        r.Form["ota-features"],
		UpdateFramework:    strings.TrimSpace(r.FormValue("update-framework")),
		ArtifactAccess:     r.FormValue("artifact-access"),
		EstimatedSize:      r.FormValue("estimated-size"),
		AvailableArtifacts: r.Form["available-artifacts"],
		Timeline:           r.FormValue("timeline"),
		Concerns:           strings.TrimSpace(r.FormValue("concerns")),
		AdditionalNotes:    strings.TrimSpace(r.FormValue("additional-notes")),
	}

	// Validate required fields
	if submission.FirstName == "" || submission.LastName == "" || submission.Email == "" ||
		submission.Company == "" || submission.BuildSystem == "" || submission.TargetArchitecture == "" ||
		submission.KernelVersion == "" || submission.Libc == "" || submission.ProductName == "" ||
		submission.ProductCategory == "" || submission.AnnualVolume == "" || submission.ArtifactAccess == "" {
		log.Printf("❌ Missing required fields in snapshot submission from %s", submission.Email)
		http.Error(w, "Alle verplichte velden moeten ingevuld zijn", http.StatusBadRequest)
		return
	}

	// Validate email
	if !strings.Contains(submission.Email, "@") {
		log.Printf("❌ Invalid email in snapshot submission: %s", submission.Email)
		http.Error(w, "Geldig e-mailadres vereist", http.StatusBadRequest)
		return
	}

	// Log full submission for debugging
	submissionJSON, err := json.MarshalIndent(submission, "", "  ")
	if err == nil {
		log.Printf("📦 Full snapshot submission:\n%s", string(submissionJSON))
	}

	// Save to database
	if db != nil {
		message := fmt.Sprintf("Snapshot Audit Request - %s (%s)\nBuild: %s %s\nKernel: %s\nProduct: %s (%s)",
			submission.ProductName, submission.ProductCategory,
			submission.BuildSystem, submission.BuildSystemVersion,
			submission.KernelVersion,
			submission.ProductName, submission.TargetArchitecture)

		contact := &database.ContactSubmission{
			FirstName: submission.FirstName,
			LastName:  submission.LastName,
			Email:     submission.Email,
			Company:   submission.Company,
			Message:   message,
			Solution:  "snapshot-audit",
			Status:    "pending-payment",
		}
		if err := db.CreateContactSubmission(r.Context(), contact); err != nil {
			log.Printf("⚠️  Failed to save snapshot submission: %v", err)
		}
	}

	// Send detailed notification email to Kim
	if err := h.sendSnapshotNotification(submission); err != nil {
		log.Printf("❌ Failed to send snapshot notification to kim@leonacompliance.be: %v", err)
	} else {
		log.Printf("✅ Snapshot notification sent to kim@leonacompliance.be")
	}

	// Create Mollie payment
	paymentURL, err := h.createSnapshotPayment(r.Context(), submission)
	if err != nil {
		log.Printf("❌ Failed to create Mollie payment: %v", err)
		http.Error(w, "Payment initialization failed", http.StatusInternalServerError)
		return
	}

	log.Printf("💳 Mollie payment created for %s %s (%s) - redirecting to: %s",
		submission.FirstName, submission.LastName, submission.Email, paymentURL)

	// Return HTMX response with redirect to payment
	w.Header().Set("Content-Type", "text/html")
	//nolint:lll,errcheck // HTML template
	_, _ = w.Write([]byte(fmt.Sprintf(` //nolint:errcheck // HTTP error already handled
		<div class="text-center py-12">
			<div class="mb-6">
				<svg class="w-20 h-20 mx-auto text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
				</svg>
			</div>
			<h2 class="text-2xl font-bold text-gray-900 mb-4">Snapshot Audit Aangevraagd</h2>
			<p class="text-gray-600 mb-6">U wordt doorgestuurd naar de beveiligde betaalomgeving...</p>
			<p class="text-sm text-gray-500 mb-4">Indien u niet automatisch wordt doorgestuurd:</p>
			<a href="%s" class="inline-block bg-blue-600 hover:bg-blue-700 text-white font-semibold px-8 py-3 rounded-lg transition-colors">
				Ga naar Betaling
			</a>
		</div>
		<script>
			setTimeout(() => {
				window.location.href = "%s";
			}, 2000);
		</script>
	`, paymentURL, paymentURL)))
}

// createSnapshotPayment creates a Mollie payment for the snapshot audit
func (h *HTTPHandlerV2) createSnapshotPayment(ctx context.Context, submission *SnapshotSubmission) (string, error) {
	mollieAPIKey := os.Getenv("MOLLIE_API_KEY")
	if mollieAPIKey == "" {
		return "", fmt.Errorf("MOLLIE_API_KEY not configured")
	}

	config := mollie.NewConfig(true, mollie.APITokenEnv)
	client, err := mollie.NewClient(nil, config)
	if err != nil {
		return "", fmt.Errorf("failed to initialize Mollie client: %w", err)
	}

	// Create payment description
	description := fmt.Sprintf("Snapshot Audit - %s (%s)", submission.ProductName, submission.Company)

	// Create payment
	payment := mollie.Payment{
		Amount: &mollie.Amount{
			Currency: "EUR",
			Value:    "2495.00",
		},
		Description: description,
		RedirectURL: "https://leonacompliance.be/snapshot/success",
		WebhookURL:  "https://leonacompliance.be/webhook/mollie",
		Metadata: map[string]interface{}{
			"type":           "snapshot-audit",
			"customer_email": submission.Email,
			"customer_name":  fmt.Sprintf("%s %s", submission.FirstName, submission.LastName),
			"company":        submission.Company,
			"product_name":   submission.ProductName,
			"build_system":   submission.BuildSystem,
			"kernel_version": submission.KernelVersion,
			"target_arch":    submission.TargetArchitecture,
		},
	}

	_, createdPayment, err := client.Payments.Create(ctx, payment, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create payment: %w", err)
	}

	// Get checkout URL
	if createdPayment.Links.Checkout == nil {
		return "", fmt.Errorf("no checkout URL in payment response")
	}

	return createdPayment.Links.Checkout.Href, nil
}

// sendSnapshotNotification sends a detailed email to Kim with all submission data
//
//nolint:funlen // Email template functions are naturally longer
func (h *HTTPHandlerV2) sendSnapshotNotification(s *SnapshotSubmission) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "kim@leonacompliance.be")
	m.SetHeader("Subject", fmt.Sprintf("💰 SNAPSHOT AUDIT: %s - %s (%s)", s.ProductName, s.Company, s.Email))

	// Build HTML email with all details
	connectivity := strings.Join(s.Connectivity, ", ")
	if connectivity == "" {
		connectivity = "Niet gespecificeerd"
	}

	otaFeatures := strings.Join(s.OTAFeatures, ", ")
	if otaFeatures == "" {
		otaFeatures = "Niet gespecificeerd"
	}

	artifacts := strings.Join(s.AvailableArtifacts, ", ")
	if artifacts == "" {
		artifacts = "Niet gespecificeerd"
	}

	//nolint:lll,misspell // Email template with Dutch text
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; background: #f5f5f5; }
        .container { max-width: 700px; margin: 0 auto; background: white; }
        .header { background: linear-gradient(135deg, #1e3a8a 0%%, #3b82f6 100%%); color: white; padding: 40px; }
        .header h1 { margin: 0; font-size: 28px; }
        .header .price { background: #10b981; display: inline-block; padding: 8px 16px; border-radius: 6px; font-weight: bold; margin-top: 12px; }
        .section { padding: 30px; border-bottom: 1px solid #e5e5e5; }
        .section h2 { color: #1e3a8a; font-size: 18px; margin: 0 0 20px 0; text-transform: uppercase; letter-spacing: 0.5px; }
        .field { margin-bottom: 16px; }
        .field-label { font-weight: 600; color: #666; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
        .field-value { margin-top: 4px; font-size: 15px; color: #1a1a1a; }
        .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        .highlight { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 16px; margin-top: 8px; }
        .footer { padding: 30px; background: #f9fafb; text-align: center; color: #666; font-size: 13px; }
        .badge { display: inline-block; padding: 4px 12px; border-radius: 4px; font-size: 11px; font-weight: bold; }
        .badge-primary { background: #dbeafe; color: #1e40af; }
        .badge-success { background: #d1fae5; color: #065f46; }
        .badge-warning { background: #fef3c7; color: #92400e; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>💰 SNAPSHOT AUDIT AANVRAAG</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">Nieuwe betaalde audit aanvraag</p>
            <div class="price">€2.495</div>
        </div>

        <div class="section">
            <h2>📋 Contact & Bedrijf</h2>
            <div class="grid">
                <div class="field">
                    <div class="field-label">Contactpersoon</div>
                    <div class="field-value">%s %s</div>
                </div>
                <div class="field">
                    <div class="field-label">E-mail</div>
                    <div class="field-value"><a href="mailto:%s">%s</a></div>
                </div>
                <div class="field">
                    <div class="field-label">Bedrijf</div>
                    <div class="field-value">%s</div>
                </div>
                <div class="field">
                    <div class="field-label">Telefoon</div>
                    <div class="field-value">%s</div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2>🏭 Product Informatie</h2>
            <div class="field">
                <div class="field-label">Product Naam</div>
                <div class="field-value"><strong>%s</strong></div>
            </div>
            <div class="grid">
                <div class="field">
                    <div class="field-label">Categorie</div>
                    <div class="field-value">%s</div>
                </div>
                <div class="field">
                    <div class="field-label">Jaarlijks Volume</div>
                    <div class="field-value">%s</div>
                </div>
            </div>
            <div class="field">
                <div class="field-label">Connectiviteit</div>
                <div class="field-value">%s</div>
            </div>
            <div class="field">
                <div class="field-label">Market Entry Tijdlijn</div>
                <div class="field-value"><span class="badge badge-warning">%s</span></div>
            </div>
        </div>

        <div class="section">
            <h2>🔧 Build System</h2>
            <div class="grid">
                <div class="field">
                    <div class="field-label">Build Framework</div>
                    <div class="field-value"><strong>%s</strong> %s</div>
                </div>
                <div class="field">
                    <div class="field-label">Target Architectuur</div>
                    <div class="field-value">%s</div>
                </div>
                <div class="field">
                    <div class="field-label">Kernel Versie</div>
                    <div class="field-value"><span class="badge badge-primary">%s</span></div>
                </div>
                <div class="field">
                    <div class="field-label">C Library</div>
                    <div class="field-value">%s</div>
                </div>
            </div>
        </div>

        <div class="section">
            <h2>🔒 Security Features</h2>
            <div class="grid">
                <div class="field">
                    <div class="field-label">Secure Boot</div>
                    <div class="field-value">%s</div>
                </div>
                <div class="field">
                    <div class="field-label">TPM / Secure Element</div>
                    <div class="field-value">%s</div>
                </div>
            </div>
            <div class="field">
                <div class="field-label">OTA Update Features</div>
                <div class="field-value">%s</div>
            </div>
            <div class="field">
                <div class="field-label">Update Framework</div>
                <div class="field-value">%s</div>
            </div>
        </div>

        <div class="section">
            <h2>📦 Build Artifacts</h2>
            <div class="grid">
                <div class="field">
                    <div class="field-label">Transfer Methode</div>
                    <div class="field-value">%s</div>
                </div>
                <div class="field">
                    <div class="field-label">Geschatte Grootte</div>
                    <div class="field-value">%s</div>
                </div>
            </div>
            <div class="field">
                <div class="field-label">Beschikbare Artifacts</div>
                <div class="field-value">%s</div>
            </div>
        </div>

        <div class="section">
            <h2>💭 Context & Zorgen</h2>
            <div class="field">
                <div class="field-label">Grootste CRA Zorgen</div>
                <div class="highlight">%s</div>
            </div>
            <div class="field">
                <div class="field-label">Aanvullende Opmerkingen</div>
                <div class="field-value">%s</div>
            </div>
        </div>

        <div class="footer">
            <p><strong>LEONA Compliance</strong> | Snapshot Audit Request<br/>
            Automatisch gegenereerd vanuit <a href="https://leonacompliance.be/snapshot">leonacompliance.be/snapshot</a></p>
            <p style="margin-top: 16px;"><span class="badge badge-success">BETALING PENDING</span></p>
        </div>
    </div>
</body>
</html>
`,
		s.FirstName, s.LastName,
		s.Email, s.Email,
		s.Company,
		s.Phone,
		s.ProductName,
		s.ProductCategory,
		s.AnnualVolume,
		connectivity,
		s.Timeline,
		s.BuildSystem, s.BuildSystemVersion,
		s.TargetArchitecture,
		s.KernelVersion,
		s.Libc,
		s.SecureBoot,
		s.TPM,
		otaFeatures,
		s.UpdateFramework,
		s.ArtifactAccess,
		s.EstimatedSize,
		artifacts,
		s.Concerns,
		s.AdditionalNotes,
	)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true

	return d.DialAndSend(m)
}
