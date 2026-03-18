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
	jobTitle := strings.TrimSpace(r.FormValue("job-title"))
	companySize := r.FormValue("company-size")
	country := r.FormValue("country")
	phone := strings.TrimSpace(r.FormValue("phone"))
	marketingConsent := r.FormValue("marketing-consent")

	// Validate required fields
	if firstName == "" || lastName == "" || email == "" || company == "" || jobTitle == "" {
		http.Error(w, "Alle verplichte velden moeten ingevuld zijn", http.StatusBadRequest)
		return
	}

	// Validate email
	if !strings.Contains(email, "@") {
		http.Error(w, "Geldig e-mailadres vereist", http.StatusBadRequest)
		return
	}

	// Save to database (if available)
	if db != nil {
		demo := &database.ContactSubmission{
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Company:   company,
			Message:   fmt.Sprintf("Job Title: %s | Company Size: %s | Country: %s | Phone: %s", jobTitle, companySize, country, phone),
			Solution:  "demo-request",
			Status:    "new",
		}
		if err := db.CreateContactSubmission(r.Context(), demo); err != nil {
			log.Printf("Failed to save demo submission: %v", err)
			// Continue anyway - don't block user
		}
	}

	// Send notification email to support@leonacompliance.be
	if err := h.sendDemoNotification(firstName, lastName, email, company, jobTitle, companySize, country, phone, marketingConsent); err != nil {
		log.Printf("❌ ERROR: Failed to send demo notification to kim@leonacompliance.be: %v", err)
		// Still send confirmation to user
	} else {
		log.Printf("✅ SUCCESS: Demo request notification sent to kim@leonacompliance.be from %s %s (%s)", firstName, lastName, email)
	}

	// Send confirmation email to submitter
	if err := h.sendDemoConfirmation(email, firstName); err != nil {
		log.Printf("⚠️  WARNING: Failed to send demo confirmation email to %s: %v", email, err)
	} else {
		log.Printf("📧 Demo confirmation email sent to %s", email)
	}

	// Return success message HTML for HTMX
	w.Header().Set("Content-Type", "text/html")
	//nolint:misspell // "informatie" is correct Dutch
	w.Write([]byte(`
		<div class="text-center py-12">
			<div class="mb-6">
				<svg class="w-20 h-20 mx-auto text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
				</svg>
			</div>
			<h2 class="text-3xl font-bold text-gray-900 mb-4">Bedankt, ` + firstName + `!</h2>
			<p class="text-xl text-gray-600 mb-2">Je demo-aanvraag is ontvangen.</p>
			<p class="text-gray-600 mb-4">We contacteren je zo snel mogelijk om je persoonlijke demo in te plannen.</p>
			<div class="bg-blue-50 border border-blue-200 rounded-lg p-6 max-w-md mx-auto">
				<p class="text-blue-900 font-semibold mb-2">Check je mailbox</p>
				<p class="text-blue-700 text-sm">Je ontvangt zo een bevestigingsmail met meer informatie.</p>
			</div>
			<a href="https://leonacompliance.be" class="inline-block mt-8 bg-blue-900 hover:bg-blue-800 text-white font-semibold px-8 py-3 rounded-lg transition-colors">
				Terug naar Home
			</a>
		</div>
		<script src="https://cdn.jsdelivr.net/npm/canvas-confetti@1.6.0/dist/confetti.browser.min.js"></script>
		<script>
			// Confetti celebration
			confetti({
				particleCount: 100,
				spread: 70,
				origin: { y: 0.6 }
			});
		</script>
	`))
}

// sendDemoNotification sends an email to kim@leonacompliance.be
//
//nolint:funlen // Email template functions are naturally longer
func (h *HTTPHandlerV2) sendDemoNotification(firstName, lastName, email, company, jobTitle, companySize, country, phone, marketingConsent string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465 // SSL/TLS for Netim
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	consent := "Nee"
	if marketingConsent == "yes" {
		consent = "Ja"
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "support@leonacompliance.be")
	m.SetHeader("Cc", "kim@leonacompliance.be") // CC Kim for visibility
	m.SetHeader("Subject", fmt.Sprintf("🎬 DEMO Aanvraag: %s %s (%s)", firstName, lastName, company))

	// Email body
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #1e3a8a 0%%, #1e40af 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .field { margin-bottom: 20px; }
        .label { font-weight: bold; color: #666; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
        .value { margin-top: 5px; font-size: 16px; }
        .highlight { background: white; border-left: 4px solid #FF6B35; padding: 15px; margin-top: 10px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">🎬 DEMO Aanvraag</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">Nieuwe demo-aanvraag van website</p>
        </div>
        
        <div class="content">
            <div class="field">
                <div class="label">Contactpersoon</div>
                <div class="value">%s %s</div>
            </div>

            <div class="field">
                <div class="label">Functie</div>
                <div class="value">%s</div>
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
                <div class="label">Bedrijfsgrootte</div>
                <div class="value">%s</div>
            </div>

            <div class="field">
                <div class="label">Land</div>
                <div class="value">%s</div>
            </div>

            <div class="field">
                <div class="label">Telefoonnummer</div>
                <div class="value">%s</div>
            </div>

            <div class="field">
                <div class="label">Marketing Toestemming</div>
                <div class="value">%s</div>
            </div>
        </div>

        <div class="footer">
			<p><strong>LEONA Compliance</strong> | Demo Request Notification<br/>
			Deze email is automatisch gegenereerd vanuit <a href="https://leonacompliance.be/demo">leonacompliance.be/demo</a></p>
        </div>
    </div>
</body>
</html>
`, firstName, lastName, jobTitle, email, email, company, companySize, country, phone, consent)

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
	m.SetHeader("Subject", "Je LEONA demo is onderweg 🎬")

	//nolint:misspell // "Executie" is correct Dutch word meaning "execution"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="nl" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;800&display=swap" rel="stylesheet">
    <style>
        @media only screen and (max-width: 600px) {
            .container { width: 100%% !important; }
            .content { padding: 40px 20px !important; }
            h1 { font-size: 28px !important; }
        }
    </style>
</head>
<body style="margin: 0; padding: 0; background-color: #ffffff; font-family: 'Inter', -apple-system, sans-serif; color: #000000; -webkit-font-smoothing: antialiased;">
    <div align="center" style="background-color: #ffffff;">
        <div class="container" style="max-width: 600px; margin: 0 auto; text-align: left;">
            
            <img src="https://res.cloudinary.com/dg0qxqj4a/image/upload/v1773871167/CRA_COMPLIANT_LINUX_SYSTEM-5_shs5we.png" 
                 alt="CRA Compliance Platform" 
                 style="width: 100%%; max-width: 600px; height: auto; display: block; border: 0;">

            <div class="content" style="padding: 60px 40px;">
                <p style="font-size: 14px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.1em; color: #fd7e14; margin: 0 0 12px 0;">Volgende Stap: Actie</p>
                <h1 style="font-size: 36px; font-weight: 800; letter-spacing: -0.03em; line-height: 1.1; margin: 0 0 32px 0;">Laten we schakelen, %s.</h1>
                
                <p style="font-size: 18px; line-height: 1.6; margin: 0 0 24px 0; color: #000000;">
                    Luister, de CRA komt eraan en de tijd tikt. Je hebt een demo aangevraagd omdat je resultaat wilt, geen verkooppraatje. Dat begrijpen we.
                </p>

                <p style="font-size: 18px; line-height: 1.6; margin: 0 0 40px 0; color: #000000;">
                    <strong>Een van onze engineers belt je binnen 24 uur.</strong> Geen ruis. Gewoon een direct gesprek om je demo in te plannen en te kijken of LEONA past bij jouw stack.
                </p>

                <div style="margin-bottom: 48px;">
                    <p style="font-size: 14px; font-weight: 700; text-transform: uppercase; color: #888888; margin-bottom: 20px; border-bottom: 1px solid #eeeeee; padding-bottom: 8px;">De blauwdruk</p>
                    <div style="font-size: 16px; margin-bottom: 12px; font-weight: 600;">01. Live Platform Walkthrough</div>
                    <div style="font-size: 16px; margin-bottom: 12px; font-weight: 600;">02. Jouw specifieke CRA-uitdagingen</div>
                    <div style="font-size: 16px; margin-bottom: 12px; font-weight: 600;">03. Geautomatiseerde SBOM Executie</div>
                    <div style="font-size: 16px; margin-bottom: 12px; font-weight: 600;">04. Technische Q&A</div>
                </div>

                <a href="https://leonacompliance.be/cra-compliance" 
                   style="display: inline-block; background-color: #003366; color: #ffffff; padding: 18px 32px; font-size: 14px; font-weight: 700; text-decoration: none; border-radius: 4px; text-transform: uppercase; letter-spacing: 0.05em;">
                   Bekijk de CRA Roadmap
                </a>
            </div>

            <div class="footer" style="padding: 0 40px 60px 40px; font-size: 12px; color: #888888; line-height: 1.8;">
                <p style="margin: 0;">
                    <strong>LEONA Compliance</strong><br>
                    Compliance as Code<br>
                    <a href="https://leonacompliance.be" style="color: #888888; text-decoration: underline;">leonacompliance.be</a> | <a href="mailto:kim@leonacompliance.be" style="color: #888888; text-decoration: underline;">support@leonacompliance.be</a>
                </p>
                <p style="margin: 20px 0 0 0;">
                    Je ontvangt deze mail omdat je een demo hebt aangevraagd. Geen spam. Alleen executie.
                </p>
            </div>
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
