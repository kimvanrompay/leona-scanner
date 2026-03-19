package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"leona-scanner/internal/database"

	"gopkg.in/gomail.v2"
)

// LegalAssessmentSubmit handles the legal assessment form submission
func LegalAssessmentSubmit(db *database.SupabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Parse form data
		name := strings.TrimSpace(r.FormValue("name"))
		lawFirm := strings.TrimSpace(r.FormValue("law-firm"))
		emailAddr := strings.TrimSpace(r.FormValue("email"))
		scoreStr := r.FormValue("score")
		answersJSON := r.FormValue("answers")

		// Validate required fields
		if name == "" || lawFirm == "" || emailAddr == "" || scoreStr == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Parse answers
		var answers []int
		if err := json.Unmarshal([]byte(answersJSON), &answers); err != nil {
			http.Error(w, "Invalid answers format", http.StatusBadRequest)
			return
		}

		// Store in database as lead (if db is available)
		if db != nil {
			notes := fmt.Sprintf("CRA Legal Assessment - Score: %s/100", scoreStr)
			lead := &database.Lead{
				Email:       emailAddr,
				FirstName:   &name,
				CompanyName: &lawFirm,
				Notes:       &notes,
				LeadType:    "legal-assessment",
				Source:      "website",
				Status:      "new",
			}

			if err := db.CreateLead(r.Context(), lead); err != nil {
				log.Printf("Error saving assessment lead: %v\n", err)
				// Continue anyway - don't block user
			}
		}

		// Send email notification to kim@leonacompliance.be
		notificationSubject := fmt.Sprintf("🎯 Nieuwe CRA Legal Assessment - %s/100 punten", scoreStr)
		notificationBody := buildNotificationEmail(name, lawFirm, emailAddr, scoreStr, answers)

		if err := sendAssessmentEmail("kim@leonacompliance.be", notificationSubject, notificationBody); err != nil {
			log.Printf("❌ ERROR: Failed to send notification email: %v\n", err)
		} else {
			log.Printf("✅ SUCCESS: Assessment notification sent for %s (%s)", name, lawFirm)
		}

		// Send confirmation email to the lawyer with detailed analysis
		confirmationSubject := "Uw CRA Legal-Tech Gap Assessment Rapport"
		confirmationBody := buildConfirmationEmail(name, scoreStr, answers)

		if err := sendAssessmentEmail(emailAddr, confirmationSubject, confirmationBody); err != nil {
			log.Printf("⚠️  WARNING: Failed to send confirmation email to %s: %v\n", emailAddr, err)
		} else {
			log.Printf("📧 Confirmation email sent to %s", emailAddr)
		}

		// Return success message
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<div class="rounded-lg bg-green-50 border border-green-200 p-4">
				<p class="text-sm font-semibold text-green-800">✓ Rapport verzonden!</p>
				<p class="text-sm text-green-700 mt-1">U ontvangt uw volledige analyse binnen enkele minuten op ` + emailAddr + `</p>
			</div>
			<script>
				// Reset form after 5 seconds
				setTimeout(() => {
					const form = document.querySelector('form');
					if (form) form.reset();
				}, 5000);
			</script>
		`))
	}
}

// sendAssessmentEmail sends HTML email using SMTP
func sendAssessmentEmail(to, subject, htmlBody string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 465
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := "support@leonacompliance.be"

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("email service not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	d.SSL = true

	return d.DialAndSend(m)
}

func buildNotificationEmail(name, lawFirm, email, score string, answers []int) string {
	answerDetails := buildAnswerSummary(answers)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f9fafb;">
	<table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f9fafb; padding: 40px 20px;">
		<tr>
			<td align="center">
				<table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
					
					<!-- Header -->
					<tr>
						<td style="background: linear-gradient(135deg, #1e40af 0%%, #3b82f6 100%%); padding: 40px 40px 30px 40px; text-align: center;">
							<h1 style="margin: 0; color: #ffffff; font-size: 24px; font-weight: bold;">
								🎯 Nieuwe CRA Legal Assessment
							</h1>
						</td>
					</tr>

					<!-- Score Badge -->
					<tr>
						<td style="padding: 30px 40px; text-align: center; background-color: #f9fafb;">
							<div style="display: inline-block; background-color: %s; color: white; font-size: 48px; font-weight: bold; width: 120px; height: 120px; border-radius: 60px; line-height: 120px; margin-bottom: 15px;">
								%s
							</div>
							<p style="margin: 0; font-size: 18px; font-weight: bold; color: #111827;">
								%s
							</p>
						</td>
					</tr>

					<!-- Contact Details -->
					<tr>
						<td style="padding: 0 40px 30px 40px;">
							<table width="100%%" cellpadding="0" cellspacing="0">
								<tr>
									<td style="padding: 15px; background-color: #f9fafb; border-left: 4px solid #3b82f6;">
										<p style="margin: 0 0 5px 0; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em;">Naam</p>
										<p style="margin: 0; font-size: 16px; color: #111827; font-weight: 600;">%s</p>
									</td>
								</tr>
								<tr><td style="height: 10px;"></td></tr>
								<tr>
									<td style="padding: 15px; background-color: #f9fafb; border-left: 4px solid #3b82f6;">
										<p style="margin: 0 0 5px 0; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em;">Advocatenkantoor</p>
										<p style="margin: 0; font-size: 16px; color: #111827; font-weight: 600;">%s</p>
									</td>
								</tr>
								<tr><td style="height: 10px;"></td></tr>
								<tr>
									<td style="padding: 15px; background-color: #f9fafb; border-left: 4px solid #3b82f6;">
										<p style="margin: 0 0 5px 0; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em;">Email</p>
										<p style="margin: 0; font-size: 16px; color: #111827; font-weight: 600;">
											<a href="mailto:%s" style="color: #2563eb; text-decoration: none;">%s</a>
										</p>
									</td>
								</tr>
							</table>
						</td>
					</tr>

					<!-- Answer Summary -->
					<tr>
						<td style="padding: 0 40px 40px 40px;">
							<h2 style="margin: 0 0 20px 0; font-size: 18px; color: #111827; font-weight: bold;">Antwoorddetails</h2>
							%s
						</td>
					</tr>

					<!-- Footer -->
					<tr>
						<td style="padding: 30px 40px; background-color: #f9fafb; text-align: center; border-top: 1px solid #e5e7eb;">
							<p style="margin: 0; font-size: 14px; color: #6b7280;">
								LEONA Scanner · CRA Compliance Platform<br>
								<a href="https://leonacompliance.be" style="color: #2563eb; text-decoration: none;">leonacompliance.be</a>
							</p>
						</td>
					</tr>

				</table>
			</td>
		</tr>
	</table>
</body>
</html>
	`, getScoreColorHex(score), score, getScoreLabel(score), name, lawFirm, email, email, answerDetails)
}

func buildConfirmationEmail(name, score string, answers []int) string {
	detailedAnalysis := buildDetailedAnalysis(answers)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f9fafb;">
	<table width="100%%" cellpadding="0" cellspacing="0" style="background-color: #f9fafb; padding: 40px 20px;">
		<tr>
			<td align="center">
				<table width="600" cellpadding="0" cellspacing="0" style="background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
					
					<!-- Header -->
					<tr>
						<td style="background: linear-gradient(135deg, #1e40af 0%%, #3b82f6 100%%); padding: 40px 40px 30px 40px;">
							<h1 style="margin: 0 0 10px 0; color: #ffffff; font-size: 24px; font-weight: bold;">
								Uw CRA Legal-Tech Gap Assessment
							</h1>
							<p style="margin: 0; color: #dbeafe; font-size: 16px;">
								Persoonlijk rapport voor %s
							</p>
						</td>
					</tr>

					<!-- Score Section -->
					<tr>
						<td style="padding: 30px 40px; text-align: center; background-color: #f9fafb;">
							<div style="display: inline-block; background-color: %s; color: white; font-size: 48px; font-weight: bold; width: 120px; height: 120px; border-radius: 60px; line-height: 120px; margin-bottom: 15px;">
								%s
							</div>
							<p style="margin: 0 0 10px 0; font-size: 20px; font-weight: bold; color: #111827;">
								%s
							</p>
							<p style="margin: 0; font-size: 14px; color: #6b7280; max-width: 500px; margin-left: auto; margin-right: auto; line-height: 1.6;">
								%s
							</p>
						</td>
					</tr>

					<!-- Detailed Analysis -->
					<tr>
						<td style="padding: 0 40px 40px 40px;">
							<h2 style="margin: 0 0 20px 0; font-size: 18px; color: #111827; font-weight: bold;">Uw Antwoorden & Aanbevelingen</h2>
							%s
						</td>
					</tr>

					<!-- CTA Section -->
					<tr>
						<td style="padding: 30px 40px; background-color: #eff6ff; border-top: 1px solid #dbeafe; border-bottom: 1px solid #dbeafe;">
							<h3 style="margin: 0 0 15px 0; font-size: 18px; color: #111827; font-weight: bold; text-align: center;">
								Interesse in een strategisch overleg?
							</h3>
							<p style="margin: 0 0 20px 0; font-size: 14px; color: #4b5563; text-align: center; line-height: 1.6;">
								Wij bespreken hoe LEONA uw juridische praktijk kan ondersteunen met binaire bewijslast en geautomatiseerde compliance-monitoring.
							</p>
							<table width="100%%" cellpadding="0" cellspacing="0">
								<tr>
									<td align="center">
										<a href="https://leonacompliance.be/partner-overleg" style="display: inline-block; background-color: #2563eb; color: #ffffff; text-decoration: none; padding: 14px 28px; border-radius: 6px; font-weight: 600; font-size: 16px;">
											Plan een Partner Overleg
										</a>
									</td>
								</tr>
							</table>
						</td>
					</tr>

					<!-- Footer -->
					<tr>
						<td style="padding: 30px 40px; background-color: #f9fafb; text-align: center;">
							<p style="margin: 0 0 10px 0; font-size: 14px; color: #6b7280;">
								LEONA Scanner · CRA Compliance Platform
							</p>
							<p style="margin: 0; font-size: 14px; color: #6b7280;">
								<a href="https://leonacompliance.be" style="color: #2563eb; text-decoration: none;">leonacompliance.be</a> · 
								<a href="https://leonacompliance.be/legal-partners" style="color: #2563eb; text-decoration: none;">Legal Partnership</a> · 
								<a href="https://leonacompliance.be/specs" style="color: #2563eb; text-decoration: none;">Technical Specs</a>
							</p>
						</td>
					</tr>

				</table>
			</td>
		</tr>
	</table>
</body>
</html>
	`, name, getScoreColorHex(score), score, getScoreLabel(score), getScoreDescription(score), detailedAnalysis)
}

func buildAnswerSummary(answers []int) string {
	questions := []string{
		"1. De Binaire Waarheid (Art. 10)",
		"2. De 24-uurs Triage (Art. 11, Lid 1)",
		"3. Bewijs van Vulnerability Handling",
		"4. Annex I: Default Security Mapping",
		"5. De 10-jarige Bewaarplicht",
		"6. Supply Chain Liability (Art. 13)",
		"7. Security Updates & Transparantie",
		"8. De 'Appropriate Level' Toets",
		"9. Product Lifecycle Support",
		"10. Persoonlijke Bestuurdersaansprakelijkheid",
	}

	var summary strings.Builder
	summary.WriteString(`<table width="100%%" cellpadding="0" cellspacing="0" style="font-size: 14px;">`)

	for i, answer := range answers {
		bgColor := "#fee"
		if answer >= 5 {
			bgColor = "#fef3c7"
		}
		if answer >= 10 {
			bgColor = "#d1fae5"
		}

		summary.WriteString(fmt.Sprintf(`
			<tr>
				<td style="padding: 10px; background-color: %s; border-bottom: 1px solid #e5e7eb;">
					<strong style="color: #111827;">%s</strong>
				</td>
				<td style="padding: 10px; background-color: %s; border-bottom: 1px solid #e5e7eb; text-align: right;">
					<strong style="color: #111827;">%d/10 punten</strong>
				</td>
			</tr>
		`, bgColor, questions[i], bgColor, answer))
	}

	summary.WriteString(`</table>`)
	return summary.String()
}

func buildDetailedAnalysis(answers []int) string {
	questions := []struct {
		title          string
		recommendation string
	}{
		{
			"De Binaire Waarheid (Art. 10)",
			"Voor juridisch verdedigbare SBOM's heeft u geautomatiseerde extractie nodig die ook kernel configurations en transitive dependencies detecteert.",
		},
		{
			"De 24-uurs Triage (Art. 11, Lid 1)",
			"Real-time CVE-matching met exploitability validation is essentieel om binnen de wettelijke deadlines te blijven zonder reputatierisico.",
		},
		{
			"Bewijs van Vulnerability Handling",
			"Historische binaire scandata met immutable timestamps vormt uw juridische 'paper trail' bij een security incident.",
		},
		{
			"Annex I: Default Security Mapping",
			"Binaire configuratiescans tonen aan dat security requirements niet alleen contractueel maar ook technisch zijn geïmplementeerd.",
		},
		{
			"De 10-jarige Bewaarplicht",
			"Een onveranderlijk archief met binaire artifacts en compliance snapshots is wettelijk verplicht onder Art. 10, Lid 8.",
		},
		{
			"Supply Chain Liability (Art. 13)",
			"Supply chain verification met cryptografische hashes beschermt u tegen upstream liability.",
		},
		{
			"Security Updates & Transparantie",
			"Voor/na integriteitsscans bij updates documenteren dat patches geen nieuwe risico's introduceren.",
		},
		{
			"De 'Appropriate Level' Toets",
			"Benchmark-rapporten tegen NIST/CIS standaarden maken 'gepaste beveiliging' objectief meetbaar.",
		},
		{
			"Product Lifecycle Support",
			"Continue CVE-monitoring gedurende de productlevensduur maakt SLA's juridisch afdwingbaar.",
		},
		{
			"Persoonlijke Bestuurdersaansprakelijkheid",
			"Volledige transparantie met bestuurders over technical gaps en een concreet mitigatieplan is uw beroepsaansprakelijkheidsverzekering.",
		},
	}

	var analysis strings.Builder

	for i, answer := range answers {
		status := "❌ Hoog Risico"
		statusColor := "#dc2626"
		recommendation := "Direct actie vereist: " + questions[i].recommendation

		if answer >= 5 {
			status = "⚠️ Gedeeltelijke Dekking"
			statusColor = "#f59e0b"
			recommendation = "Verbetering mogelijk: " + questions[i].recommendation
		}
		if answer >= 10 {
			status = "✅ Compliant"
			statusColor = "#059669"
			recommendation = "Uitstekend: Uw proces voldoet aan CRA-vereisten."
		}

		analysis.WriteString(fmt.Sprintf(`
			<div style="margin-bottom: 20px; padding: 15px; background-color: #f9fafb; border-left: 4px solid %s; border-radius: 4px;">
				<p style="margin: 0 0 5px 0; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em;">Vraag %d</p>
				<h3 style="margin: 0 0 8px 0; font-size: 16px; color: #111827; font-weight: bold;">%s</h3>
				<p style="margin: 0 0 8px 0; font-size: 14px; color: %s; font-weight: 600;">%s · %d/10 punten</p>
				<p style="margin: 0; font-size: 14px; color: #4b5563; line-height: 1.6;">%s</p>
			</div>
		`, statusColor, i+1, questions[i].title, statusColor, status, answer, recommendation))
	}

	return analysis.String()
}

func getScoreColorHex(score string) string {
	// Parse score as int
	var scoreInt int
	fmt.Sscanf(score, "%d", &scoreInt)

	if scoreInt <= 40 {
		return "#dc2626" // red-600
	}
	if scoreInt <= 80 {
		return "#f59e0b" // orange-500
	}
	return "#059669" // green-600
}

func getScoreLabel(score string) string {
	var scoreInt int
	fmt.Sscanf(score, "%d", &scoreInt)

	if scoreInt <= 40 {
		return "Technisch Blind"
	}
	if scoreInt <= 80 {
		return "Juridisch Sterk, Technisch Kwetsbaar"
	}
	return "CRA-Ready"
}

func getScoreDescription(score string) string {
	var scoreInt int
	fmt.Sscanf(score, "%d", &scoreInt)

	if scoreInt <= 40 {
		return "Uw juridische dossiers rusten op ongeteste aannames. U heeft geen technische verificatie van de conformiteitsverklaringen die u opstelt. Dit creëert een hoog risico op beroepsaansprakelijkheid wanneer een incident de binaire realiteit blootlegt."
	}
	if scoreInt <= 80 {
		return "U begrijpt de CRA-kaders en heeft processen ingericht, maar mist de automatisering om de 24-uurs deadlines en 10-jarige bewaarplicht te borgen. Uw cliënten zijn kwetsbaar voor compliance-inbreuken die juridisch correct lijken maar technisch niet haalbaar zijn."
	}
	return "U beschikt over een volledige technische infrastructuur om de CRA uit te voeren. Uw juridische adviezen zijn onderbouwd met binaire bewijslast. (Dit niveau haalt niemand zonder gespecialiseerde legal-tech zoals LEONA.)"
}
