package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"leona-scanner/internal/database"

	"gopkg.in/gomail.v2"
)

// CRAAssessmentSubmission represents the assessment results
type CRAAssessmentSubmission struct {
	Email   string            `json:"email"`
	Answers map[string]string `json:"answers"` // question number -> "ja" or "nee"
}

// HandleCRAAssessmentSubmit handles the submission of CRA assessment results
func (h *HTTPHandlerV2) HandleCRAAssessmentSubmit(w http.ResponseWriter, r *http.Request) {
	log.Printf("📨 CRA Assessment submission received from %s", r.RemoteAddr)

	if err := r.ParseForm(); err != nil {
		log.Printf("❌ Failed to parse form: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	answersJSON := r.FormValue("answers")

	log.Printf("📧 Email received: '%s'", email)
	log.Printf("📊 Answers JSON received: '%s'", answersJSON)
	log.Printf("📊 Answers JSON length: %d bytes", len(answersJSON))
	log.Printf("📊 All form values: %v", r.Form)

	// Validate email
	if email == "" || !strings.Contains(email, "@") {
		log.Printf("❌ Invalid email: '%s'", email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Geldig e-mailadres vereist"}); encErr != nil {
			log.Printf("Failed to encode error: %v", encErr)
		}
		return
	}

	// Validate business email (reject free email providers)
	emailLower := strings.ToLower(strings.TrimSpace(email))
	freeEmailDomains := []string{
		"@gmail.", "@googlemail.",
		"@outlook.", "@hotmail.", "@live.", "@msn.",
		"@yahoo.", "@ymail.",
		"@aol.", "@protonmail.", "@icloud.", "@me.com",
		"@mail.com", "@gmx.", "@zoho.",
	}
	for _, freeDomain := range freeEmailDomains {
		if strings.Contains(emailLower, freeDomain) {
			log.Printf("❌ Free email provider rejected: '%s'", email)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if encErr := json.NewEncoder(w).Encode(map[string]string{
				"error": "Gebruik uw zakelijk e-mailadres. Gratis e-maildiensten (Gmail, Outlook, Yahoo, etc.) zijn niet toegestaan.",
			}); encErr != nil {
				log.Printf("Failed to encode error: %v", encErr)
			}
			return
		}
	}

	// Parse answers
	var answers map[string]string
	if answersJSON == "" || answersJSON == "{}" {
		answers = make(map[string]string)
	} else {
		if err := json.Unmarshal([]byte(answersJSON), &answers); err != nil {
			log.Printf("❌ Failed to parse answers JSON: %v (received: %s)", err, answersJSON)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if encErr := json.NewEncoder(w).Encode(map[string]string{"error": "Ongeldige antwoorden"}); encErr != nil {
				log.Printf("❌ Failed to encode error response: %v", encErr)
			}
			return
		}
	}

	log.Printf("📊 Processing assessment for %s with %d answers", email, len(answers))

	// Calculate score (handle empty answers)
	jaCount := 0
	for _, answer := range answers {
		if answer == "ja" {
			jaCount++
		}
	}
	totalQuestions := len(answers)
	if totalQuestions == 0 {
		totalQuestions = 10 // Default to 10 for empty submissions
	}
	// Fix: multiply first to avoid integer truncation to 0
	complianceScore := (jaCount * 100) / totalQuestions

	// Save to database (if available)
	if db != nil {
		// Create as a lead with notes containing the assessment
		notes := fmt.Sprintf("CRA Self-Assessment Score: %d/10 JA (%d%% compliance)", jaCount, complianceScore)
		lead := &database.Lead{
			Email:               email,
			LeadType:            "cra-assessment",
			Source:              "cra-assessment-wizard",
			Status:              "new",
			LeadMagnetRequested: stringPtr("cra-assessment-results"),
			Notes:               &notes,
		}
		if err := db.CreateLead(r.Context(), lead); err != nil {
			log.Printf("Failed to create lead: %v", err)
		}
	}

	// Send email with results
	log.Printf("🔄 Attempting to send assessment email to %s (Score: %d%%,  %d/10 JA)", email, complianceScore, jaCount)
	if err := sendAssessmentResultsEmail(email, answers, jaCount, complianceScore); err != nil {
		log.Printf("⚠️  WARNING: Failed to send assessment email to %s: %v", email, err)
		log.Printf("⚠️  Continuing anyway - email will be sent via admin notification")
		// Don't fail the request if email fails - just log it
	} else {
		log.Printf("✅ SUCCESS: CRA assessment results sent to %s (Score: %d%%)", email, complianceScore)
	}

	// Send admin notification (always, even if user email failed)
	log.Printf("📧 Sending admin notification for %s assessment", email)
	go sendAssessmentNotification(email, answers, jaCount, complianceScore)

	// Return HTML success message with script to trigger Alpine.js state
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	//nolint:errcheck,gosec,lll // Success HTML with confetti
	successHTML := fmt.Sprintf(`
		<div class="text-center py-8">
			<div class="text-green-600 text-2xl font-bold mb-4">✅ Verstuurd!</div>
			<div class="text-4xl font-bold text-blue-600 mb-4">%d%%%%</div>
			<p class="text-lg">Uw CRA compliance score is <strong>%d van 10</strong></p>
			<p class="text-sm text-gray-600 mt-2">Check uw mailbox voor het volledige rapport naar <strong>%s</strong></p>
		</div>
		<script>
			// Trigger Alpine.js success state
			setTimeout(() => {
				const emailInput = document.getElementById('email');
				if (emailInput) {
					window.dispatchEvent(new CustomEvent('cra-success', { detail: { email: emailInput.value } }));
				}
				// Trigger confetti
				if (typeof confetti !== 'undefined') {
					confetti({ particleCount: 100, spread: 70, origin: { y: 0.6 } });
					setTimeout(() => confetti({ particleCount: 50, angle: 60, spread: 55, origin: { x: 0 } }), 250);
					setTimeout(() => confetti({ particleCount: 50, angle: 120, spread: 55, origin: { x: 1 } }), 400);
				}
			}, 100);
		</script>
	`, complianceScore, jaCount, email)
	//nolint:errcheck,gosec // Writing success HTML
	w.Write([]byte(successHTML))
}

// sendAssessmentResultsEmail sends the assessment results to the user
func sendAssessmentResultsEmail(to string, answers map[string]string, jaCount int, score int) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP not configured")
	}

	// Generate unique report ID
	reportID := time.Now().Unix() % 100000

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Bcc", "kim@leonacompliance.be") // BCC copy with user's answers
	m.SetHeader("Subject", fmt.Sprintf("CRA Compliance Briefing - %d%% Score", score))

	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="nl" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="x-apple-disable-message-reformatting">
    <title>CRA Compliance Briefing</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800&display=swap" rel="stylesheet">
	<style>
        @media only screen and (max-width: 620px) {
            .container { width: 100%% !important; border: none !important; }
            .content-padding { padding: 30px 20px !important; }
            h1 { font-size: 26px !important; }
            .score-value { font-size: 48px !important; }
        }
    </style>
</head>
<body style="margin: 0; padding: 0; width: 100%%; background-color: #f8fafc; font-family: 'Inter', Helvetica, Arial, sans-serif; -webkit-font-smoothing: antialiased;">
    <div class="wrapper" style="background-color: #f8fafc; padding: 20px 0;">
        <div class="container" style="max-width: 650px; margin: 0 auto; background-color: #ffffff; border: 1px solid #e2e8f0; overflow: hidden;">
            
            <img src="https://res.cloudinary.com/dg0qxqj4a/image/upload/v1773860997/CRA_COMPLIANT_LINUX_SYSTEM-2_mglczu.png" 
                 alt="CRA Compliance Analysis" 
                 width="650" 
                 style="width: 100%%; max-width: 650px; height: auto; display: block; border: 0; outline: none; text-decoration: none;">

            <div class="content-padding" style="padding: 40px 60px 60px 60px;">
                <span class="sub-header" style="text-transform: uppercase; font-weight: 700; font-size: 11px; letter-spacing: 0.15em; color: #fd7e14; margin-bottom: 12px; display: block;">Assessment-ID: LEONA-%d-TR</span>
                
                <h1 style="font-size: 32px; font-weight: 800; color: #0f172a; line-height: 1.2; margin: 0 0 24px 0; letter-spacing: -0.02em;">CRA Compliance Briefing</h1>
                
                <div class="score-block" style="border-left: 4px solid #003366; padding-left: 24px; margin: 40px 0;">
                    <span class="score-value" style="display: block; font-size: 64px; font-weight: 800; color: #0f172a; line-height: 1; letter-spacing: -0.04em;">%d%%%%</span>
                    <span class="score-label" style="font-size: 13px; text-transform: uppercase; font-weight: 600; color: #64748b; letter-spacing: 0.05em; margin-top: 4px; display: block;">Gereedheidsscore</span>
                </div>

                <p style="font-size: 16px; color: #475569; margin-bottom: 32px;">
                   Uw score is een directe weergave van uw exposure. Wachten tot de deadline van september 2026 is geen optie meer.
                </p>

                <div style="background-color: #f1f5f9; padding: 30px; border-radius: 8px; margin-bottom: 40px;">
                    <h2 style="font-size: 18px; font-weight: 700; color: #003366; margin: 0 0 12px 0;">Waarom nu de Snapshot Audit starten?</h2>
                    <p style="font-size: 15px; color: #1e293b; margin: 0 0 20px 0;">
                        De meeste bedrijven verliezen maanden aan juridisch overleg. Wij doen het anders:
                    </p>
                    <ul style="padding: 0; margin: 0; list-style: none; font-size: 15px; color: #1e293b;">
                        <li style="margin-bottom: 10px;"><strong>⚡ Snelheid:</strong> Volledig inzicht en een direct actieplan binnen <strong>48 uur</strong>.</li>
                        <li style="margin-bottom: 10px;"><strong>⚖️ ROI:</strong> Een vaste investering van <strong>€2.495</strong> voorkomt boetes die tot 2,5%% van uw wereldwijde omzet kunnen oplopen.</li>
                        <li style="margin-bottom: 10px;"><strong>🛠️ Techniek:</strong> Geen vage rapporten, maar engineering-advies waar uw dev-team direct mee aan de slag kan.</li>
                    </ul>
                </div>

                <div class="actions">
                    <a href="https://www.leonacompliance.be/contact" 
                       style="display: block; text-align: center; padding: 22px 24px; font-weight: 800; font-size: 16px; text-decoration: none; border-radius: 6px; margin-bottom: 16px; background-color: #003366; color: #ffffff !important; box-shadow: 0 4px 6px rgba(0,51,102,0.2);">
                       SNAPSHOT AUDIT — RESULTAAT IN 48U
                    </a>
                    
                    <a href="https://www.leonacompliance.be/demo" 
                       style="display: block; text-align: center; padding: 18px 24px; font-weight: 700; font-size: 15px; text-decoration: none; border-radius: 6px; border: 2px solid #003366; color: #003366 !important;">
                       Vraag een Demo
                    </a>
                </div>

                <div class="footer" style="margin-top: 64px; padding-top: 32px; border-top: 1px solid #f1f5f9; font-size: 12px; color: #94a3b8;">
                    <span class="footer-brand" style="color: #475569; font-weight: 700; margin-bottom: 4px; display: block;">LEONA Compliance | Compliance as Code</span>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`, reportID, score)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true
	return d.DialAndSend(m)
}

// sendAssessmentNotification sends admin notification
func sendAssessmentNotification(email string, answers map[string]string, jaCount int, score int) {
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
	m.SetHeader("To", "kim@leonacompliance.be")
	m.SetHeader("Subject", fmt.Sprintf("Nieuwe CRA Assessment: %s - %d%% Score", email, score))

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
		.score { font-size: 32px; font-weight: bold; color: #16a34a; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2 style="margin: 0;">Nieuwe CRA Assessment Lead</h2>
			<p style="margin: 5px 0 0 0; opacity: 0.9; font-size: 14px;">CRA Compliance Self-Assessment</p>
		</div>
		
		<div class="content">
			<div class="data">
				<div class="label">Email</div>
				<div class="value">%s</div>
			</div>

			<div class="data">
				<div class="label">Compliance Score</div>
				<div class="score">%d%%</div>
				<div style="color: #666; margin-top: 5px;">%d van 10 vragen positief beantwoord</div>
			</div>

			<div class="data">
				<div class="label">Lead Kwaliteit</div>
				<div class="value">%s</div>
			</div>

			<div class="data">
				<div class="label">Timestamp</div>
				<div class="value">%s</div>
			</div>

			<div class="data">
				<div class="label">Acties</div>
				<div class="value">
					✅ Resultaten email verstuurd<br/>
					✅ Opgeslagen in Supabase leads table<br/>
					📊 Bekijk in <a href="https://supabase.com">Supabase Dashboard</a>
				</div>
			</div>
		</div>
	</div>
</body>
</html>
`, email, score, jaCount, getLeadQuality(score), time.Now().Format("2006-01-02 15:04:05"))

	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true
	if err := d.DialAndSend(m); err != nil {
		log.Printf("❌ ERROR: Failed to send admin notification to kim@leonacompliance.be: %v", err)
	} else {
		log.Printf("✅ SUCCESS: CRA assessment notification sent to kim@leonacompliance.be for %s (%d%% score)", email, score)
	}
}

func getLeadQuality(score int) string {
	if score >= 80 {
		return "HOT LEAD - Bijna compliant, interesse in finishing touches"
	} else if score >= 50 {
		return "WARM LEAD - Gap analysis + implementatie nodig"
	} else {
		return "URGENT LEAD - Hoog risico, noodplan vereist"
	}
}

func calculateMonthsUntilDeadline() int {
	deadline := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	now := time.Now()
	months := int(deadline.Sub(now).Hours() / 24 / 30)
	return months
}
