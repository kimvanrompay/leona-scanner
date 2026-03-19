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

// HandlePartnerMeetingSubmit processes partner meeting form submissions
func (h *HTTPHandlerV2) HandlePartnerMeetingSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Extract form data
	firstName := strings.TrimSpace(r.FormValue("first-name"))
	lastName := strings.TrimSpace(r.FormValue("last-name"))
	email := strings.TrimSpace(r.FormValue("email"))
	lawFirm := strings.TrimSpace(r.FormValue("law-firm"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	specializations := r.Form["specialization"] // Can be multiple
	partnershipModel := r.FormValue("partnership-model")
	clientPortfolio := strings.TrimSpace(r.FormValue("client-portfolio"))
	preferredDate := strings.TrimSpace(r.FormValue("preferred-date"))

	// Validate required fields
	if firstName == "" || lastName == "" || email == "" || lawFirm == "" {
		http.Error(w, "Voornaam, achternaam, email en advocatenkantoor zijn verplicht", http.StatusBadRequest)
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
		notes := "Legal Partnership - " + partnershipModel
		lead := &database.Lead{
			Email:       email,
			FirstName:   &firstName,
			LastName:    &lastName,
			CompanyName: &lawFirm,
			Phone:       &phone,
			LeadType:    "partner-meeting",
			Source:      "website",
			Status:      "new",
			Notes:       &notes,
		}
		if err := db.CreateLead(r.Context(), lead); err != nil {
			log.Printf("Failed to save partner meeting lead: %v", err)
			// Continue anyway - don't block user
		}
	}

	// Send notification email to kim@leonacompliance.be
	if err := h.sendPartnerMeetingNotification(firstName, lastName, email, lawFirm, phone, specializations, partnershipModel, clientPortfolio, preferredDate); err != nil {
		log.Printf("❌ ERROR: Failed to send partner meeting notification to kim@leonacompliance.be: %v", err)
		// Still send confirmation to user
	} else {
		log.Printf("✅ SUCCESS: Partner meeting request from %s %s (%s) sent to kim@leonacompliance.be", firstName, lastName, lawFirm)
	}

	// Send confirmation email to submitter
	if err := h.sendPartnerMeetingConfirmation(email, firstName, lawFirm); err != nil {
		log.Printf("⚠️  WARNING: Failed to send partner meeting confirmation to %s: %v", email, err)
	} else {
		log.Printf("📧 Partner meeting confirmation sent to %s", email)
	}

	// Return success message HTML for HTMX
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<div class="p-4 bg-green-900/30 border border-green-500/30 rounded-lg">
			<p class="text-green-300 font-semibold">✅ Uitstekend, ` + firstName + `!</p>
			<p class="text-green-200 text-sm mt-2">Uw partner-overleg aanvraag is ontvangen. We nemen binnen 24 uur contact op via ` + email + ` om een strategisch overleg in te plannen.</p>
			<p class="text-green-200 text-xs mt-2">We kijken uit naar de samenwerking met ` + lawFirm + `.</p>
		</div>
		<script>
			// Reset form after 5 seconds
			setTimeout(() => {
				const form = document.getElementById('partner-meeting-form');
				if (form) form.reset();
				document.getElementById('form-messages').innerHTML = '';
			}, 5000);
		</script>
	`))
}

// sendPartnerMeetingNotification sends an email to kim@leonacompliance.be
func (h *HTTPHandlerV2) sendPartnerMeetingNotification(firstName, lastName, email, lawFirm, phone string, specializations []string, partnershipModel, clientPortfolio, preferredDate string) error {
	// Use Mailgun if configured, fallback to SMTP
	if h.mailgunService != nil {
		return h.sendPartnerMeetingNotificationViaMailgun(firstName, lastName, email, lawFirm, phone, specializations, partnershipModel, clientPortfolio, preferredDate)
	}

	// Fallback to SMTP
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("email service not configured")
	}

	// Format specializations
	specializationLabels := map[string]string{
		"ip-ict":            "IP / ICT & Privacy Recht",
		"product-liability": "Product Liability & Ondernemingsrecht",
		"other":             "Ander Specialisme",
	}
	specList := ""
	for _, spec := range specializations {
		if label, ok := specializationLabels[spec]; ok {
			specList += "• " + label + "<br/>"
		}
	}
	if specList == "" {
		specList = "Niet opgegeven"
	}

	// Format partnership model
	modelLabels := map[string]string{
		"integrated": "Integrated Partnership (Bulk licenties, 30-40% markup)",
		"referral":   "Referral Model (15-20% commissie)",
		"explore":    "Verkennen (Meer informatie gewenst)",
	}
	modelLabel := modelLabels[partnershipModel]
	if modelLabel == "" {
		modelLabel = partnershipModel
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "kim@leonacompliance.be")
	m.SetHeader("Subject", fmt.Sprintf("🤝 Partner Overleg Aanvraag: %s %s (%s)", firstName, lastName, lawFirm))

	// Email body
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #003366 0%%, #1e40af 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .field { margin-bottom: 20px; }
        .label { font-weight: bold; color: #666; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
        .value { margin-top: 5px; font-size: 16px; }
        .highlight { background: white; border-left: 4px solid #003366; padding: 15px; margin-top: 10px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">🤝 Partner Overleg Aanvraag</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">Legal Partnership Request</p>
        </div>
        
        <div class="content">
            <div class="field">
                <div class="label">Contactpersoon</div>
                <div class="value">%s %s</div>
            </div>

            <div class="field">
                <div class="label">Advocatenkantoor</div>
                <div class="value"><strong>%s</strong></div>
            </div>

            <div class="field">
                <div class="label">E-mailadres</div>
                <div class="value"><a href="mailto:%s">%s</a></div>
            </div>

            <div class="field">
                <div class="label">Telefoonnummer</div>
                <div class="value">%s</div>
            </div>

            <div class="field">
                <div class="label">Specialisaties</div>
                <div class="value">%s</div>
            </div>

            <div class="field">
                <div class="label">Partnership Model Interesse</div>
                <div class="value"><strong>%s</strong></div>
            </div>

            <div class="field">
                <div class="label">Cliëntbestand Context</div>
                <div class="highlight">%s</div>
            </div>

            <div class="field">
                <div class="label">Gewenste Week</div>
                <div class="value">%s</div>
            </div>
        </div>

        <div class="footer">
			<p><strong>LEONA Compliance</strong> | Partner Meeting Request<br/>
			Deze email is automatisch gegenereerd vanuit <a href="https://leonacompliance.be/partner-overleg">leonacompliance.be/partner-overleg</a></p>
        </div>
    </div>
</body>
</html>
`, firstName, lastName, lawFirm, email, email, phone, specList, modelLabel, clientPortfolio, preferredDate)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true

	return d.DialAndSend(m)
}

// sendPartnerMeetingConfirmation sends a confirmation email to the submitter
func (h *HTTPHandlerV2) sendPartnerMeetingConfirmation(to, firstName, lawFirm string) error {
	// Use Mailgun if configured, fallback to SMTP
	if h.mailgunService != nil {
		return h.sendPartnerMeetingConfirmationViaMailgun(to, firstName, lawFirm)
	}

	// Fallback to SMTP
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
	m.SetHeader("Subject", "Partner Overleg Bevestiging - LEONA")

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #003366 0%%, #1e40af 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">✅ Partner Overleg Ingepland</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">We nemen binnen 24 uur contact op</p>
        </div>
        
        <div class="content">
            <p>Beste %s,</p>
            
            <p>Bedankt voor uw interesse in een juridische samenwerking met LEONA. We hebben uw aanvraag voor een partner-overleg ontvangen en waarderen het vertrouwen van %s.</p>
            
            <p><strong>Wat gebeurt er nu?</strong></p>
            <ol>
                <li>We nemen binnen 24 uur contact op om een 30-minuten overleg in te plannen</li>
                <li>We bespreken uw praktijk, cliëntbestand en revenue-sharing mogelijkheden</li>
                <li>We demonstreren live hoe de V-Assessor™ werkt</li>
                <li>U krijgt een partnerschapsvoorstel op maat</li>
            </ol>

            <p><strong>Voorbereiding:</strong><br/>
            Om het overleg zo efficiënt mogelijk te maken, kunt u alvast nadenken over:</p>
            <ul>
                <li>Hoeveel tech-cliënten u momenteel adviseert</li>
                <li>Welk percentage IoT/embedded systemen gebruikt</li>
                <li>Of u interesse heeft in Integrated of Referral partnership</li>
            </ul>

            <p>We kijken ernaar uit om de mogelijkheden te verkennen!</p>

            <p>Met vriendelijke groet,<br/>
            <strong>Kim Van Rompay</strong><br/>
            Founder & Technical Lead<br/>
            LEONA Compliance</p>
        </div>

        <div class="footer">
            <p><strong>LEONA</strong> | Legal Partnership Program<br/>
            <a href="https://leonacompliance.be">leonacompliance.be</a> | kim@leonacompliance.be</p>
        </div>
    </div>
</body>
</html>
`, firstName, lawFirm)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true

	return d.DialAndSend(m)
}

// sendPartnerMeetingNotificationViaMailgun sends notification via Mailgun
func (h *HTTPHandlerV2) sendPartnerMeetingNotificationViaMailgun(firstName, lastName, email, lawFirm, phone string, specializations []string, partnershipModel, clientPortfolio, preferredDate string) error {
	// Format specializations
	specializationLabels := map[string]string{
		"ip-ict":            "IP / ICT & Privacy Recht",
		"product-liability": "Product Liability & Ondernemingsrecht",
		"other":             "Ander Specialisme",
	}
	specList := ""
	for _, spec := range specializations {
		if label, ok := specializationLabels[spec]; ok {
			specList += "• " + label + "<br/>"
		}
	}
	if specList == "" {
		specList = "Niet opgegeven"
	}

	// Format partnership model
	modelLabels := map[string]string{
		"integrated": "Integrated Partnership (Bulk licenties, 30-40%%%% markup)",
		"referral":   "Referral Model (15-20%%%% commissie)",
		"explore":    "Verkennen (Meer informatie gewenst)",
	}
	modelLabel := modelLabels[partnershipModel]
	if modelLabel == "" {
		modelLabel = partnershipModel
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #003366 0%%%%, #1e40af 100%%%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .field { margin-bottom: 20px; }
        .label { font-weight: bold; color: #666; font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
        .value { margin-top: 5px; font-size: 16px; }
        .highlight { background: white; border-left: 4px solid #003366; padding: 15px; margin-top: 10px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">🤝 Partner Overleg Aanvraag</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">Legal Partnership Request</p>
        </div>
        
        <div class="content">
            <div class="field">
                <div class="label">Contactpersoon</div>
                <div class="value">%s %s</div>
            </div>

            <div class="field">
                <div class="label">Advocatenkantoor</div>
                <div class="value"><strong>%s</strong></div>
            </div>

            <div class="field">
                <div class="label">E-mailadres</div>
                <div class="value"><a href="mailto:%s">%s</a></div>
            </div>

            <div class="field">
                <div class="label">Telefoonnummer</div>
                <div class="value">%s</div>
            </div>

            <div class="field">
                <div class="label">Specialisaties</div>
                <div class="value">%s</div>
            </div>

            <div class="field">
                <div class="label">Partnership Model Interesse</div>
                <div class="value"><strong>%s</strong></div>
            </div>

            <div class="field">
                <div class="label">Cliëntbestand Context</div>
                <div class="highlight">%s</div>
            </div>

            <div class="field">
                <div class="label">Gewenste Week</div>
                <div class="value">%s</div>
            </div>
        </div>

        <div class="footer">
			<p><strong>LEONA Compliance</strong> | Partner Meeting Request<br/>
			Deze email is automatisch gegenereerd vanuit <a href="https://leonacompliance.be/partner-overleg">leonacompliance.be/partner-overleg</a></p>
        </div>
    </div>
</body>
</html>
`, firstName, lastName, lawFirm, email, email, phone, specList, modelLabel, clientPortfolio, preferredDate)

	subject := fmt.Sprintf("🤝 Partner Overleg Aanvraag: %s %s (%s)", firstName, lastName, lawFirm)
	return h.mailgunService.SendHTMLEmail("kim@leonacompliance.be", subject, body)
}

// sendPartnerMeetingConfirmationViaMailgun sends confirmation via Mailgun
func (h *HTTPHandlerV2) sendPartnerMeetingConfirmationViaMailgun(to, firstName, lawFirm string) error {
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #003366 0%%%%, #1e40af 100%%%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">✅ Partner Overleg Ingepland</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">We nemen binnen 24 uur contact op</p>
        </div>
        
        <div class="content">
            <p>Beste %s,</p>
            
            <p>Bedankt voor uw interesse in een juridische samenwerking met LEONA. We hebben uw aanvraag voor een partner-overleg ontvangen en waarderen het vertrouwen van %s.</p>
            
            <p><strong>Wat gebeurt er nu?</strong></p>
            <ol>
                <li>We nemen binnen 24 uur contact op om een 30-minuten overleg in te plannen</li>
                <li>We bespreken uw praktijk, cliëntbestand en revenue-sharing mogelijkheden</li>
                <li>We demonstreren live hoe de V-Assessor™ werkt</li>
                <li>U krijgt een partnerschapsvoorstel op maat</li>
            </ol>

            <p><strong>Voorbereiding:</strong><br/>
            Om het overleg zo efficiënt mogelijk te maken, kunt u alvast nadenken over:</p>
            <ul>
                <li>Hoeveel tech-cliënten u momenteel adviseert</li>
                <li>Welk percentage IoT/embedded systemen gebruikt</li>
                <li>Of u interesse heeft in Integrated of Referral partnership</li>
            </ul>

            <p>We kijken ernaar uit om de mogelijkheden te verkennen!</p>

            <p>Met vriendelijke groet,<br/>
            <strong>Kim Van Rompay</strong><br/>
            Founder & Technical Lead<br/>
            LEONA Compliance</p>
        </div>

        <div class="footer">
            <p><strong>LEONA</strong> | Legal Partnership Program<br/>
            <a href="https://leonacompliance.be">leonacompliance.be</a> | kim@leonacompliance.be</p>
        </div>
    </div>
</body>
</html>
`, firstName, lawFirm)

	return h.mailgunService.SendHTMLEmail(to, "Partner Overleg Bevestiging - LEONA", body)
}
