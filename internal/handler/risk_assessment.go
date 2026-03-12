package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"leona-scanner/internal/database"

	"gopkg.in/gomail.v2"
)

// RiskAssessmentRequest represents the quiz submission
type RiskAssessmentRequest struct {
	Email                 string `json:"email"`
	CompanyName           string `json:"company_name"`
	SellsToInfrastructure bool   `json:"sells_to_infrastructure"` // Critical infra (telecom, energy)
	UsesOpenSource        bool   `json:"uses_open_source"`        // Open source in kernel/firmware
	HasSBOM               bool   `json:"has_sbom"`                // Current SBOM process
	HasVulnProcess        bool   `json:"has_vuln_process"`        // Vulnerability disclosure
	ProductsInEU          bool   `json:"products_in_eu"`          // Selling in EU market
	CompanySize           string `json:"company_size"`            // "1-10", "11-50", "51-250", "250+"
}

// HandleRiskAssessment processes the 2-minute CRA exposure quiz
func (h *HTTPHandlerV2) HandleRiskAssessment(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Parse form data
	req := RiskAssessmentRequest{
		Email:                 r.FormValue("email"),
		CompanyName:           r.FormValue("company_name"),
		SellsToInfrastructure: r.FormValue("sells_to_infrastructure") == "true",
		UsesOpenSource:        r.FormValue("uses_open_source") == "true",
		HasSBOM:               r.FormValue("has_sbom") == "true",
		HasVulnProcess:        r.FormValue("has_vuln_process") == "true",
		ProductsInEU:          r.FormValue("products_in_eu") == "true",
		CompanySize:           r.FormValue("company_size"),
	}

	// Validate
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		http.Error(w, "Geldig e-mailadres vereist", http.StatusBadRequest)
		return
	}

	// Business email validation (no Gmail, Outlook, etc.)
	if !isBusinessEmail(req.Email) {
		http.Error(w, getBusinessEmailError(), http.StatusBadRequest)
		return
	}

	// Calculate risk score (0-100, higher = more risk)
	riskScore := calculateRiskScore(req)
	riskLevel := getRiskLevel(riskScore)

	// Store lead in database
	if db != nil {
		lead := &database.Lead{
			Email:               req.Email,
			LeadType:            "risk-assessment",
			Source:              "website",
			Status:              "hot", // Risk assessment leads are high-intent
			LeadMagnetRequested: stringPtr("risk-score-pdf"),
		}
		if err := db.CreateLead(r.Context(), lead); err != nil {
			log.Printf("Failed to create lead: %v", err)
		}
	}

	// Send risk score PDF via email
	if err := h.sendRiskScoreEmail(req.Email, req.CompanyName, riskScore, riskLevel); err != nil {
		log.Printf("Failed to send risk score email: %v", err)
		http.Error(w, "Email verzenden mislukt", http.StatusInternalServerError)
		return
	}

	// Send admin notification (hot lead!)
	go h.sendRiskAssessmentNotification(req, riskScore, riskLevel)

	// Count compliance gaps (answered 'no')
	gapCount := 0
	if req.SellsToInfrastructure {
		gapCount++
	}
	if req.UsesOpenSource {
		gapCount++
	}
	if req.HasSBOM {
		gapCount++
	}
	if req.HasVulnProcess {
		gapCount++
	}
	if req.ProductsInEU {
		gapCount++
	}

	// Return Turbo Stream to show thank you page
	w.Header().Set("Content-Type", "text/vnd.turbo-stream.html")
	w.Write([]byte(fmt.Sprintf(`
		<turbo-stream action="replace" target="current-question">
			<template>
				<div class="relative isolate overflow-hidden bg-gradient-to-br from-gray-900 to-blue-950 px-6 py-16 rounded-2xl">
					<div class="absolute inset-0 -z-10 bg-[radial-gradient(45rem_50rem_at_top,theme(colors.blue.600),transparent)] opacity-20"></div>
					<div class="absolute inset-y-0 right-1/2 -z-10 mr-16 w-[200%%%%] origin-bottom-left skew-x-[-30deg] bg-gray-900 shadow-xl shadow-blue-500/10 ring-1 ring-white/10"></div>
					
					<div class="mx-auto max-w-2xl text-center">
						<div class="mx-auto mb-8 h-16 w-16 rounded-full bg-gradient-to-br from-green-400 to-green-600 flex items-center justify-center shadow-lg">
							<svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"/>
							</svg>
						</div>
						
						<h3 class="text-3xl font-bold text-white mb-4">Uw Risk Report is onderweg! 📧</h3>
						<p class="text-lg text-gray-300 mb-2">We sturen binnen <strong class="text-orange-400">2 minuten</strong> uw persoonlijke 42-pagina CRA compliance rapport naar:</p>
						<p class="text-xl font-semibold text-orange-400 mb-6">%s</p>
						<p class="text-sm text-gray-400 mb-8">📊 Uw rapport bevat %d compliance gaps met concrete remediation stappen</p>
						
						<div class="bg-white/5 backdrop-blur-sm border border-white/10 rounded-xl p-6 mb-6">
							<h4 class="text-lg font-bold text-white mb-4">💡 Terwijl u wacht, bekijk ook:</h4>
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
								<button onclick="document.getElementById('risk-assessment-modal').close(); document.getElementById('sample-report-modal').showModal();" class="bg-blue-600 hover:bg-blue-500 text-white font-semibold py-3 px-4 rounded-lg transition-all text-sm">
									📄 Voorbeeld TCF Report
								</button>
								<button onclick="document.getElementById('risk-assessment-modal').close(); document.querySelector('a[href=\\"#assessor\\"]').click();" class="bg-orange-600 hover:bg-orange-500 text-white font-semibold py-3 px-4 rounded-lg transition-all text-sm">
									🔍 Scan Uw SBOM
								</button>
							</div>
						</div>
						
						<button onclick="document.getElementById('risk-assessment-modal').close();" class="text-gray-400 hover:text-white text-sm underline">
							Sluiten
						</button>
					</div>
				</div>
			</template>
		</turbo-stream>
	`, req.Email, gapCount)))
}

// calculateRiskScore determines CRA compliance risk (0-100)
func calculateRiskScore(req RiskAssessmentRequest) int {
	score := 0

	// Critical infrastructure = massive risk
	if req.SellsToInfrastructure {
		score += 30
	}

	// Open source without SBOM = high risk
	if req.UsesOpenSource && !req.HasSBOM {
		score += 25
	}

	// No vulnerability disclosure process
	if !req.HasVulnProcess {
		score += 20
	}

	// Products in EU market
	if req.ProductsInEU {
		score += 15
	}

	// Company size (larger = more liability)
	switch req.CompanySize {
	case "250+":
		score += 10
	case "51-250":
		score += 8
	case "11-50":
		score += 5
	default:
		score += 2
	}

	return score
}

func getRiskLevel(score int) string {
	if score >= 70 {
		return "HOOG"
	} else if score >= 40 {
		return "MIDDEN"
	}
	return "LAAG"
}

// sendRiskScoreEmail sends personalized risk score PDF
func (h *HTTPHandlerV2) sendRiskScoreEmail(to, companyName string, score int, level string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leona-cravit.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", fmt.Sprintf("Uw CRA Exposure Score: %d/100 (%s risico)", score, level))

	// Determine color based on risk level
	color := "#22c55e" // Green (low)
	if level == "MIDDEN" {
		color = "#f59e0b" // Orange
	} else if level == "HOOG" {
		color = "#ef4444" // Red
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #1e293b 0%, #334155 100%); color: white; padding: 30px; border-radius: 8px; }
        .score-box { background: %s; color: white; padding: 40px; text-align: center; border-radius: 12px; margin: 20px 0; }
        .score-number { font-size: 72px; font-weight: bold; line-height: 1; }
        .content { background: #f9fafb; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .risk-item { background: white; padding: 15px; margin: 10px 0; border-left: 4px solid %s; border-radius: 6px; }
        .button { display: inline-block; background: #4f46e5; color: white; padding: 14px 35px; text-decoration: none; border-radius: 6px; margin: 20px 0; font-weight: bold; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0; font-size: 28px;">⚖️ CRA Exposure Assessment</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">%s</p>
        </div>
        
        <div class="score-box">
            <div class="score-number">%d</div>
            <div style="font-size: 24px; margin-top: 10px; font-weight: 600;">%s RISICO</div>
            <p style="margin: 15px 0 0 0; opacity: 0.9; font-size: 14px;">CRA Cyber Resilience Act Compliance Score</p>
        </div>
        
        <div class="content">
            <h2 style="color: #1e293b;">Wat betekent deze score?</h2>
            
            %s
            
            <h3 style="color: #1e293b; margin-top: 30px;">🎯 Aanbevolen acties</h3>
            <div class="risk-item">
                <strong>1. Valideer uw SBOM</strong><br/>
                Upload uw Software Bill of Materials naar onze V-Assessor™ voor een gratis technische gap-analyse.
            </div>
            <div class="risk-item">
                <strong>2. Identificeer CVE's</strong><br/>
                Ontdek welke componenten bekende beveiligingslekken bevatten (NVD database scan).
            </div>
            <div class="risk-item">
                <strong>3. Annex I Mapping</strong><br/>
                Krijg een kant-en-klaar Technical Construction File voor uw notified body.
            </div>

            <a href="https://leona-cravit.be/#assessor" class="button">Start Gratis SBOM Scan</a>
            
            <p style="margin-top: 30px; padding: 20px; background: #fef3c7; border-left: 4px solid #f59e0b; border-radius: 6px;">
                <strong>⏰ Deadline:</strong> CRA wordt verplicht op <strong>11 december 2027</strong>. 
                Producten zonder compliant SBOM mogen niet meer verkocht worden in de EU.
            </p>
        </div>

        <div class="footer">
            <p><strong>LEONA & CRAVIT</strong> | CRA Compliance Engineering<br/>
            <a href="https://leona-cravit.be">leona-cravit.be</a> | <a href="mailto:support@leona-cravit.be">support@leona-cravit.be</a></p>
            <p style="margin-top: 10px; font-size: 11px; color: #999;">
                Deze score is gebaseerd op uw antwoorden en is indicatief. Voor een volledige audit neem contact op.
            </p>
        </div>
    </div>
</body>
</html>
`, color, color, companyName, score, level, getRiskExplanation(score, level))

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	return d.DialAndSend(m)
}

func getRiskExplanation(score int, level string) string {
	if level == "HOOG" {
		return `
            <p style="color: #ef4444; font-weight: 600;">⚠️ Uw organisatie heeft een verhoogd CRA compliance risico.</p>
            <p>Met een score van ` + fmt.Sprintf("%d", score) + `/100 is er een reële kans dat u:</p>
            <ul style="line-height: 2;">
                <li>Persoonlijk aansprakelijk bent onder EU boete-structuur (tot €15M)</li>
                <li>Producten niet mag verkopen zonder Technical Construction File</li>
                <li>Vulnerability disclosure verplichtingen heeft binnen 24 uur</li>
            </ul>
        `
	} else if level == "MIDDEN" {
		return `
            <p style="color: #f59e0b; font-weight: 600;">⚡ Uw organisatie heeft matige CRA risico's.</p>
            <p>Bepaalde aspecten van uw product lifecycle vereisen aandacht:</p>
            <ul style="line-height: 2;">
                <li>SBOM (Software Bill of Materials) proces moet geformaliseerd worden</li>
                <li>Vulnerability tracking vereist een gestructureerde aanpak</li>
                <li>Open Source licentie compliance moet gedocumenteerd worden</li>
            </ul>
        `
	}
	return `
        <p style="color: #22c55e; font-weight: 600;">✅ Uw organisatie heeft relatief lage CRA risico's.</p>
        <p>U bent op de goede weg, maar een formele audit blijft aanbevolen:</p>
        <ul style="line-height: 2;">
            <li>Documenteer uw huidige SBOM proces</li>
            <li>Valideer dat alle componenten CPE identifiers hebben</li>
            <li>Stel een vulnerability disclosure proces op</li>
        </ul>
    `
}

// sendRiskAssessmentNotification notifies admin of hot lead
func (h *HTTPHandlerV2) sendRiskAssessmentNotification(req RiskAssessmentRequest, score int, level string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leona-cravit.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "kim@eliama.agency")
	m.SetHeader("Subject", fmt.Sprintf("🔥 HOT LEAD: %s - Risk Score %d (%s)", req.Email, score, level))

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: system-ui, sans-serif;">
    <h2>🔥 Hete Lead via Risk Assessment</h2>
    <table style="background: #f3f4f6; padding: 20px; border-radius: 8px;">
        <tr><td><strong>Email:</strong></td><td>%s</td></tr>
        <tr><td><strong>Bedrijf:</strong></td><td>%s</td></tr>
        <tr><td><strong>Risk Score:</strong></td><td>%d/100 (%s)</td></tr>
        <tr><td><strong>Grootte:</strong></td><td>%s werknemers</td></tr>
        <tr><td><strong>Kritieke Infra:</strong></td><td>%s</td></tr>
        <tr><td><strong>Heeft SBOM:</strong></td><td>%s</td></tr>
        <tr><td><strong>Timestamp:</strong></td><td>%s</td></tr>
    </table>
    
    <h3>🎯 Follow-up Actie</h3>
    <p>Deze lead heeft een <strong>%s risicoprofiel</strong> - stuur binnen 24 uur een gepersonaliseerde follow-up.</p>
</body>
</html>
`, req.Email, req.CompanyName, score, level, req.CompanySize,
		boolToJaNee(req.SellsToInfrastructure),
		boolToJaNee(req.HasSBOM),
		time.Now().Format("2006-01-02 15:04:05"),
		level)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send risk assessment notification: %v", err)
	} else {
		log.Printf("✅ Hot lead notification sent: %s (score: %d/%s)", req.Email, score, level)
	}
}

func boolToJaNee(b bool) string {
	if b {
		return "JA"
	}
	return "NEE"
}
