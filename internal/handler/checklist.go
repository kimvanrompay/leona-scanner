package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"leona-scanner/internal/database"

	"gopkg.in/gomail.v2"
)

// HandleChecklistDownload handles all checklist downloads with admin notification
func (h *HTTPHandlerV2) HandleChecklistDownload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	checklistType := r.FormValue("checklist_type")

	// Validate input
	if email == "" || !strings.Contains(email, "@") {
		http.Error(w, "Geldig e-mailadres vereist", http.StatusBadRequest)
		return
	}

	// Business email validation
	if !isBusinessEmail(email) {
		http.Error(w, getBusinessEmailError(), http.StatusBadRequest)
		return
	}

	if checklistType == "" {
		http.Error(w, "Checklist type vereist", http.StatusBadRequest)
		return
	}

	// Determine lead type based on checklist
	leadType := "engineer"
	if checklistType == "annex-i-template" || checklistType == "nis2-cra-cross" {
		leadType = "lawyer"
	}

	// Create lead in database
	if db != nil {
		lead := &database.Lead{
			Email:               email,
			LeadType:            leadType,
			Source:              "checklist-page",
			Status:              "new",
			LeadMagnetRequested: stringPtr(checklistType),
		}
		if err := db.CreateLead(r.Context(), lead); err != nil {
			log.Printf("Failed to create lead: %v", err)
		}
	}

	// Send checklist email to user
	if err := h.sendChecklistEmail(email, checklistType); err != nil {
		log.Printf("Failed to send checklist email: %v", err)
		http.Error(w, "Email verzenden mislukt", http.StatusInternalServerError)
		return
	}

	// Send admin notification to kim@eliama.agency (async)
	go h.sendAdminNotification(email, checklistType)

	// Return success response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<p class="text-sm text-green-600 bg-green-50 px-4 py-2 rounded-lg border border-green-200">✅ Check je inbox! We hebben de checklist naar ` + email + ` gestuurd.</p>`))
}

// sendChecklistEmail sends the requested checklist to the user
func (h *HTTPHandlerV2) sendChecklistEmail(to, checklistType string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	// Map checklist type to friendly name and description
	checklistInfo := map[string]struct {
		Name        string
		Description string
		FileName    string
	}{
		"yocto-sbom":       {"Yocto SBOM Readiness Checklist", "22-punt checklist voor CycloneDX SBOM generatie", "Yocto_SBOM_Checklist.pdf"},
		"annex-i-template": {"CRA Annex I Audit Template", "Excel template met volledige Annex I mapping", "CRA_Annex_I_Template.xlsx"},
		"nis2-cra-cross":   {"NIS2 × CRA Cross-Compliance", "Vermijd dubbel werk met deze overlap-analyse", "NIS2_CRA_Cross_Compliance.pdf"},
		"kernel-eol":       {"Linux Kernel EOL Validator", "LTS lifecycle calculator voor kernel support", "Kernel_EOL_Validator.xlsx"},
		"busybox-audit":    {"BusyBox Security Audit Recipe", "Bitbake bbclass voor BusyBox security checks", "busybox-audit.bbclass"},
		"gpl-scanner":      {"GPL License Risk Scanner", "Python script voor copyleft licentie detectie", "gpl_scanner.py"},
	}

	info, exists := checklistInfo[checklistType]
	if !exists {
		info = checklistInfo["yocto-sbom"] // Default fallback
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", fmt.Sprintf("Jouw %s", info.Name))

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #4f46e5 0%%, #6366f1 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9fafb; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .button { display: inline-block; background: #4f46e5; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; margin: 20px 0; font-weight: bold; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">📋 %s</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">%s</p>
        </div>
        
        <div class="content">
            <p>Bedankt voor je interesse in LEONA compliance tools!</p>
            
            <p>Hieronder vind je de download link voor: <strong>%s</strong></p>

            <a href="https://leonacompliance.be/downloads/%s" class="button">Download %s</a>

            <h3>🎯 Volgende stap</h3>
            <p>Wil je een volledige CRA gap analysis van je product?</p>
            <a href="https://leonacompliance.be/#assessor" class="button">Upload je SBOM (gratis)</a>
        </div>

        <div class="footer">
            <p><strong>LEONA</strong> | CRA Compliance Engineering<br/>
            Vragen? Reply op deze email of bezoek <a href="https://leonacompliance.be">leonacompliance.be</a></p>
        </div>
    </div>
</body>
</html>
`, info.Name, info.Description, info.Name, info.FileName, info.Name)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	return d.DialAndSend(m)
}

// sendAdminNotification sends email to kim@eliama.agency for every lead
func (h *HTTPHandlerV2) sendAdminNotification(email, checklistType string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		log.Println("SMTP not configured, skipping admin notification")
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "kim@eliama.agency")
	m.SetHeader("Subject", fmt.Sprintf("🎯 Nieuwe Lead: %s - %s", email, checklistType))

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #16a34a 0%%, #22c55e 100%%); color: white; padding: 20px; border-radius: 8px; }
        .content { background: #f0fdf4; padding: 20px; margin-top: 20px; border-radius: 8px; border-left: 4px solid #16a34a; }
        .data { background: white; padding: 15px; border-radius: 6px; margin: 15px 0; }
        .label { font-size: 12px; color: #666; text-transform: uppercase; font-weight: bold; }
        .value { font-size: 16px; color: #111; margin-top: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2 style="margin: 0;">🎯 Nieuwe Lead!</h2>
            <p style="margin: 5px 0 0 0; opacity: 0.9; font-size: 14px;">LEONA Checklist Download</p>
        </div>
        
        <div class="content">
            <div class="data">
                <div class="label">Email</div>
                <div class="value">%s</div>
            </div>

            <div class="data">
                <div class="label">Aangevraagde Checklist</div>
                <div class="value">%s</div>
            </div>

            <div class="data">
                <div class="label">Lead Type</div>
                <div class="value">%s</div>
            </div>

            <div class="data">
                <div class="label">Timestamp</div>
                <div class="value">%s</div>
            </div>

            <div class="data">
                <div class="label">Acties</div>
                <div class="value">
                    ✅ Email verstuurd naar lead<br/>
                    ✅ Opgeslagen in Supabase<br/>
                    📊 Bekijk in dashboard: <a href="https://supabase.com">Supabase</a>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
`, email, checklistType, getLeadType(checklistType), time.Now().Format("2006-01-02 15:04:05"))

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send admin notification: %v", err)
	} else {
		log.Printf("✅ Admin notification sent to kim@eliama.agency for lead: %s (%s)", email, checklistType)
	}
}

func getLeadType(checklistType string) string {
	if checklistType == "annex-i-template" || checklistType == "nis2-cra-cross" {
		return "Lawyer / Compliance Team"
	}
	return "Engineer / Technical Team"
}

// HandleChecklistPage serves the checklist landing page
func (h *HTTPHandlerV2) HandleChecklistPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/lead-magnets.html")
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
