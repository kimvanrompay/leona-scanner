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

// HandleSampleReportDownload handles the "Download Voorbeeld" lead magnet
func (h *HTTPHandlerV2) HandleSampleReportDownload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	companyName := r.FormValue("company_name")

	// Validate
	if email == "" || !strings.Contains(email, "@") {
		http.Error(w, "Geldig e-mailadres vereist", http.StatusBadRequest)
		return
	}

	// Business email validation
	if !isBusinessEmail(email) {
		http.Error(w, getBusinessEmailError(), http.StatusBadRequest)
		return
	}

	// Store lead (sample download = warm lead)
	if db != nil {
		lead := &database.Lead{
			Email:               email,
			LeadType:            "sample-download",
			Source:              "website",
			Status:              "warm",
			LeadMagnetRequested: stringPtr("sample-tcf-report"),
		}
		if err := db.CreateLead(r.Context(), lead); err != nil {
			log.Printf("Failed to create lead: %v", err)
		}
	}

	// Send sample report via email
	if err := h.sendSampleReportEmail(email, companyName); err != nil {
		log.Printf("Failed to send sample report email: %v", err)
		http.Error(w, "Email verzenden mislukt", http.StatusInternalServerError)
		return
	}

	// Admin notification
	go h.sendSampleDownloadNotification(email, companyName)

	// Success response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`
		<div class="bg-blue-900/30 px-6 py-4 rounded-lg border border-blue-500/30">
			<p class="text-white font-semibold">Voorbeeldrapport onderweg!</p>
			<p class="text-sm text-gray-300 mt-2">We hebben het Technical Construction File naar %s gestuurd.</p>
			<p class="text-xs text-gray-400 mt-2">Check uw inbox - dit is de "gouden standaard" voor CRA compliance.</p>
		</div>
	`, email)))
}

// sendSampleReportEmail sends the gold standard sample TCF
func (h *HTTPHandlerV2) sendSampleReportEmail(to, companyName string) error {
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
	m.SetHeader("Subject", "Uw CRA Technical Construction File Voorbeeldrapport")

	recipientName := companyName
	if recipientName == "" {
		recipientName = "daar"
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Georgia, serif; line-height: 1.7; color: #1a1a1a; }
        .container { max-width: 650px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #1428A0 0%%, #FF6B35 100%%); color: white; padding: 40px; border-radius: 12px; text-align: center; }
        .logo { font-size: 32px; font-weight: bold; letter-spacing: 2px; }
        .content { background: #fafaf9; padding: 40px; margin-top: 20px; border-radius: 12px; border-left: 6px solid #FF6B35; }
        .highlight-box { background: #fff; padding: 25px; margin: 25px 0; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .stat { display: inline-block; text-align: center; margin: 15px 20px; }
        .stat-number { font-size: 42px; font-weight: bold; color: #1428A0; }
        .stat-label { font-size: 13px; color: #666; text-transform: uppercase; letter-spacing: 1px; }
        .button { display: inline-block; background: linear-gradient(135deg, #1428A0 0%%, #FF6B35 100%%); color: white; padding: 16px 40px; text-decoration: none; border-radius: 8px; margin: 25px 0; font-weight: 600; font-size: 16px; box-shadow: 0 4px 12px rgba(255,107,53,0.3); }
        .checklist { background: #f0fdf4; padding: 20px; border-left: 4px solid #22c55e; border-radius: 6px; margin: 20px 0; }
        .footer { margin-top: 40px; padding-top: 25px; border-top: 2px solid #e5e7eb; font-size: 12px; color: #666; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">LEONA</div>
            <p style="margin: 15px 0 0 0; opacity: 0.95; font-size: 16px; font-family: system-ui, sans-serif;">Technical Compliance Engineering</p>
        </div>
        
        <div class="content">
            <p style="font-size: 18px; margin-bottom: 25px;">Beste %s,</p>
            
            <p>Bedankt voor uw interesse in professionele CRA compliance documentatie. Bijgevoegd treft u ons <strong>Sample Technical Construction File</strong> - de standaard die wij hanteren voor notified body audits.</p>

            <div class="highlight-box" style="text-align: center;">
                <h2 style="color: #1428A0; margin-top: 0;">📄 Wat zit in dit voorbeeldrapport?</h2>
                
                <div class="stat">
                    <div class="stat-number">42</div>
                    <div class="stat-label">Pagina's</div>
                </div>
                <div class="stat">
                    <div class="stat-number">187</div>
                    <div class="stat-label">Componenten</div>
                </div>
                <div class="stat">
                    <div class="stat-number">14</div>
                    <div class="stat-label">CVE's</div>
                </div>
                <div class="stat">
                    <div class="stat-number">92%%</div>
                    <div class="stat-label">Score</div>
                </div>
            </div>

            <h3 style="color: #1428A0;">📋 Rapport inhoud (volledige structuur)</h3>
            <div class="checklist">
                ✅ <strong>Executive Summary</strong> - Board-ready compliance overzicht<br/>
                ✅ <strong>CRA Annex I Mapping</strong> - Artikel-voor-artikel compliance<br/>
                ✅ <strong>SBOM Analysis</strong> - 187 componenten met CPE traceability<br/>
                ✅ <strong>CVE Vulnerability Report</strong> - NVD database integratie<br/>
                ✅ <strong>License Compliance Matrix</strong> - GPL risk assessment<br/>
                ✅ <strong>Kernel EOL Validation</strong> - LTS lifecycle analysis<br/>
                ✅ <strong>Remediation Roadmap</strong> - Prioritized fix recommendations<br/>
                ✅ <strong>Notified Body Checklist</strong> - Pre-audit preparation guide
            </div>

            <div style="text-align: center;">
                <a href="https://leonacompliance.be/downloads/Sample_TCF_Report_LEONA_CRAVIT.pdf" class="button">
                    📥 Download Voorbeeldrapport (PDF, 42 pagina's)
                </a>
            </div>

            <h3 style="color: #1428A0; margin-top: 40px;">🎯 Waarom is dit relevant voor u?</h3>
            <p>Dit rapport toont <em>exact</em> wat uw embedded product nodig heeft voor CRA compliance:</p>
            <ul style="line-height: 2;">
                <li><strong>Audit-ready format:</strong> Voldoet aan eisen van notified bodies (TÜV, SGS, BSI)</li>
                <li><strong>Technisch + Juridisch:</strong> Vertaalt SBOM data naar Article 10.4 / Annex I requirements</li>
                <li><strong>Real-world voorbeeld:</strong> Gebaseerd op Yocto 4.3 Nanbield embedded Linux stack</li>
                <li><strong>Automated generation:</strong> Dit rapport kan binnen 60 seconden gemaakt worden voor <em>uw</em> product</li>
            </ul>

            <div style="background: #fef3c7; padding: 25px; border-left: 4px solid #f59e0b; border-radius: 8px; margin: 30px 0;">
                <h4 style="color: #92400e; margin-top: 0;">💼 Wilt u dit rapport voor uw eigen product?</h4>
                <p style="margin-bottom: 15px;">Upload uw SBOM naar onze V-Assessor™ en ontvang binnen 60 secunden:</p>
                <ul style="margin: 0; padding-left: 20px;">
                    <li>Real-time CVE detection via NVD database</li>
                    <li>Automated Annex I compliance mapping</li>
                    <li>Executive-ready PDF rapport (€499)</li>
                    <li>Prioritized remediation roadmap</li>
                </ul>
            </div>

            <div style="text-align: center; margin-top: 35px;">
                <a href="https://leonacompliance.be/#assessor" class="button">
                    🚀 Scan Mijn Product Nu (Gratis Preview)
                </a>
            </div>

            <h3 style="color: #1428A0; margin-top: 40px;">📞 Vragen over het rapport?</h3>
            <p>Reply direct op deze email of plan een 30-minuten demo:</p>
            <p style="margin-top: 15px;">
                📧 <a href="mailto:support@leonacompliance.be" style="color: #1428A0; font-weight: 600;">support@leonacompliance.be</a><br/>
                🌐 <a href="https://leonacompliance.be" style="color: #1428A0; font-weight: 600;">leonacompliance.be</a>
            </p>
        </div>

        <div class="footer">
            <p style="font-size: 14px; font-weight: 600; color: #1428A0;">LEONA</p>
            <p>Royal Blue (#1428A0) & Davis Orange (#FF6B35)</p>
            <p style="margin-top: 15px; font-size: 11px; color: #999;">
                Dit voorbeeldrapport is voor educatieve doeleinden. De data is geanonimiseerd en gebaseerd op een referentie-implementatie.<br/>
                Voor uw eigen product scan geldt de volledige NDA confidentiality guarantee.
            </p>
        </div>
    </div>
</body>
</html>
`, recipientName)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	return d.DialAndSend(m)
}

// sendSampleDownloadNotification notifies admin
func (h *HTTPHandlerV2) sendSampleDownloadNotification(email, companyName string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "kim@eliama.agency")
	m.SetHeader("Subject", fmt.Sprintf("📄 Sample Report Lead: %s", email))

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: system-ui, sans-serif;">
    <h2>📄 Sample TCF Download Lead</h2>
    <table style="background: #f3f4f6; padding: 20px; border-radius: 8px;">
        <tr><td><strong>Email:</strong></td><td>%s</td></tr>
        <tr><td><strong>Bedrijf:</strong></td><td>%s</td></tr>
        <tr><td><strong>Lead Type:</strong></td><td>Warm (Sample Download)</td></tr>
    </table>
    
    <h3>🎯 Follow-up Strategie</h3>
    <p><strong>Dag 1:</strong> Sample report verstuurd (✅ Done)<br/>
    <strong>Dag 2:</strong> Email: "Vragen over het rapport?"<br/>
    <strong>Dag 5:</strong> Email: "Upload uw SBOM voor een gratis preview scan"<br/>
    <strong>Dag 10:</strong> Email: "Case study - hoe bedrijf X in 3 dagen compliant werd"</p>
</body>
</html>
`, email, companyName)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send sample download notification: %v", err)
	} else {
		log.Printf("✅ Sample download notification sent: %s", email)
	}
}
