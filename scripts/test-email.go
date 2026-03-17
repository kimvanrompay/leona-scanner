// Package main provides a test utility for verifying SMTP email configuration.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

func getEmailBody(smtpHost, smtpFrom string) string {
	return `
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: system-ui, sans-serif; line-height: 1.6; color: #1a1a1a; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: linear-gradient(135deg, #1e40af 0%, #3b82f6 100%); color: white; padding: 30px; border-radius: 8px; text-align: center; }
		.content { background: #f9fafb; padding: 30px; margin-top: 20px; border-radius: 8px; }
		.success { background: #d1fae5; border-left: 4px solid #10b981; padding: 15px; margin: 20px 0; border-radius: 4px; color: #065f46; }
		.footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; text-align: center; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1 style="margin: 0;">✅ Email Configuration Test</h1>
			<p style="margin: 10px 0 0 0; opacity: 0.9;">LEONA Compliance SMTP Setup</p>
		</div>
		
		<div class="content">
			<div class="success">
				<strong>✅ Success!</strong><br/>
				Your SMTP configuration is working correctly.
			</div>

			<h3 style="color: #1e40af;">Configuration Details</h3>
			<ul style="line-height: 2;">
				<li><strong>SMTP Host:</strong> ` + smtpHost + `</li>
				<li><strong>Port:</strong> 465 (SSL/TLS)</li>
				<li><strong>From Address:</strong> ` + smtpFrom + `</li>
				<li><strong>Authentication:</strong> ✅ Successful</li>
			</ul>

			<h3 style="color: #1e40af; margin-top: 40px;">Next Steps</h3>
			<ol style="line-height: 2;">
				<li>✅ SMTP configuration is complete</li>
				<li>Test the CRA assessment form at <code>/cra-assessment</code></li>
				<li>Test the contact form at <code>/contact</code></li>
			</ol>
		</div>

		<div class="footer">
			<p><strong>LEONA Compliance</strong> | CRA Compliance Engineering<br/>
			<a href="https://leonacompliance.be">leonacompliance.be</a> | <a href="mailto:support@leonacompliance.be">support@leonacompliance.be</a></p>
		</div>
	</div>
</body>
</html>
`
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	// Get SMTP configuration
	smtpHost := os.Getenv("SMTP_HOST")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		log.Fatal("❌ SMTP configuration missing. Check your .env file.")
	}

	fmt.Println("📧 Testing SMTP Configuration...")
	fmt.Printf("   Host: %s\n", smtpHost)
	fmt.Printf("   Port: 465 (SSL/TLS)\n")
	fmt.Printf("   User: %s\n", smtpUser)
	fmt.Printf("   From: %s\n", smtpFrom)
	fmt.Println()

	// Create and send test email
	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", "kim@eliama.agency")
	m.SetHeader("Subject", "🧪 LEONA Email Test - Configuration Successful")
	m.SetBody("text/html", getEmailBody(smtpHost, smtpFrom))

	fmt.Println("📤 Sending test email...")
	d := gomail.NewDialer(smtpHost, 465, smtpUser, smtpPass)
	d.SSL = true

	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("❌ Failed to send email: %v", err)
	}

	fmt.Println("✅ Test email sent successfully!")
	fmt.Println("📬 Check kim@eliama.agency inbox")
}
