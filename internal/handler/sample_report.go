package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
		log.Printf("⚠️  WARNING: Failed to send sample report email: %v", err)
		log.Printf("⚠️  Continuing anyway - user will still see success message")
		// Don't fail the request if email fails - just log it
	} else {
		log.Printf("✅ Sample report email sent successfully to %s", email)
	}

	// Admin notification (always send, even if user email failed)
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
	// Use Mailgun if configured, fallback to SMTP
	if h.mailgunService != nil {
		return h.sendSampleReportEmailViaMailgun(to, companyName)
	}

	// Fallback to SMTP
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := getSMTPPort()
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("email service not configured")
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
<html lang="nl" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800&display=swap" rel="stylesheet">
    <style>
        body { font-family: 'Inter', -apple-system, sans-serif; line-height: 1.6; color: #1e293b; background-color: #f8fafc; margin: 0; padding: 0; -webkit-font-smoothing: antialiased; }
        .wrapper { padding: 40px 20px; }
        .container { max-width: 650px; margin: 0 auto; background: #ffffff; border: 1px solid #e2e8f0; box-shadow: 0 1px 3px rgba(0,0,0,0.05); overflow: hidden; }
        .hero-image { width: 100%; height: auto; display: block; border: 0; }
        .content-padding { padding: 40px 60px 60px 60px; }
        
        .sub-header { text-transform: uppercase; font-weight: 700; font-size: 11px; letter-spacing: 0.15em; color: #fd7e14; margin-bottom: 12px; display: block; }
        h1 { font-size: 32px; font-weight: 800; color: #0f172a; line-height: 1.2; margin: 0 0 24px 0; letter-spacing: -0.02em; }
        
        /* Stats Grid */
        .stats-grid { margin: 40px 0; border-top: 1px solid #f1f5f9; border-bottom: 1px solid #f1f5f9; padding: 24px 0; display: table; width: 100%; }
        .stat-item { display: table-cell; text-align: center; width: 25%; }
        .stat-number { display: block; font-size: 20px; font-weight: 800; color: #003366; }
        .stat-label { display: block; font-size: 9px; text-transform: uppercase; font-weight: 700; color: #64748b; letter-spacing: 0.05em; }

        .section-title { font-size: 14px; text-transform: uppercase; letter-spacing: 0.1em; font-weight: 800; color: #0f172a; margin: 40px 0 20px 0; border-left: 3px solid #003366; padding-left: 15px; }
        .list-item { margin-bottom: 15px; font-size: 14px; color: #475569; }
        .list-item strong { color: #0f172a; }

        /* CTA Section */
        .cta-box { background-color: #f1f5f9; padding: 32px; border-radius: 8px; margin: 40px 0; text-align: center; }
        .btn { display: block; text-align: center; padding: 18px 24px; font-weight: 700; font-size: 15px; text-decoration: none; border-radius: 6px; margin-bottom: 12px; }
        .btn-primary { background-color: #003366; color: #ffffff !important; }
        .btn-outline { border: 2px solid #003366; color: #003366 !important; }

        .footer { margin-top: 60px; padding-top: 30px; border-top: 1px solid #f1f5f9; font-size: 12px; color: #94a3b8; }
        
        @media (max-width: 600px) {
            .content-padding { padding: 30px; }
            .stat-item { width: 50%; display: inline-block; margin-bottom: 20px; }
        }
    </style>
</head>
<body>
    <div class="wrapper">
        <div class="container">
            <img src="https://res.cloudinary.com/dg0qxqj4a/image/upload/v1773871167/CRA_COMPLIANT_LINUX_SYSTEM-5_shs5we.png" alt="CRA Technical Construction File Review" class="hero-image">
            
            <div class="content-padding">
                <span class="sub-header">Conformiteitsbeoordeling Module B+C</span>
                <h1>Uw Sample TCF Rapport: SE-500 Gateway</h1>
                
                <p style="font-size: 16px; color: #475569; margin-bottom: 32px;">
                    Beste %s, stop met gissen naar de drempelwaarden van het BIPT. Dit <strong>Sample TCF Rapport</strong> voor de SE-500 Gateway is geen samenvatting, maar een juridisch en technisch bewijslast-dossier zoals geëist voor <strong>Class I Critical Products</strong>.
                </p>

                <div class="stats-grid">
                    <div class="stat-item">
                        <span class="stat-number">SL2+</span>
                        <span class="stat-label">Security Level</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-number">192</span>
                        <span class="stat-label">SBOM Libs</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-number">10J</span>
                        <span class="stat-label">Support</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-number">500u</span>
                        <span class="stat-label">Fuzzing</span>
                    </div>
                </div>

                <div class="section-title">De Bewijslast in dit Dossier</div>
                <div class="list-item"><strong>01. STRIDE Risk Assessment</strong> — Volledige dreigingsmatrix met mitigatie-strategieën conform Artikel 10.</div>
                <div class="list-item"><strong>02. Hardware RoT Evidence</strong> — Gedetailleerde onderbouwing van i.MX8M HAB, TrustZone en CAAM isolatie.</div>
                <div class="list-item"><strong>03. Binary Hardening Audit</strong> — Checksec bewijzen (Full RELRO, Canary, PIE) voor alle kritieke systeem-services.</div>
                <div class="list-item"><strong>04. Incident Reporting</strong> — Het formele 24u/72u protocol voor meldingen aan ENISA en het CSIRT.</div>

                <div style="margin: 40px 0;">
                    <a href="https://leonacompliance.be/downloads/Sample_TCF_SE500_BIPT_Ready.pdf" class="btn btn-primary">Download Audit-Ready Rapport (PDF)</a>
                </div>

                <div class="cta-box">
                    <p style="font-weight: 700; color: #0f172a; margin-bottom: 12px; font-size: 16px;">Voldoet uw huidige documentatie?</p>
                    <p style="font-size: 14px; color: #475569; margin-bottom: 24px;">
                        Onze <strong>Snapshot Audit</strong> vertaalt uw broncode en SBOM naar dit exacte format in minder dan 48 uur. Voorkom afkeuring door uw Notified Body.
                    </p>
                    <a href="https://leonacompliance.be/#assessor" class="btn btn-outline">Start Mijn Audit Review Nu</a>
                </div>

                <div class="footer">
                    <p><strong>LEONA Compliance</strong> | Engineering-Led CRA Solutions</p>
                    <p style="margin-top: 15px; font-size: 11px;">
                        Dit voorbeeld is gebaseerd op een geharde NXP i.MX8M Plus stack. Elk dossier dat wij leveren is specifiek ontworpen om te voldoen aan de Annex I en Annex III eisen van de Verordening (EU) 2024/2847.
                    </p>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
`, recipientName)

	m.SetBody("text/html", body)

	// Attach the sample PDF
	pdfPath := "./static/downloads/Sample_TCF_Report_LEONA.pdf"
	if _, err := os.Stat(pdfPath); err == nil {
		m.Attach(pdfPath)
	} else {
		log.Printf("⚠️  Warning: Sample PDF not found at %s", pdfPath)
	}

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	// Use SSL for port 465, STARTTLS for 587
	if smtpPort == 465 {
		d.SSL = true
	} else {
		d.SSL = false // Use STARTTLS for port 587
	}
	return d.DialAndSend(m)
}

// sendSampleDownloadNotification notifies admin
func (h *HTTPHandlerV2) sendSampleDownloadNotification(email, companyName string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := getSMTPPort()
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "kim@leonacompliance.be")
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
	if smtpPort == 465 {
		d.SSL = true
	} else {
		d.SSL = false
	}
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send sample download notification: %v", err)
	} else {
		log.Printf("✅ Sample download notification sent: %s", email)
	}
}

// getSMTPPort reads SMTP_PORT from env, defaults to 465 if not set
func getSMTPPort() int {
	portStr := os.Getenv("SMTP_PORT")
	if portStr == "" {
		return 465 // Default to SSL
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("⚠️  Invalid SMTP_PORT %s, using default 465", portStr)
		return 465
	}
	return port
}

// sendSampleReportEmailViaMailgun sends email via Mailgun with attachment
func (h *HTTPHandlerV2) sendSampleReportEmailViaMailgun(to, companyName string) error {
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
            
            <p>Bedankt voor uw interesse in professionele CRA compliance documentatie. <strong>Bijgevoegd als PDF-attachment</strong> treft u ons Sample Technical Construction File - de standaard die wij hanteren voor notified body audits.</p>

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
                ✅ <strong>CVE Vulnerability Report</strong> - NVD database integration<br/>
                ✅ <strong>License Compliance Matrix</strong> - GPL risk assessment<br/>
                ✅ <strong>Kernel EOL Validation</strong> - LTS lifecycle analysis<br/>
                ✅ <strong>Remediation Roadmap</strong> - Prioritized fix recommendations<br/>
                ✅ <strong>Notified Body Checklist</strong> - Pre-audit preparation guide
            </div>

            <h3 style="color: #1428A0; margin-top: 40px;">🎯 Waarom is dit relevant voor u?</h3>
            <p>Dit rapport toont <em>exact</em> wat uw embedded product nodig heeft voor CRA compliance:</p>
            <ul style="line-height: 2;">
                <li><strong>Audit-ready format:</strong> Voldoet aan eisen van notified bodies (TÜV, SGS, BSI)</li>
                <li><strong>Technisch + Juridisch:</strong> Vertaalt SBOM data naar Article 10.4 / Annex I requirements</li>
                <li><strong>Real-world voorbeeld:</strong> Gebaseerd op Yocto 4.3 Nanbield embedded Linux stack</li>
                <li><strong>Automated generation:</strong> Dit rapport kan binnen 60 seconden gemaakt worden voor <em>uw</em> product</li>
            </ul>

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

	pdfPath := "./static/downloads/Sample_TCF_Report_LEONA.pdf"
	subject := "Uw CRA Technical Construction File Voorbeeldrapport"

	return h.mailgunService.SendHTMLEmailWithAttachment(to, subject, body, pdfPath)
}
