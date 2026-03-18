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
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	answersJSON := r.FormValue("answers")

	// Validate email
	if email == "" || !strings.Contains(email, "@") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Geldig e-mailadres vereist"})
		return
	}

	// Parse answers
	var answers map[string]string
	if err := json.Unmarshal([]byte(answersJSON), &answers); err != nil {
		log.Printf("Failed to parse answers: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ongeldige antwoorden"})
		return
	}

	// Calculate score
	jaCount := 0
	for _, answer := range answers {
		if answer == "ja" {
			jaCount++
		}
	}
	complianceScore := (jaCount * 100) / len(answers)

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

	// Determine risk level and recommendations (uplifting tone)
	var riskLevel, riskColor, recommendation, encouragement string
	if score >= 80 {
		riskLevel = "UITSTEKEND"
		riskColor = "#22c55e"
		recommendation = "Fantastisch werk! Uw systeem is al grotendeels CRA-compliant. Met enkele kleine aanpassingen bent u helemaal klaar voor de deadline."
		encouragement = "U bent op de goede weg! De basis staat stevig en met onze hulp maakt u het af."
	} else if score >= 50 {
		riskLevel = "GOED BEZIG"
		riskColor = "#f59e0b"
		recommendation = "U heeft al een solide basis gelegd! Er zijn nog enkele punten die aandacht nodig hebben, maar met gerichte actie bent u op tijd klaar."
		encouragement = "U heeft de juiste stappen al gezet. Laten we samen de resterende hiaten aanpakken!"
	} else {
		riskLevel = "MOGELIJKHEID VOOR GROEI"
		riskColor = "#3b82f6"
		recommendation = "Geen zorgen - u heeft deze assessment gedaan en dat is al een belangrijke eerste stap! We helpen bedrijven dagelijks om van nul naar volledig compliant te gaan. Samen maken we dit haalbaar."
		encouragement = "Elke reis begint met de eerste stap - en die heeft u nu gezet! Wij begeleiden u naar volledige compliance."
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

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", fmt.Sprintf("Uw CRA Compliance Assessment Resultaten - %d%%", score))

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: system-ui, sans-serif; line-height: 1.6; color: #1a1a1a; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(135deg, #1e40af 0%%, #3b82f6 100%%); color: white; padding: 30px; border-radius: 8px; text-align: center; }
		.score-badge { background: %s; color: white; padding: 20px; border-radius: 12px; text-align: center; margin: 30px 0; }
		.score-number { font-size: 48px; font-weight: bold; margin: 10px 0; }
		.content { background: #f9fafb; padding: 30px; margin-top: 20px; border-radius: 8px; }
		.results-table { width: 100%%; border-collapse: collapse; margin: 20px 0; background: white; border-radius: 8px; overflow: hidden; }
		.button { display: inline-block; background: #1e40af; color: white; padding: 14px 30px; text-decoration: none; border-radius: 6px; margin: 20px 0; font-weight: bold; }
		.footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
		.warning { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1 style="margin: 0;">Uw CRA Assessment Resultaten</h1>
			<p style="margin: 10px 0 0 0; opacity: 0.9;">Bedankt voor het invullen! Hier is uw persoonlijke score.</p>
		</div>
		
		<div class="score-badge">
			<div style="font-size: 16px; opacity: 0.9;">Uw Compliance Score</div>
			<div class="score-number">%d%%</div>
			<div style="font-size: 18px; font-weight: bold; margin-top: 10px;">%s</div>
			<div style="font-size: 14px; margin-top: 5px; opacity: 0.9;">%d van de 10 vragen positief beantwoord</div>
		</div>
		
		<div style="background: #ecfdf5; padding: 20px; border-radius: 8px; margin: 20px 0; text-align: center; border: 2px solid #10b981;">
			<p style="margin: 0; font-size: 16px; color: #047857; font-weight: 600;">%s</p>
		</div>
		
		<div class="content">
			<h2 style="color: #1e40af; margin-top: 0;">Uw Antwoorden</h2>
			<table class="results-table">
				%s
			</table>

			<div style="background: #eff6ff; border-left: 4px solid #3b82f6; padding: 15px; margin: 20px 0; border-radius: 4px;">
				<strong>Onze Aanbeveling</strong><br/>
				%s
			</div>

			<h3 style="color: #1e40af;">Volgende Stappen</h3>
			<ol style="line-height: 2;">
				<li><strong>Download uw volledige assessment rapport</strong> (PDF met aanbevelingen per vraag)</li>
				<li><strong>Plan een gratis 30-minuten consultatie</strong> met onze CRA experts</li>
				<li><strong>Vraag een offerte aan</strong> voor gap analysis + implementatie</li>
			</ol>

			<div style="text-align: center;">
				<a href="https://leonacompliance.be/cra-assessment" class="button">Download Volledig Rapport</a>
				<a href="https://calendly.com/leonacompliance/cra-consult" class="button" style="background: #22c55e;">Plan Gratis Consult</a>
			</div>

			<h3 style="color: #1e40af; margin-top: 40px;">Belangrijke Deadline</h3>
			<p style="background: #fee2e2; padding: 15px; border-radius: 6px; border-left: 4px solid #ef4444;">
				<strong>11 September 2026</strong> - vanaf deze datum moet uw embedded product volledig CRA-compliant zijn om in de EU te mogen verkopen. U heeft nog <strong>%d maanden</strong> om uw systeem compliant te maken.
			</p>
		</div>

		<div class="footer">
			<p><strong>LEONA Compliance</strong> | CRA Compliance Engineering<br/>
			Vragen? Reply op deze email of bezoek <a href="https://leonacompliance.be">leonacompliance.be</a></p>
		</div>
	</div>
</body>
</html>
`, riskColor, score, riskLevel, jaCount, encouragement, questionResults, recommendation, calculateMonthsUntilDeadline())

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
