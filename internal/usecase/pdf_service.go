package usecase

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"

	"leona-scanner/internal/scanner"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

// GeneratePDF creates a professional CRA compliance report
func (s *PDFService) GeneratePDF(result *scanner.AnalysisResult, platform string) ([]byte, error) {
	cfg := config.NewBuilder().
		WithMargins(10, 15, 10).
		Build()

	m := maroto.New(cfg)

	// Header
	m.AddRows(
		row.New(20).Add(
			col.New(12).Add(
				text.New("LEONA CRA Compliance Rapport", props.Text{
					Size:  20,
					Style: fontstyle.Bold,
					Align: align.Center,
					Color: &props.Color{Red: 255, Green: 69, Blue: 0}, // Deep Orange
				}),
			),
		),
		row.New(10).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("Platform: %s | Datum: %s", platform, time.Now().Format("02-01-2006")), props.Text{
					Size:  10,
					Align: align.Center,
				}),
			),
		),
	)

	// Overall Score Section
	statusColor := getStatusColor(result.Status)
	m.AddRows(
		row.New(15).Add(
			col.New(12).Add(
				text.New("OVERALL COMPLIANCE SCORE", props.Text{
					Size:  14,
					Style: fontstyle.Bold,
				}),
			),
		),
		row.New(20).Add(
			col.New(6).Add(
				text.New(fmt.Sprintf("Score: %d/100", result.OverallScore), props.Text{
					Size:  16,
					Style: fontstyle.Bold,
				}),
			),
			col.New(6).Add(
				text.New(fmt.Sprintf("Status: %s", result.Status), props.Text{
					Size:  16,
					Style: fontstyle.Bold,
					Color: statusColor,
				}),
			),
		),
	)

	// Summary
	m.AddRow(10,
		col.New(12).Add(
			text.New("SAMENVATTING", props.Text{
				Size:  12,
				Style: fontstyle.Bold,
			}),
		),
	)
	m.AddRow(8,
		col.New(12).Add(
			text.New(fmt.Sprintf("Totaal componenten geanalyseerd: %d", result.TotalComponents), props.Text{
				Size: 10,
			}),
		),
	)
	m.AddRow(8,
		col.New(12).Add(
			text.New(fmt.Sprintf("Aantal problemen gevonden: %d", len(result.Issues)), props.Text{
				Size: 10,
			}),
		),
	)

	// Issues by Severity
	criticalCount, highCount, mediumCount, lowCount := 0, 0, 0, 0
	for _, issue := range result.Issues {
		switch issue.Severity {
		case "CRITICAL":
			criticalCount++
		case "HIGH":
			highCount++
		case "MEDIUM":
			mediumCount++
		case "LOW":
			lowCount++
		}
	}

	m.AddRow(15,
		col.New(12).Add(
			text.New("PROBLEMEN PER ERNST", props.Text{
				Size:  12,
				Style: fontstyle.Bold,
			}),
		),
	)
	m.AddRows(
		row.New(6).Add(
			col.New(3).Add(text.New(fmt.Sprintf("🔴 CRITICAL: %d", criticalCount), props.Text{Size: 9})),
			col.New(3).Add(text.New(fmt.Sprintf("🟠 HIGH: %d", highCount), props.Text{Size: 9})),
			col.New(3).Add(text.New(fmt.Sprintf("🟡 MEDIUM: %d", mediumCount), props.Text{Size: 9})),
			col.New(3).Add(text.New(fmt.Sprintf("🟢 LOW: %d", lowCount), props.Text{Size: 9})),
		),
	)

	// Detailed Issues
	if len(result.Issues) > 0 {
		m.AddRow(15,
			col.New(12).Add(
				text.New("GEDETAILLEERDE BEVINDINGEN", props.Text{
					Size:  12,
					Style: fontstyle.Bold,
				}),
			),
		)

		for i, issue := range result.Issues {
			if i >= 50 { // Limit to first 50 issues for PDF size
				m.AddRow(8,
					col.New(12).Add(
						text.New(fmt.Sprintf("... en %d meer problemen", len(result.Issues)-50), props.Text{
							Size:  9,
							Style: fontstyle.Italic,
						}),
					),
				)
				break
			}

			severityEmoji := getSeverityEmoji(issue.Severity)
			m.AddRow(10,
				col.New(12).Add(
					text.New(fmt.Sprintf("%s [%s] %s", severityEmoji, issue.Severity, issue.Component), props.Text{
						Size:  9,
						Style: fontstyle.Bold,
					}),
				),
			)
			m.AddRow(6,
				col.New(12).Add(
					text.New(issue.Description, props.Text{
						Size: 8,
					}),
				),
			)
			if issue.Recommendation != "" {
				m.AddRow(6,
					col.New(12).Add(
						text.New(fmt.Sprintf("→ %s", issue.Recommendation), props.Text{
							Size:  8,
							Style: fontstyle.Italic,
						}),
					),
				)
			}
			m.AddRow(4)
		}
	}

	// Recommendations
	m.AddRow(15,
		col.New(12).Add(
			text.New("AANBEVELINGEN", props.Text{
				Size:  12,
				Style: fontstyle.Bold,
			}),
		),
	)

	recommendations := []string{
		"Zorg dat alle componenten voorzien zijn van CPE identifiers voor traceerbaarheid",
		"Gebruik duidelijke versienummering voor alle dependencies",
		"Vermijd GPL-3.0 licenties in embedded systemen vanwege anti-tivoization clausules",
		"Implementeer een proces voor regelmatige security updates",
		"Documenteer alle third-party componenten en hun licenties",
	}

	for _, rec := range recommendations {
		m.AddRow(8,
			col.New(12).Add(
				text.New(fmt.Sprintf("• %s", rec), props.Text{
					Size: 9,
				}),
			),
		)
	}

	// Footer
	m.AddRow(20,
		col.New(12).Add(
			text.New("Dit rapport is gegenereerd door LEONA - CRA Compliance Scanner", props.Text{
				Size:  8,
				Style: fontstyle.Italic,
				Align: align.Center,
			}),
		),
	)
	m.AddRow(6,
		col.New(12).Add(
			text.New(fmt.Sprintf("Rapportage deadline: September 2026 | Gegenereerd: %s", time.Now().Format("02-01-2006 15:04")), props.Text{
				Size:  7,
				Align: align.Center,
			}),
		),
	)

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("PDF generatie fout: %w", err)
	}

	return document.GetBytes(), nil
}

// SendPDF sends the PDF report via email
func (s *PDFService) SendPDF(toEmail string, pdfData []byte, scanID string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP configuratie ontbreekt")
	}

	// Build email with PDF attachment
	var emailBody bytes.Buffer
	boundary := "boundary123456789"

	emailBody.WriteString(fmt.Sprintf("From: LEONA CRA Scanner <%s>\r\n", fromEmail))
	emailBody.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	emailBody.WriteString("Subject: Uw CRA Compliance Rapport\r\n")
	emailBody.WriteString("MIME-Version: 1.0\r\n")
	emailBody.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))

	// Email body text
	emailBody.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	emailBody.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	emailBody.WriteString("Beste,\n\n")
	emailBody.WriteString("Bedankt voor uw aankoop van het LEONA CRA Compliance Rapport.\n\n")
	emailBody.WriteString(fmt.Sprintf("In de bijlage vindt u het volledige compliance rapport voor scan ID: %s\n\n", scanID))
	emailBody.WriteString("Met vriendelijke groet,\n")
	emailBody.WriteString("Het LEONA Team\n\n")
	emailBody.WriteString("---\n")
	emailBody.WriteString("CRA deadline: September 2026\n")
	emailBody.WriteString(fmt.Sprintf("--%s\r\n", boundary))

	// PDF attachment
	emailBody.WriteString("Content-Type: application/pdf\r\n")
	emailBody.WriteString("Content-Transfer-Encoding: base64\r\n")
	emailBody.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"CRA-Report-%s.pdf\"\r\n\r\n", scanID))

	// Base64 encode PDF
	encoded := make([]byte, len(pdfData)*2)
	base64Encode(pdfData, encoded)

	// Write in chunks of 76 characters (RFC 2045)
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		emailBody.Write(encoded[i:end])
		emailBody.WriteString("\r\n")
	}

	emailBody.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	// Send email
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpHost, smtpPort),
		auth,
		fromEmail,
		[]string{toEmail},
		emailBody.Bytes(),
	)

	if err != nil {
		return fmt.Errorf("email verzenden fout: %w", err)
	}

	return nil
}

func getStatusColor(status string) *props.Color {
	switch strings.ToUpper(status) {
	case "CONFORM":
		return &props.Color{Red: 0, Green: 200, Blue: 0}
	case "VOORWAARDELIJK CONFORM":
		return &props.Color{Red: 255, Green: 165, Blue: 0}
	default:
		return &props.Color{Red: 255, Green: 0, Blue: 0}
	}
}

func getSeverityEmoji(severity string) string {
	switch severity {
	case "CRITICAL":
		return "🔴"
	case "HIGH":
		return "🟠"
	case "MEDIUM":
		return "🟡"
	case "LOW":
		return "🟢"
	default:
		return "⚪"
	}
}

// Simple base64 encoding (standard encoding)
func base64Encode(src []byte, dst []byte) {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	di, si := 0, 0
	n := (len(src) / 3) * 3
	for si < n {
		val := uint(src[si])<<16 | uint(src[si+1])<<8 | uint(src[si+2])

		dst[di] = encodeStd[val>>18&0x3F]
		dst[di+1] = encodeStd[val>>12&0x3F]
		dst[di+2] = encodeStd[val>>6&0x3F]
		dst[di+3] = encodeStd[val&0x3F]

		si += 3
		di += 4
	}

	remain := len(src) - si
	if remain == 0 {
		return
	}

	val := uint(src[si]) << 16
	if remain == 2 {
		val |= uint(src[si+1]) << 8
	}

	dst[di] = encodeStd[val>>18&0x3F]
	dst[di+1] = encodeStd[val>>12&0x3F]

	if remain == 2 {
		dst[di+2] = encodeStd[val>>6&0x3F]
		dst[di+3] = '='
	} else {
		dst[di+2] = '='
		dst[di+3] = '='
	}
}
