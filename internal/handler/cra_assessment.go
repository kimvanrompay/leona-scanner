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

	// Validate email
	if email == "" || !strings.Contains(email, "@") {
		log.Printf("❌ Invalid email: '%s'", email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Geldig e-mailadres vereist"})
		return
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
	if err := sendAssessmentResultsEmail(email, answers, jaCount, complianceScore); err != nil {
		log.Printf("❌ ERROR: Failed to send assessment email to %s: %v", email, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email verzenden mislukt"})
		return
	}
	log.Printf("✅ SUCCESS: CRA assessment results sent to %s (Score: %d%%)", email, complianceScore)

	// Send admin notification
	go sendAssessmentNotification(email, answers, jaCount, complianceScore)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"score":   complianceScore,
		"message": "Resultaten verstuurd naar " + email,
	})
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

	// Determine risk level and recommendations
	var riskLevel, recommendation string
	if score >= 80 {
		riskLevel = "EXCELLENT"
		recommendation = "Uw technische basis is sterk. Focus op documentatie en procesformalisatie om de laatste 20% te bereiken."
	} else if score >= 50 {
		riskLevel = "MEDIUM RISK"
		recommendation = "Belangrijke gaps gedetecteerd in cryptografie, secure boot of supply chain tracking. Prioriteer deze high-impact gebieden."
	} else {
		riskLevel = "HIGH RISK"
		recommendation = "Kritieke tekortkomingen op meerdere fronten. Een volledige remediation roadmap is nodig om de deadline te halen."
	}

	// Build question results HTML
	questionResults := ""
	questionTitles := map[string]string{
		"1":  "SBOM Generatie",
		"2":  "CVE Tracking",
		"3":  "Secure Boot",
		"4":  "OTA Updates",
		"5":  "Cryptografie",
		"6":  "TCF Documentatie",
		"7":  "Supply Chain",
		"8":  "Incident Response",
		"9":  "Support Lifecycle",
		"10": "Access Control",
	}

	for i := 1; i <= 10; i++ {
		qNum := fmt.Sprintf("%d", i)
		answer := answers[qNum]
		status := "❌"
		statusColor := "#ef4444"
		if answer == "ja" {
			status = "✅"
			statusColor = "#22c55e"
		}
		questionResults += fmt.Sprintf(`
			<tr>
				<td style="padding: 12px; border-bottom: 1px solid #e5e7eb;">%d. %s</td>
				<td style="padding: 12px; border-bottom: 1px solid #e5e7eb; text-align: center; color: %s; font-size: 18px;">%s</td>
			</tr>
		`, i, questionTitles[qNum], statusColor, status)
	}

	// Generate unique report ID
	reportID := time.Now().Unix() % 100000

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Bcc", "kim@leonacompliance.be") // BCC copy with user's answers
	m.SetHeader("Subject", fmt.Sprintf("Uw CRA Technical Assessment - %d%% Score", score))

	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="nl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CRA Technical Assessment Report</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600;800&display=swap" rel="stylesheet">
	<style>
		body { 
            font-family: 'Inter', -apple-system, sans-serif; 
            line-height: 1.6; 
            color: #1e293b; 
            background-color: #f8fafc; 
            margin: 0; 
            padding: 0; 
        }
		.container { 
            max-width: 650px; 
            margin: 40px auto; 
            background: #ffffff; 
            border-top: 8px solid #003366;
            box-shadow: 0 4px 15px rgba(0,0,0,0.05);
        }
		
        .header { 
            padding: 40px 40px 20px 40px; 
            text-align: left; 
        }
		.header h1 { 
            margin: 0; 
            font-size: 22px; 
            color: #003366; 
            letter-spacing: 0.5px; 
            font-weight: 800; 
            text-transform: uppercase; 
        }
        .report-meta { 
            font-size: 12px; 
            color: #64748b; 
            margin-top: 5px; 
            letter-spacing: 1px;
        }
		
		.score-section { 
            background-color: #f1f5f9; 
            padding: 40px; 
            text-align: center; 
            margin: 0 40px;
            border-radius: 4px;
        }
		.score-number { 
            font-size: 82px; 
            font-weight: 800; 
            margin: 0; 
            line-height: 1; 
            color: #003366;
        }
		.score-label { 
            font-size: 14px; 
            color: #fd7e14;
            text-transform: uppercase; 
            font-weight: 600;
            letter-spacing: 2px;
        }

		.status-callout { 
            margin: 30px 40px; 
            padding: 20px; 
            border-left: 4px solid #fd7e14; 
            background: #fffcf9;
            font-size: 16px;
            color: #334155;
        }
		
		.content { padding: 0 40px 40px 40px; }
		.section-title { 
            color: #003366; 
            text-transform: uppercase; 
            font-weight: 600; 
            font-size: 14px; 
            border-bottom: 1px solid #e2e8f0; 
            display: block; 
            margin-bottom: 15px; 
            padding-bottom: 5px; 
            letter-spacing: 1px;
        }
		
		.results-table { 
            width: 100%%; 
            border-collapse: collapse; 
            margin: 20px 0; 
            font-size: 13px; 
        }
		.results-table td { 
            padding: 12px 0; 
            border-bottom: 1px solid #f1f5f9; 
        }

		.cta-container { margin: 30px 0; }
		.button { 
            display: block; 
            text-align: center; 
            background: #fd7e14;
            color: #ffffff !important; 
            padding: 18px; 
            text-decoration: none; 
            font-weight: 600; 
            font-size: 16px; 
            border-radius: 4px;
            margin-bottom: 12px;
            transition: background 0.2s ease;
        }
		.button:hover { background: #e86b00; }
		.button-outline { 
            background: transparent; 
            color: #003366 !important; 
            border: 2px solid #003366; 
        }

		.deadline-footer { 
            background: #003366; 
            color: #ffffff; 
            padding: 30px; 
            text-align: center; 
        }
        .months-left { 
            font-size: 28px; 
            color: #fd7e14; 
            display: block; 
            font-weight: 800; 
        }

		.footer-legal { 
            padding: 40px; 
            font-size: 11px; 
            color: #94a3b8; 
            background: #ffffff;
        }
        .footer-legal a { color: #64748b; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>CRA Compliance Audit</h1>
            <div class="report-meta">ID: LEONA-%d-TR | DATUM: 18 MRT 2026</div>
		</div>
		
		<div class="score-section">
			<div class="score-label">Technical Readiness Score</div>
			<div class="score-number">%d%%%%</div>
			<div style="font-size: 16px; font-weight: 400; margin-top: 5px; color: #64748b;">Status: %s</div>
		</div>
		
		<div class="status-callout">
			"De cijfers liegen niet. %d van de 10 kritieke checks zijn groen. De rest vormt een direct risico voor uw markttoegang in de EU."
		</div>
		
		<div class="content">
			<h2 class="section-title">Analyse Resultaat</h2>
			<div style="margin-bottom: 30px; color: #475569;">
				%s
			</div>

			<h2 class="section-title">Aanbevolen Volgende Stappen</h2>
			<div class="cta-container">
				<a href="https://www.leonacompliance.be/contact" class="button">
					Start Snapshot Audit (€2.495)
				</a>
				<a href="https://www.leonacompliance.be/demo" class="button button-outline">
					Bekijk de Compliance Pipeline Demo
				</a>
			</div>

			<h2 class="section-title">Gedetailleerde Audit Matrix</h2>
			<table class="results-table">
				%s
			</table>
		</div>

		<div class="deadline-footer">
			<span style="text-transform: uppercase; font-size: 12px; letter-spacing: 2px;">Deadline: 11 September 2026</span>
			<span class="months-left">%d Maanden Resterend</span>
			<p style="font-size: 13px; opacity: 0.8; margin-top: 10px;">Vanaf deze datum is CRA compliance verplicht voor verkoop in de EU.</p>
		</div>

		<div class="footer-legal">
			<p><strong>LEONA Compliance</strong> | Wetenschapstraat 14, 1040 Brussel, België<br/>
			<a href="mailto:expert@leonacompliance.be">expert@leonacompliance.be</a> | <a href="https://leonacompliance.be">leonacompliance.be</a></p>
            
            <div style="margin-top: 20px; border-top: 1px solid #f1f5f9; padding-top: 20px;">
                U ontvangt dit rapport naar aanleiding van uw deelname aan de CRA Assessment. Wij verwerken uw data conform onze privacy policy.<br><br>
                <a href="https://www.leonacompliance.be/privacy">Privacy Policy</a>
            </div>
		</div>
	</div>
</body>
</html>`, reportID, score, riskLevel, jaCount, recommendation, questionResults, calculateMonthsUntilDeadline())

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
