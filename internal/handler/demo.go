package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"leona-scanner/internal/database"

	"gopkg.in/gomail.v2"
)

// HandleDemoSubmit processes demo request form submissions
func (h *HTTPHandlerV2) HandleDemoSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Extract form data
	firstName := strings.TrimSpace(r.FormValue("first-name"))
	lastName := strings.TrimSpace(r.FormValue("last-name"))
	email := strings.TrimSpace(r.FormValue("email"))
	company := strings.TrimSpace(r.FormValue("company"))
	buildSystem := r.FormValue("build-system")
	message := strings.TrimSpace(r.FormValue("message"))

	// Validate required fields (message is optional)
	if firstName == "" || lastName == "" || email == "" || company == "" {
		http.Error(w, "Alle velden zijn verplicht behalve het bericht", http.StatusBadRequest)
		return
	}

	// Validate email
	if !strings.Contains(email, "@") {
		http.Error(w, "Geldig e-mailadres vereist", http.StatusBadRequest)
		return
	}

	// Business email validation
	if !isBusinessEmail(email) {
		http.Error(w, getBusinessEmailError(), http.StatusBadRequest)
		return
	}

	// Save to database (if available)
	if db != nil {
		demo := &database.DemoSubmission{
			FirstName:   firstName,
			LastName:    lastName,
			Email:       email,
			Company:     company,
			BuildSystem: buildSystem,
			Message:     message,
			Status:      "new",
		}
		if err := db.CreateDemoSubmission(r.Context(), demo); err != nil {
			log.Printf("Failed to save demo submission: %v", err)
			// Continue anyway - don't block user
		}
	}

	// Send notification email to kim@eliama.agency
	if err := h.sendDemoNotification(firstName, lastName, email, company, buildSystem, message); err != nil {
		log.Printf("Failed to send demo notification email: %v", err)
		// Still send confirmation to user
	}

	// Send confirmation email to submitter
	if err := h.sendDemoConfirmation(email, firstName); err != nil {
		log.Printf("Failed to send demo confirmation email: %v", err)
	}

	// Return success message HTML for HTMX
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<div class="p-4 bg-green-900/30 border border-green-500/30 rounded-lg">
			<p class="text-green-300 font-semibold">✅ Bedankt ` + firstName + `!</p>
			<p class="text-green-200 text-sm mt-2">Je demo aanvraag is ontvangen. We sturen binnen 24 uur een voorstel naar ` + email + `.</p>
		</div>
		<script>
			// Reset form after 5 seconds
			setTimeout(() => {
				document.querySelector('form').reset();
				document.getElementById('form-messages').innerHTML = '';
			}, 5000);
		</script>
	`))
}

// sendDemoNotification sends an email to kim@eliama.agency
func (h *HTTPHandlerV2) sendDemoNotification(firstName, lastName, email, company, buildSystem, message string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465 // SSL/TLS for Netim
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	buildSystemLabels := map[string]string{
		"yocto":     "Yocto (Bitbake)",
		"buildroot": "Buildroot",
		"debian":    "Debian / Ubuntu Core",
		"custom":    "Custom / Anders",
	}
	buildSystemLabel := buildSystemLabels[buildSystem]
	if buildSystemLabel == "" {
		buildSystemLabel = buildSystem
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "kim@eliama.agency")
	m.SetHeader("Subject", fmt.Sprintf("🎯 Nieuwe Demo Aanvraag: %s %s (%s)", firstName, lastName, company))

	// Email body
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #3b82f6 0%%, #1e40af 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .field { margin-bottom: 20px; }
        .label { font-weight: bold; color: #666; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
        .value { margin-top: 5px; font-size: 16px; }
        .message-box { background: white; border-left: 4px solid #3b82f6; padding: 15px; margin-top: 10px; }
        .badge { display: inline-block; background: #3b82f6; color: white; padding: 5px 10px; border-radius: 4px; font-size: 12px; font-weight: bold; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">🎯 Nieuwe Demo Aanvraag</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">LEONA Scanner Demo Request</p>
        </div>
        
        <div class="content">
            <div class="field">
                <div class="label">Contactpersoon</div>
                <div class="value">%s %s</div>
            </div>

            <div class="field">
                <div class="label">E-mailadres</div>
                <div class="value"><a href="mailto:%s">%s</a></div>
            </div>

            <div class="field">
                <div class="label">Bedrijf</div>
                <div class="value">%s</div>
            </div>

            <div class="field">
                <div class="label">Linux Build Systeem</div>
                <div class="value"><span class="badge">%s</span></div>
            </div>

            <div class="field">
                <div class="label">Bericht</div>
                <div class="message-box">%s</div>
            </div>
        </div>

        <div class="footer">
            <p><strong>LEONA & CRAVIT</strong> | Demo Request Notification<br/>
            Deze email is automatisch gegenereerd vanuit <a href="https://leonacompliance.be/demo">leonacompliance.be/demo</a></p>
        </div>
    </div>
</body>
</html>
`, firstName, lastName, email, email, company, buildSystemLabel, message)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true

	return d.DialAndSend(m)
}

// sendDemoConfirmation sends a confirmation email to the submitter
func (h *HTTPHandlerV2) sendDemoConfirmation(to, firstName string) error {
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
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Je demo is onderweg - LEONA & CRAVIT")

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #3b82f6 0%%, #1e40af 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .cta { display: inline-block; background: #3b82f6; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: bold; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">🎯 Demo Aanvraag Bevestiging</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">We sturen je binnen 24 uur een voorstel</p>
        </div>
        
        <div class="content">
            <p>Beste %s,</p>
            
            <p>Super dat je geïnteresseerd bent in LEONA! We hebben je demo aanvraag ontvangen en gaan meteen voor je aan de slag.</p>
            
            <p><strong>Wat gebeurt er nu?</strong></p>
            <ol>
                <li>We bereiden een demo-omgeving voor met jouw Linux build systeem</li>
                <li>Je ontvangt binnen 24 uur een gedetailleerd voorstel</li>
                <li>We plannen een live demo sessie (optioneel)</li>
            </ol>

            <p>Ondertussen kun je alvast kennismaken met onze V-Assessor™ voor een gratis SBOM scan:</p>
            <a href="https://leonacompliance.be" class="cta">Start Gratis Scan</a>

            <p><strong>Vragen?</strong><br/>
            Stuur gerust een email naar <a href="mailto:support@leonacompliance.be">support@leonacompliance.be</a></p>
        </div>

        <div class="footer">
            <p><strong>LEONA & CRAVIT</strong> | CRA Compliance Engineering<br/>
            <a href="https://leonacompliance.be">leonacompliance.be</a> | support@leonacompliance.be</p>
        </div>
    </div>
</body>
</html>
`, firstName)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true

	return d.DialAndSend(m)
}
