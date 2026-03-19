package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

// MailgunService handles email sending via Mailgun
type MailgunService struct {
	mg     *mailgun.MailgunImpl
	domain string
	from   string
}

// NewMailgunService creates a new Mailgun email service
func NewMailgunService() *MailgunService {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")
	from := os.Getenv("MAILGUN_FROM")

	if apiKey == "" || domain == "" {
		log.Println("⚠️  Mailgun not configured - emails will not be sent")
		return nil
	}

	if from == "" {
		from = "support@leonacompliance.be"
	}

	mg := mailgun.NewMailgun(domain, apiKey)

	// If using EU domain, set EU endpoint
	region := os.Getenv("MAILGUN_REGION")
	if region == "EU" {
		mg.SetAPIBase("https://api.eu.mailgun.net")
	}

	log.Printf("✅ Mailgun service initialized (domain: %s, region: %s)", domain, region)

	return &MailgunService{
		mg:     mg,
		domain: domain,
		from:   from,
	}
}

// SendHTMLEmail sends an HTML email via Mailgun
func (m *MailgunService) SendHTMLEmail(to, subject, htmlBody string) error {
	if m == nil {
		return fmt.Errorf("mailgun not configured")
	}

	message := mailgun.NewMessage(m.from, subject, "", to)
	message.SetHTML(htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := m.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("✅ Email sent via Mailgun (ID: %s, to: %s)", id, to)
	return nil
}

// SendHTMLEmailWithAttachment sends an HTML email with a file attachment
func (m *MailgunService) SendHTMLEmailWithAttachment(to, subject, htmlBody, attachmentPath string) error {
	if m == nil {
		return fmt.Errorf("mailgun not configured")
	}

	message := mailgun.NewMessage(m.from, subject, "", to)
	message.SetHTML(htmlBody)

	// Add attachment if path is provided
	if attachmentPath != "" {
		message.AddAttachment(attachmentPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := m.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email with attachment: %w", err)
	}

	log.Printf("✅ Email with attachment sent via Mailgun (ID: %s, to: %s)", id, to)
	return nil
}

// SendHTMLEmailWithBCC sends an HTML email with BCC
func (m *MailgunService) SendHTMLEmailWithBCC(to, bcc, subject, htmlBody string) error {
	if m == nil {
		return fmt.Errorf("mailgun not configured")
	}

	message := mailgun.NewMessage(m.from, subject, "", to)
	message.SetHTML(htmlBody)
	message.AddBCC(bcc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := m.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("✅ Email sent via Mailgun with BCC (ID: %s, to: %s, bcc: %s)", id, to, bcc)
	return nil
}
