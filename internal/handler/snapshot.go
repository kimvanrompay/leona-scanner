package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"leona-scanner/internal/database"

	"github.com/VictorAvelar/mollie-api-go/v3/mollie"
	"gopkg.in/gomail.v2"
)

// SnapshotSubmission contains all the data from the snapshot audit form
type SnapshotSubmission struct {
	// Order Tracking
	OrderUUID string `json:"order_uuid"`

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

	// Legal
	NDAAccepted string `json:"nda_accepted"`
}

// HandleSnapshotSubmit processes snapshot audit submissions
//
//nolint:funlen,gocognit,gocyclo,cyclop // Complex form processing and payment flow
func (h *HTTPHandlerV2) HandleSnapshotSubmit(w http.ResponseWriter, r *http.Request) {
	log.Printf("[SNAPSHOT] Submission ontvangen van %s", r.RemoteAddr)

	if err := r.ParseForm(); err != nil {
		log.Printf("[ERROR] Formulier parsing mislukt: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		//nolint:errcheck,gosec,lll
		w.Write([]byte(`<div class="p-4 bg-red-50 border border-red-200 rounded-lg"><p class="text-red-800 font-semibold">Ongeldige aanvraag</p><p class="text-red-700 text-sm mt-2">Het formulier kon niet worden verwerkt. Probeer de pagina te vernieuwen.</p></div>`))
		return
	}

	// Extract and validate form data
	submission := &SnapshotSubmission{
		OrderUUID:          strings.TrimSpace(r.FormValue("order-uuid")),
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
		NDAAccepted:        r.FormValue("nda-accepted"),
	}

	// Validate required fields
	//nolint:lll,errcheck,gosec // Field validation
	if submission.OrderUUID == "" || submission.FirstName == "" || submission.LastName == "" ||
		submission.Email == "" || submission.Company == "" || submission.BuildSystem == "" ||
		submission.TargetArchitecture == "" || submission.KernelVersion == "" || submission.Libc == "" ||
		submission.ProductName == "" || submission.ProductCategory == "" ||
		submission.AnnualVolume == "" || submission.ArtifactAccess == "" || submission.NDAAccepted != "on" {
		log.Printf("[VALIDATIE] Ontbrekende velden of NDA niet geaccepteerd: %s", submission.Email)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<div class="p-4 bg-red-50 border border-red-200 rounded-lg"><p class="text-red-800 font-semibold">Formulier incompleet</p><p class="text-red-700 text-sm mt-2">Vul alle verplichte velden in en accepteer de NDA om verder te gaan.</p><button onclick="window.location.reload()" class="mt-3 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 text-sm">Opnieuw proberen</button></div>`))
		return
	}

	// Validate email
	if !strings.Contains(submission.Email, "@") {
		log.Printf("[VALIDATIE] Ongeldig e-mailadres: %s", submission.Email)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		//nolint:errcheck,gosec,lll
		w.Write([]byte(`<div class="p-4 bg-red-50 border border-red-200 rounded-lg"><p class="text-red-800 font-semibold">Ongeldig e-mailadres</p><p class="text-red-700 text-sm mt-2">Voer een geldig e-mailadres in (bijv. naam@bedrijf.be).</p><button onclick="window.location.reload()" class="mt-3 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 text-sm">Opnieuw proberen</button></div>`))
		return
	}

	// Log full submission for debugging
	submissionJSON, err := json.MarshalIndent(submission, "", "  ")
	if err == nil {
		log.Printf("[DATA] Volledige snapshot aanvraag:\n%s", string(submissionJSON))
	}

	// Save to database
	if db != nil {
		// Convert strings to pointers for optional fields
		var phone, buildSystemVersion, secureBoot, tpm, updateFramework *string
		var estimatedSize, timeline, concerns, additionalNotes *string

		if submission.Phone != "" {
			phone = &submission.Phone
		}
		if submission.BuildSystemVersion != "" {
			buildSystemVersion = &submission.BuildSystemVersion
		}
		if submission.SecureBoot != "" {
			secureBoot = &submission.SecureBoot
		}
		if submission.TPM != "" {
			tpm = &submission.TPM
		}
		if submission.UpdateFramework != "" {
			updateFramework = &submission.UpdateFramework
		}
		if submission.EstimatedSize != "" {
			estimatedSize = &submission.EstimatedSize
		}
		if submission.Timeline != "" {
			timeline = &submission.Timeline
		}
		if submission.Concerns != "" {
			concerns = &submission.Concerns
		}
		if submission.AdditionalNotes != "" {
			additionalNotes = &submission.AdditionalNotes
		}

		dbSubmission := &database.SnapshotSubmission{
			OrderUUID:          submission.OrderUUID,
			FirstName:          submission.FirstName,
			LastName:           submission.LastName,
			Email:              submission.Email,
			Company:            submission.Company,
			Phone:              phone,
			BuildSystem:        submission.BuildSystem,
			BuildSystemVersion: buildSystemVersion,
			TargetArchitecture: submission.TargetArchitecture,
			KernelVersion:      submission.KernelVersion,
			Libc:               submission.Libc,
			ProductName:        submission.ProductName,
			ProductCategory:    submission.ProductCategory,
			Connectivity:       submission.Connectivity,
			AnnualVolume:       submission.AnnualVolume,
			SecureBoot:         secureBoot,
			TPM:                tpm,
			OTAFeatures:        submission.OTAFeatures,
			UpdateFramework:    updateFramework,
			ArtifactAccess:     submission.ArtifactAccess,
			EstimatedSize:      estimatedSize,
			AvailableArtifacts: submission.AvailableArtifacts,
			Timeline:           timeline,
			Concerns:           concerns,
			AdditionalNotes:    additionalNotes,
			NDAAccepted:        true,
			PaymentStatus:      "pending",
			Status:             "payment-pending",
		}

		if err := db.CreateSnapshotSubmission(r.Context(), dbSubmission); err != nil {
			log.Printf("[WAARSCHUWING] Opslaan snapshot database mislukt: %v", err)
		} else {
			log.Printf("[DATABASE] Snapshot opgeslagen: ID=%s, Order=%s", dbSubmission.ID, dbSubmission.OrderUUID)
		}
	}

	// Send detailed notification email to Kim
	if err := h.sendSnapshotNotification(submission); err != nil {
		log.Printf("[ERROR] E-mail notificatie mislukt naar kim@leonacompliance.be: %v", err)
	} else {
		log.Printf("[SUCCESS] Notificatie verzonden naar kim@leonacompliance.be")
	}

	// Create Mollie payment
	paymentURL, err := h.createSnapshotPayment(r.Context(), submission)
	if err != nil {
		log.Printf("[ERROR] Mollie betaling mislukt: %v", err)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
		//nolint:errcheck,gosec,lll
		w.Write([]byte(`<div class="p-4 bg-red-50 border border-red-200 rounded-lg"><p class="text-red-800 font-semibold">Betalingsfout</p><p class="text-red-700 text-sm mt-2">De betaling kon niet worden geïnitialiseerd. Probeer het later opnieuw of neem contact op met kim@leonacompliance.be.</p><button onclick="window.location.reload()" class="mt-3 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 text-sm">Opnieuw proberen</button></div>`))
		return
	}

	log.Printf("[BETALING] Mollie betaling aangemaakt voor %s %s (%s) [Order: %s] -> %s",
		submission.FirstName, submission.LastName, submission.Email, submission.OrderUUID, paymentURL)

	// Return HTMX response with redirect to payment
	w.Header().Set("Content-Type", "text/html")
	//nolint:lll,errcheck,gosec // HTML template
	w.Write([]byte(fmt.Sprintf(`
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
	// Mollie SDK will read from MOLLIE_API_TOKEN environment variable
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
			"order_uuid":     submission.OrderUUID,
			"customer_email": submission.Email,
			"customer_name":  fmt.Sprintf("%s %s", submission.FirstName, submission.LastName),
			"company":        submission.Company,
			"product_name":   submission.ProductName,
			"build_system":   submission.BuildSystem,
			"kernel_version": submission.KernelVersion,
			"target_arch":    submission.TargetArchitecture,
			"nda_accepted":   "yes",
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

// HandleMollieWebhook processes Mollie webhook events for snapshot payments
//
//nolint:gocognit,nestif,gocyclo,cyclop,funlen // Webhook processing requires complex validation flow
func (h *HTTPHandlerV2) HandleMollieWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("[MOLLIE] Failed to parse webhook form: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	paymentID := r.FormValue("id")
	if paymentID == "" {
		log.Printf("[MOLLIE] Missing payment ID in webhook")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("[MOLLIE] Webhook ontvangen voor betaling: %s", paymentID)

	// Fetch payment details from Mollie (SDK reads from MOLLIE_API_TOKEN)
	config := mollie.NewConfig(true, mollie.APITokenEnv)
	client, err := mollie.NewClient(nil, config)
	if err != nil {
		log.Printf("[ERROR] Failed to initialize Mollie client: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	_, payment, err := client.Payments.Get(r.Context(), paymentID, nil)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch payment from Mollie: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	log.Printf("[MOLLIE] Betaling %s status: %s", paymentID, payment.Status)

	// Extract order UUID from metadata
	metadata, ok := payment.Metadata.(map[string]interface{})
	if !ok || metadata["type"] != "snapshot-audit" {
		log.Printf("[MOLLIE] Niet een snapshot-audit betaling, overslaan")
		w.WriteHeader(http.StatusOK)
		return
	}

	orderUUID, ok := metadata["order_uuid"].(string)
	if !ok || orderUUID == "" {
		log.Printf("[ERROR] Geen order_uuid in metadata")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Update payment status in database
	if db != nil {
		var paidAt *time.Time
		if payment.Status == "paid" && payment.PaidAt != nil {
			paidAt = payment.PaidAt
		}

		if err := db.UpdateSnapshotPaymentStatus(
			r.Context(),
			orderUUID,
			payment.Status,
			paymentID,
			paidAt,
		); err != nil {
			log.Printf("[ERROR] Database update mislukt: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		if payment.Status == "paid" {
			log.Printf("[SUCCESS] ✅ Betaling geslaagd voor order %s", orderUUID)
			// Send confirmation email to customer if desired
			if customerEmail, ok := metadata["customer_email"].(string); ok && customerEmail != "" {
				h.sendSnapshotPaymentConfirmation(customerEmail, orderUUID, metadata)
			}
		} else if payment.Status == "failed" || payment.Status == "canceled" {
			log.Printf("[WAARSCHUWING] ❌ Betaling %s voor order %s", payment.Status, orderUUID)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// sendSnapshotPaymentConfirmation sends a payment confirmation email to the customer
func (h *HTTPHandlerV2) sendSnapshotPaymentConfirmation(email, orderUUID string, metadata map[string]interface{}) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		log.Printf("[WAARSCHUWING] SMTP not configured, skipping confirmation email")
		return
	}

	customerName := "Geachte klant"
	if name, ok := metadata["customer_name"].(string); ok && name != "" {
		customerName = name
	}

	productName := "uw product"
	if product, ok := metadata["product_name"].(string); ok && product != "" {
		productName = product
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Betaling Ontvangen - Snapshot Audit - LEONA Compliance")

	//nolint:lll // Email HTML template
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, sans-serif; line-height: 1.6; color: #1a1a1a; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; }
        .header { background: linear-gradient(135deg, #1e3a8a 0%%, #3b82f6 100%%); color: white; padding: 40px; text-align: center; }
        .content { padding: 40px; }
        .footer { padding: 30px; background: #f9fafb; text-align: center; color: #666; font-size: 13px; }
        .button { display: inline-block; padding: 14px 28px; background: #1e3a8a; color: white; text-decoration: none; border-radius: 8px; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">✅ Betaling Ontvangen</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">Bedankt voor uw vertrouwen</p>
        </div>
        <div class="content">
            <p>Beste %s,</p>
            <p>Wij hebben uw betaling voor de <strong>Snapshot Audit</strong> van <strong>%s</strong> succesvol ontvangen.</p>
            <p><strong>Order ID:</strong> <code>%s</code></p>
            <p><strong>Bedrag:</strong> €2.495</p>
            <h3>Wat gebeurt er nu?</h3>
            <ul>
                <li>✅ Uw betaling is bevestigd</li>
                <li>📋 Kim van LEONA Compliance ontvangt uw aanvraag</li>
                <li>⏱️ Binnen 48 uur ontvangt u uw gedetailleerde Snapshot Audit</li>
                <li>📧 Alle deliverables worden naar dit e-mailadres gestuurd</li>
            </ul>
            <p>Bij vragen kunt u altijd contact opnemen met <a href="mailto:kim@leonacompliance.be">kim@leonacompliance.be</a>.</p>
            <p style="margin-top: 30px;">Met vriendelijke groet,<br/><strong>Team LEONA Compliance</strong></p>
        </div>
        <div class="footer">
            <p><strong>LEONA Compliance</strong> | CRA Compliance Experts<br/>
            <a href="https://leonacompliance.be">leonacompliance.be</a></p>
        </div>
    </div>
</body>
</html>
`, customerName, productName, orderUUID)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true

	if err := d.DialAndSend(m); err != nil {
		log.Printf("[WAARSCHUWING] Confirmation email mislukt: %v", err)
	} else {
		log.Printf("[SUCCESS] Confirmation email verzonden naar %s", email)
	}
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
	m.SetHeader("Subject", fmt.Sprintf("💰 SNAPSHOT AUDIT [%s]: %s - %s", s.OrderUUID, s.ProductName, s.Company))

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
            <p style="margin: 5px 0 0 0; font-size: 14px; font-family: monospace; opacity: 0.8;">Order ID: %s</p>
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
            <p style="margin-top: 12px; font-size: 11px; color: #10b981;">✓ NDA Accepted by Customer</p>
            <p style="margin-top: 8px;"><span class="badge badge-success">BETALING PENDING</span></p>
        </div>
    </div>
</body>
</html>
`,
		s.OrderUUID,
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
