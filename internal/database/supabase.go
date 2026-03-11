package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type SupabaseClient struct {
	URL        string
	ServiceKey string
	HTTPClient *http.Client
}

// Scan represents a scan record in the database
type Scan struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	Status           string     `json:"status"` // pending, analyzing, completed, failed
	PaymentStatus    string     `json:"payment_status"` // free, pending_payment, paid
	ComplianceScore  *int       `json:"compliance_score,omitempty"`
	SBOMFormat       *string    `json:"sbom_format,omitempty"`
	SBOMSizeKB       *int       `json:"sbom_size_kb,omitempty"`
	TotalComponents  *int       `json:"total_components,omitempty"`
	KernelVersion    *string    `json:"kernel_version,omitempty"`
	CriticalFindings int        `json:"critical_findings"`
	Hostname         *string    `json:"hostname,omitempty"`
	CompanyName      *string    `json:"company_name,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	AnalyzedAt       *time.Time `json:"analyzed_at,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
}

// Payment represents a payment record
type Payment struct {
	ID               uuid.UUID       `json:"id"`
	ScanID           uuid.UUID       `json:"scan_id"`
	MolliePaymentID  string          `json:"mollie_payment_id"`
	Amount           float64         `json:"amount"`
	Currency         string          `json:"currency"`
	Status           string          `json:"status"` // open, paid, failed, canceled, expired
	PaymentMethod    *string         `json:"payment_method,omitempty"`
	Tier             string          `json:"tier"` // tier1, tier2, tier3
	CustomerEmail    string          `json:"customer_email"`
	CustomerName     *string         `json:"customer_name,omitempty"`
	Description      *string         `json:"description,omitempty"`
	RedirectURL      *string         `json:"redirect_url,omitempty"`
	WebhookURL       *string         `json:"webhook_url,omitempty"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	PaidAt           *time.Time      `json:"paid_at,omitempty"`
	CanceledAt       *time.Time      `json:"canceled_at,omitempty"`
	FailedAt         *time.Time      `json:"failed_at,omitempty"`
}

// Lead represents a marketing lead
type Lead struct {
	ID                    uuid.UUID  `json:"id"`
	Email                 string     `json:"email"`
	LeadType              string     `json:"lead_type"` // engineer, lawyer, general
	FirstName             *string    `json:"first_name,omitempty"`
	LastName              *string    `json:"last_name,omitempty"`
	CompanyName           *string    `json:"company_name,omitempty"`
	JobTitle              *string    `json:"job_title,omitempty"`
	Phone                 *string    `json:"phone,omitempty"`
	Source                string     `json:"source"` // website, linkedin, referral
	Status                string     `json:"status"` // new, contacted, qualified, converted
	Notes                 *string    `json:"notes,omitempty"`
	LeadMagnetRequested   *string    `json:"lead_magnet_requested,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	LastContactedAt       *time.Time `json:"last_contacted_at,omitempty"`
	ConvertedAt           *time.Time `json:"converted_at,omitempty"`
}

// AnalysisResult represents a compliance finding
type AnalysisResult struct {
	ID                    uuid.UUID `json:"id"`
	ScanID                uuid.UUID `json:"scan_id"`
	Framework             string    `json:"framework"` // CRA, CER, NIS2
	RequirementID         string    `json:"requirement_id"`
	RequirementDescription *string  `json:"requirement_description,omitempty"`
	Status                string    `json:"status"` // compliant, partial, non_compliant
	Finding               *string   `json:"finding,omitempty"`
	Remediation           *string   `json:"remediation,omitempty"`
	Severity              *string   `json:"severity,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
}

func NewSupabaseClient(url, serviceKey string) *SupabaseClient {
	return &SupabaseClient{
		URL:        url,
		ServiceKey: serviceKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateScan inserts a new scan record
func (c *SupabaseClient) CreateScan(ctx context.Context, email string) (*Scan, error) {
	scan := &Scan{
		Email:            email,
		Status:           "pending",
		PaymentStatus:    "free",
		CriticalFindings: 0,
	}

	body, _ := json.Marshal(scan)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/rest/v1/scans", bytes.NewBuffer(body))
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create scan: %s", string(bodyBytes))
	}

	var scans []Scan
	if err := json.NewDecoder(resp.Body).Decode(&scans); err != nil {
		return nil, err
	}

	if len(scans) == 0 {
		return nil, fmt.Errorf("no scan returned")
	}

	return &scans[0], nil
}

// UpdateScanStatus updates the status and analysis results
func (c *SupabaseClient) UpdateScanStatus(ctx context.Context, scanID uuid.UUID, status string, complianceScore *int, findings int) error {
	now := time.Now()
	update := map[string]interface{}{
		"status":            status,
		"analyzed_at":       now,
		"critical_findings": findings,
	}
	if complianceScore != nil {
		update["compliance_score"] = *complianceScore
	}

	body, _ := json.Marshal(update)
	req, _ := http.NewRequestWithContext(ctx, "PATCH", 
		fmt.Sprintf("%s/rest/v1/scans?id=eq.%s", c.URL, scanID), 
		bytes.NewBuffer(body))
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update scan: %s", string(bodyBytes))
	}

	return nil
}

// CreatePayment inserts a new payment record
func (c *SupabaseClient) CreatePayment(ctx context.Context, payment *Payment) error {
	body, _ := json.Marshal(payment)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/rest/v1/payments", bytes.NewBuffer(body))
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create payment: %s", string(bodyBytes))
	}

	return nil
}

// UpdatePaymentStatus updates payment status (for webhook)
func (c *SupabaseClient) UpdatePaymentStatus(ctx context.Context, mollieID string, status string, paidAt *time.Time) error {
	update := map[string]interface{}{
		"status": status,
	}
	if paidAt != nil {
		update["paid_at"] = paidAt
	}

	body, _ := json.Marshal(update)
	req, _ := http.NewRequestWithContext(ctx, "PATCH", 
		fmt.Sprintf("%s/rest/v1/payments?mollie_payment_id=eq.%s", c.URL, mollieID), 
		bytes.NewBuffer(body))
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update payment: %s", string(bodyBytes))
	}

	return nil
}

// UnlockScan marks a scan as paid
func (c *SupabaseClient) UnlockScan(ctx context.Context, scanID uuid.UUID) error {
	now := time.Now()
	update := map[string]interface{}{
		"payment_status": "paid",
		"paid_at":        now,
	}

	body, _ := json.Marshal(update)
	req, _ := http.NewRequestWithContext(ctx, "PATCH", 
		fmt.Sprintf("%s/rest/v1/scans?id=eq.%s", c.URL, scanID), 
		bytes.NewBuffer(body))
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to unlock scan: %s", string(bodyBytes))
	}

	return nil
}

// CreateLead inserts a new lead
func (c *SupabaseClient) CreateLead(ctx context.Context, lead *Lead) error {
	body, _ := json.Marshal(lead)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/rest/v1/leads", bytes.NewBuffer(body))
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Conflict is OK (duplicate email)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create lead: %s", string(bodyBytes))
	}

	return nil
}

// SaveAnalysisResults batch inserts compliance findings
func (c *SupabaseClient) SaveAnalysisResults(ctx context.Context, results []AnalysisResult) error {
	body, _ := json.Marshal(results)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/rest/v1/analysis_results", bytes.NewBuffer(body))
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to save analysis results: %s", string(bodyBytes))
	}

	return nil
}
