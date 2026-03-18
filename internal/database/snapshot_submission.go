//nolint:errcheck // Supabase REST API calls follow existing pattern

// Package database provides Supabase database operations
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

// SnapshotSubmission represents a detailed snapshot audit request in the database
type SnapshotSubmission struct {
	ID        uuid.UUID `json:"id"`
	OrderUUID string    `json:"order_uuid"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Company   string    `json:"company"`
	Phone     *string   `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Build system
	BuildSystem        string  `json:"build_system"`
	BuildSystemVersion *string `json:"build_system_version,omitempty"`
	TargetArchitecture string  `json:"target_architecture"`
	KernelVersion      string  `json:"kernel_version"`
	Libc               string  `json:"libc"`

	// Product
	ProductName     string   `json:"product_name"`
	ProductCategory string   `json:"product_category"`
	Connectivity    []string `json:"connectivity,omitempty"`
	AnnualVolume    string   `json:"annual_volume"`

	// Security
	SecureBoot      *string  `json:"secure_boot,omitempty"`
	TPM             *string  `json:"tpm,omitempty"`
	OTAFeatures     []string `json:"ota_features,omitempty"`
	UpdateFramework *string  `json:"update_framework,omitempty"`

	// Artifacts
	ArtifactAccess     string   `json:"artifact_access"`
	EstimatedSize      *string  `json:"estimated_size,omitempty"`
	AvailableArtifacts []string `json:"available_artifacts,omitempty"`

	// Context
	Timeline        *string `json:"timeline,omitempty"`
	Concerns        *string `json:"concerns,omitempty"`
	AdditionalNotes *string `json:"additional_notes,omitempty"`

	// Legal
	NDAAccepted bool `json:"nda_accepted"`

	// Payment
	PaymentStatus      string     `json:"payment_status"`
	MolliePaymentID    *string    `json:"mollie_payment_id,omitempty"`
	PaymentCompletedAt *time.Time `json:"payment_completed_at,omitempty"`

	// Status
	Status string `json:"status"`
}

// CreateSnapshotSubmission creates a new snapshot submission in the database
func (c *SupabaseClient) CreateSnapshotSubmission(ctx context.Context, submission *SnapshotSubmission) error {
	body, _ := json.Marshal(submission)               //nolint:errcheck // REST API pattern
	req, _ := http.NewRequestWithContext(ctx, "POST", //nolint:errcheck // REST API pattern
		c.URL+"/rest/v1/snapshot_submissions",
		bytes.NewBuffer(body))
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort close

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body) //nolint:errcheck // Debug logging
		return fmt.Errorf("failed to create snapshot submission: %s", string(bodyBytes))
	}

	var submissions []SnapshotSubmission
	if err := json.NewDecoder(resp.Body).Decode(&submissions); err != nil {
		return err
	}

	if len(submissions) > 0 {
		*submission = submissions[0]
	}

	return nil
}

// GetSnapshotSubmissionByOrderUUID retrieves a snapshot submission by order UUID
//
//nolint:lll // Function signature
func (c *SupabaseClient) GetSnapshotSubmissionByOrderUUID(ctx context.Context, orderUUID string) (*SnapshotSubmission, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", //nolint:errcheck // REST API pattern
		fmt.Sprintf("%s/rest/v1/snapshot_submissions?order_uuid=eq.%s", c.URL, orderUUID), nil)
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort close

	var submissions []SnapshotSubmission
	if err := json.NewDecoder(resp.Body).Decode(&submissions); err != nil {
		return nil, err
	}

	if len(submissions) == 0 {
		return nil, fmt.Errorf("snapshot submission not found")
	}

	return &submissions[0], nil
}

// UpdateSnapshotPaymentStatus updates the payment status of a snapshot submission
func (c *SupabaseClient) UpdateSnapshotPaymentStatus(
	ctx context.Context,
	orderUUID string,
	paymentStatus string,
	molliePaymentID string,
	paidAt *time.Time,
) error {
	update := map[string]interface{}{
		"payment_status":       paymentStatus,
		"mollie_payment_id":    molliePaymentID,
		"payment_completed_at": paidAt,
	}
	if paymentStatus == "paid" {
		update["status"] = "paid"
	} else if paymentStatus == "failed" {
		update["status"] = "new"
	}

	body, _ := json.Marshal(update)                    //nolint:errcheck // REST API pattern
	req, _ := http.NewRequestWithContext(ctx, "PATCH", //nolint:errcheck // REST API pattern
		fmt.Sprintf("%s/rest/v1/snapshot_submissions?order_uuid=eq.%s", c.URL, orderUUID),
		bytes.NewBuffer(body))
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort close

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body) //nolint:errcheck // Debug logging
		return fmt.Errorf("failed to update payment status: %s", string(bodyBytes))
	}

	return nil
}

// ListSnapshotSubmissions retrieves all snapshot submissions with optional filtering
func (c *SupabaseClient) ListSnapshotSubmissions(
	ctx context.Context,
	limit int,
	offset int,
	status *string,
) ([]*SnapshotSubmission, error) {
	url := fmt.Sprintf("%s/rest/v1/snapshot_submissions?order=created_at.desc&limit=%d&offset=%d",
		c.URL, limit, offset)
	if status != nil && *status != "" {
		url += fmt.Sprintf("&status=eq.%s", *status)
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil) //nolint:errcheck // REST API pattern
	req.Header.Set("apikey", c.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+c.ServiceKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort close

	var submissions []*SnapshotSubmission
	if err := json.NewDecoder(resp.Body).Decode(&submissions); err != nil {
		return nil, err
	}

	return submissions, nil
}
