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

// HandleContactSubmit processes contact form submissions
func (h *HTTPHandlerV2) HandleContactSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Extract form data
	firstName := strings.TrimSpace(r.FormValue("first-name"))
	lastName := strings.TrimSpace(r.FormValue("last-name"))
	email := strings.TrimSpace(r.FormValue("email"))
	company := strings.TrimSpace(r.FormValue("company"))
	message := strings.TrimSpace(r.FormValue("message"))
	solution := r.FormValue("solution")

	// Validate required fields
	if firstName == "" || lastName == "" || email == "" || company == "" || message == "" {
		http.Error(w, "Alle velden zijn verplicht", http.StatusBadRequest)
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
		contact := &database.ContactSubmission{
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Company:   company,
			Message:   message,
			Solution:  solution,
			Status:    "new",
		}
		if err := db.CreateContactSubmission(r.Context(), contact); err != nil {
			log.Printf("Failed to save contact submission: %v", err)
			// Continue anyway - don't block user
		}
	}

	// Send notification email to kim@eliama.agency
	if err := h.sendContactNotification(firstName, lastName, email, company, message, solution); err != nil {
		log.Printf("Failed to send notification email: %v", err)
		// Still send confirmation to user
	}

	// Send confirmation email to submitter
	if err := h.sendContactConfirmation(email, firstName); err != nil {
		log.Printf("Failed to send confirmation email: %v", err)
	}

	// Return success message HTML for HTMX
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<div class="p-4 bg-green-900/30 border border-green-500/30 rounded-lg">
			<p class="text-green-300 font-semibold">✅ Bedankt ` + firstName + `!</p>
			<p class="text-green-200 text-sm mt-2">Je aanvraag is ontvangen. We nemen binnen 24 uur contact op via ` + email + `.</p>
		</div>
		<script>
			// Reset form after 3 seconds
			setTimeout(() => {
				document.querySelector('form').reset();
				document.getElementById('form-messages').innerHTML = '';
			}, 5000);
		</script>
	`))
}

// sendContactNotification sends an email to kim@eliama.agency
func (h *HTTPHandlerV2) sendContactNotification(firstName, lastName, email, company, message, solution string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465 // SSL/TLS for Netim
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	solutionLabels := map[string]string{
		"snapshot": "Snapshot Audit (€999)",
		"shield":   "LEONA Shield (€2.499)",
		"pipeline": "Compliance Pipeline (Vanaf €499/m)",
	}
	solutionLabel := solutionLabels[solution]
	if solutionLabel == "" {
		solutionLabel = solution
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "kim@eliama.agency")
	m.SetHeader("Subject", fmt.Sprintf("🔔 Nieuw Contact: %s %s (%s)", firstName, lastName, company))

	// Email body
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .field { margin-bottom: 20px; }
        .label { font-weight: bold; color: #666; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
        .value { margin-top: 5px; font-size: 16px; }
        .message-box { background: white; border-left: 4px solid #667eea; padding: 15px; margin-top: 10px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">🔔 Nieuw Contact Formulier</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">Audit-Aanvraag van website</p>
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
                <div class="label">Gewenste Oplossing</div>
                <div class="value">%s</div>
            </div>

            <div class="field">
                <div class="label">Bericht</div>
                <div class="message-box">%s</div>
            </div>
        </div>

        <div class="footer">
			<p><strong>LEONA Compliance</strong> | Contact Form Notification<br/>
			 Deze email is automatisch gegenereerd vanuit <a href="https://leonacompliance.be/contact">leonacompliance.be/contact</a></p> `+
		`
        </div>
    </div>
</body>
</html>
`, firstName, lastName, email, email, company, solutionLabel, message)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true

	return d.DialAndSend(m)
}

// sendContactConfirmation sends a confirmation email to the submitter
func (h *HTTPHandlerV2) sendContactConfirmation(to, firstName string) error {
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
	m.SetHeader("Subject", "Bedankt voor je aanvraag - LEONA")

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">✅ Aanvraag Ontvangen</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">We nemen zo snel mogelijk contact op</p>
        </div>
        
        <div class="content">
            <p>Beste %s,</p>
            
            <p>Bedankt voor je interesse in LEONA. We hebben je aanvraag goed ontvangen en een van onze compliance-experts neemt binnen 24 uur contact met je op.</p>
            
            <p><strong>Wat gebeurt er nu?</strong></p>
            <ol>
                <li>We bekijken je specifieke situatie</li>
                <li>We stellen een passende oplossing voor</li>
                <li>Je ontvangt een offerte op maat</li>
            </ol>

            <p>In de tussentijd kun je alvast een gratis scan doen van je SBOM via onze `+
		`<a href="https://leonacompliance.be">V-Assessor™</a>.</p>
        </div>

        <div class="footer">
            <p><strong>LEONA</strong> | CRA Compliance Engineering<br/>
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
