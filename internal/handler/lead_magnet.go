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

// LeadMagnetRequest represents an email submission for lead magnets
type LeadMagnetRequest struct {
	Email string `json:"email"`
	Type  string `json:"type"` // "engineer" or "lawyer"
}

// HandleEngineerLeadMagnet handles the meta-leona layer download
func (h *HTTPHandlerV2) HandleEngineerLeadMagnet(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")

	// Validate email
	if email == "" || !strings.Contains(email, "@") {
		http.Error(w, "Geldig e-mailadres vereist", http.StatusBadRequest)
		return
	}

	// Business email validation
	if !isBusinessEmail(email) {
		http.Error(w, getBusinessEmailError(), http.StatusBadRequest)
		return
	}

	// Create lead in database (if available)
	if db != nil {
		lead := &database.Lead{
			Email:               email,
			LeadType:            "engineer",
			Source:              "website",
			Status:              "new",
			LeadMagnetRequested: stringPtr("meta-leona"),
		}
		if err := db.CreateLead(r.Context(), lead); err != nil {
			log.Printf("Failed to create lead: %v", err)
			// Continue anyway - don't block user
		}
	}

	// Send email with download link
	if err := h.sendEngineerEmail(email); err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, "Email verzenden mislukt", http.StatusInternalServerError)
		return
	}

	// Return success response (HTML for HTMX)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<p class="text-sm text-green-300 bg-green-900/30 px-4 py-2 rounded-lg border border-green-500/30">✅ Check je inbox! We hebben de meta-leona layer naar ` + email + ` gestuurd.</p>`))
}

// HandleLawyerLeadMagnet handles the Annex I mapping template download
func (h *HTTPHandlerV2) HandleLawyerLeadMagnet(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")

	// Validate email
	if email == "" || !strings.Contains(email, "@") {
		http.Error(w, "Geldig e-mailadres vereist", http.StatusBadRequest)
		return
	}

	// Business email validation
	if !isBusinessEmail(email) {
		http.Error(w, getBusinessEmailError(), http.StatusBadRequest)
		return
	}

	// Create lead in database (if available)
	if db != nil {
		lead := &database.Lead{
			Email:               email,
			LeadType:            "lawyer",
			Source:              "website",
			Status:              "new",
			LeadMagnetRequested: stringPtr("annex-i-template"),
		}
		if err := db.CreateLead(r.Context(), lead); err != nil {
			log.Printf("Failed to create lead: %v", err)
			// Continue anyway - don't block user
		}
	}

	// Send email with download link
	if err := h.sendLawyerEmail(email); err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, "Email verzenden mislukt", http.StatusInternalServerError)
		return
	}

	// Return success response (HTML for HTMX)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<p class="text-sm text-green-300 bg-green-900/30 px-4 py-2 rounded-lg border border-green-500/30">✅ Check je inbox! We hebben de Annex I Mapping Template naar ` + email + ` gestuurd.</p>`))
}

// sendEngineerEmail sends the meta-leona layer to engineers
func (h *HTTPHandlerV2) sendEngineerEmail(to string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465 // SSL/TLS for Netim
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Jouw meta-leona CRA Validator Layer")

	// Email body
	body := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: system-ui, -apple-system, sans-serif; line-height: 1.6; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #f9f9f9; padding: 30px; margin-top: 20px; border-radius: 8px; }
        .button { display: inline-block; background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; margin: 20px 0; font-weight: bold; }
        .code { background: #2d2d2d; color: #f8f8f2; padding: 15px; border-radius: 6px; font-family: 'Courier New', monospace; font-size: 13px; overflow-x: auto; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">🚀 Jouw meta-leona Layer</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">Automatische CRA-checks tijdens je BitBake build</p>
        </div>
        
        <div class="content">
            <p>Bedankt voor je interesse in LEONA & CRAVIT!</p>
            
            <p>Hieronder vind je de <strong>meta-leona</strong> Yocto layer. Deze layer voegt automatische CRA compliance checks toe aan je build:</p>
            
            <h3>📦 Wat zit erin?</h3>
            <ul>
                <li><code>cra-check.bbclass</code> - CycloneDX SBOM generator</li>
                <li><code>kernel-eol-validator.bbclass</code> - LTS kernel check</li>
                <li><code>busybox-audit.bbclass</code> - Telnet/FTP service detector</li>
                <li>CPE traceability validator</li>
                <li>GPL license risk scanner</li>
            </ul>

            <a href="https://leonacompliance.be/downloads/meta-leona.tar.gz" class="button">Download meta-leona layer</a>

            <h3>⚙️ Installatie (2 minuten)</h3>
            <div class="code">
# 1. Extract layer naar je Yocto workspace<br/>
cd ~/yocto/sources<br/>
tar -xzf meta-leona.tar.gz<br/><br/>

# 2. Voeg toe aan bblayers.conf<br/>
bitbake-layers add-layer ../sources/meta-leona<br/><br/>

# 3. Enable CRA checks in local.conf<br/>
echo 'INHERIT += "cra-check kernel-eol-validator"' >> conf/local.conf<br/><br/>

# 4. Build je image<br/>
bitbake core-image-minimal
            </div>

            <p><strong>Output:</strong> Na de build vind je <code>tmp/deploy/images/&lt;machine&gt;/bom.json</code> - upload dit naar <a href="https://leonacompliance.be/#assessor">V-Assessor™</a> voor een gratis compliance scan.</p>

            <h3>🎯 Volgende stap</h3>
            <p>Wil je een volledige CRA gap analysis van je huidige build?</p>
            <a href="https://leonacompliance.be/#assessor" class="button">Upload je SBOM (gratis)</a>
        </div>

        <div class="footer">
            <p><strong>LEONA & CRAVIT</strong> | CRA Compliance Engineering<br/>
            Vragen? Reply op deze email of bezoek <a href="https://leonacompliance.be">leonacompliance.be</a></p>
            <p style="margin-top: 15px; font-size: 11px; color: #999;">
                Deze tools zijn community contributions. Voor productie-gebruik raden we een volledige V-Assessor™ audit aan (€499).
            </p>
        </div>
    </div>
</body>
</html>
`

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	return d.DialAndSend(m)
}

// sendLawyerEmail sends the Annex I template to lawyers
func (h *HTTPHandlerV2) sendLawyerEmail(to string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465 // SSL/TLS for Netim
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "CRA Annex I Mapping Template voor uw cliënten")

	// Email body
	body := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Georgia, serif; line-height: 1.8; color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #1e3a8a 0%%, #3730a3 100%%); color: white; padding: 30px; border-radius: 8px; }
        .content { background: #fafaf9; padding: 30px; margin-top: 20px; border-radius: 8px; border-left: 4px solid #3730a3; }
        .button { display: inline-block; background: #3730a3; color: white; padding: 14px 35px; text-decoration: none; border-radius: 6px; margin: 20px 0; font-weight: 600; }
        .table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
        .table th { background: #f3f4f6; padding: 10px; text-align: left; font-size: 12px; border-bottom: 2px solid #d1d5db; }
        .table td { padding: 10px; border-bottom: 1px solid #e5e7eb; font-size: 13px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 11px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0; font-size: 24px; font-weight: normal;">Annex I Mapping Template</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9; font-size: 14px;">Voor juridische adviseurs & compliance teams</p>
        </div>
        
        <div class="content">
            <p>Geachte,</p>
            
            <p>Bedankt voor uw interesse in LEONA & CRAVIT's compliance instrumentarium. Bijgevoegd treft u onze <strong>Annex I Mapping Template</strong> - een Excel-based tool om technische bevindingen te koppelen aan juridische eisen.</p>
            
            <h3 style="color: #1e3a8a; margin-top: 25px;">📋 Wat bevat de template?</h3>
            <ul style="line-height: 2;">
                <li><strong>CRA Artikel mapping</strong> (10.4, 14.1, 14.2, Annex I Parts I & II)</li>
                <li><strong>CPE traceability checklist</strong> voor SBOM validatie</li>
                <li><strong>Tivoization validator</strong> (GPLv3 compliance risico's)</li>
                <li><strong>Vulnerability disclosure timeline</strong> calculator</li>
                <li>Client intake vragenlijst (embedded Linux specifiek)</li>
            </ul>

            <a href="https://leonacompliance.be/downloads/CRA_Annex_I_Mapping_Template.xlsx" class="button">Download Template (Excel)</a>

            <h3 style="color: 1e3a8a; margin-top: 25px;">🔍 Use case: Cliënt heeft technisch rapport nodig</h3>
            <p>Veel machinebouwers komen bij u met de vraag: <em>"Hoe toetsen we of onze embedded Linux stack CRA-compliant is?"</em></p>
            
            <p>U kunt nu:</p>
            <ol style="line-height: 2;">
                <li>Vraag hun SBOM (via Yocto cve-check.bbclass of Buildroot)</li>
                <li>Upload naar <a href="https://leonacompliance.be/#assessor">V-Assessor™</a> (geautomatiseerde pre-scan)</li>
                <li>Ontvang gap analysis rapport binnen 60 seconden</li>
                <li>Gebruik onze template om bevindingen te vertalen naar juridische adviezen</li>
            </ol>

            <h3 style="color: #1e3a8a; margin-top: 25px;">💼 Partnership optie</h3>
            <p>Voor advocatenkantoren bieden we een <strong>Lawyer Master Account</strong> (€2.500/jaar):</p>
            <table class="table">
                <tr>
                    <th>Voordeel</th>
                    <th>Details</th>
                </tr>
                <tr>
                    <td>Unlimited client scans</td>
                    <td>Geen per-rapport kosten</td>
                </tr>
                <tr>
                    <td>White-label rapportage</td>
                    <td>Uw logo op technische dossiers</td>
                </tr>
                <tr>
                    <td>Priority support</td>
                    <td>Direct juridisch + technisch team</td>
                </tr>
                <tr>
                    <td>Quarterly CRA updates</td>
                    <td>Wijzigingen in EU guidance documents</td>
                </tr>
            </table>

            <p>Interesse? Plan een 30-minuten demo: <a href="mailto:support@leonacompliance.be">support@leonacompliance.be</a></p>
        </div>

        <div class="footer">
            <p><strong>LEONA & CRAVIT</strong><br/>
            Technisch-juridische compliance engineering<br/>
            <a href="https://leonacompliance.be">leonacompliance.be</a> | <a href="mailto:support@leonacompliance.be">support@leonacompliance.be</a></p>
            <p style="margin-top: 15px; font-size: 10px; color: #999;">
                Deze template is voor educatieve doeleinden. Juridisch advies blijft uw verantwoordelijkheid.<br/>
                Wij zijn geen geaccrediteerde notified body - wel een technische due diligence partner.
            </p>
        </div>
    </div>
</body>
</html>
`

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	return d.DialAndSend(m)
}

func stringPtr(s string) *string {
	return &s
}

// Global db variable for lead tracking
var db *database.SupabaseClient
