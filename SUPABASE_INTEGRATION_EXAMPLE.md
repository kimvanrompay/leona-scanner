# Supabase Integration Guide

## Overview
This guide shows how to integrate the Supabase client into your existing LEONA scanner endpoints.

## Environment Variables

Add to your `.env`:
```bash
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
MOLLIE_API_KEY=test_xxxxxxxxxxxxx
```

## Code Integration

### 1. Initialize Supabase Client in main.go

```go
package main

import (
    "github.com/leona-scanner/internal/database"
    "os"
)

var db *database.SupabaseClient

func init() {
    supabaseURL := os.Getenv("SUPABASE_URL")
    supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")
    
    if supabaseURL != "" && supabaseKey != "" {
        db = database.NewSupabaseClient(supabaseURL, supabaseKey)
    }
}
```

### 2. Track Scan on Upload

When user uploads SBOM file:

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")
    
    // Create scan record in database
    ctx := r.Context()
    scan, err := db.CreateScan(ctx, email)
    if err != nil {
        log.Printf("Failed to create scan: %v", err)
        // Continue anyway - don't block user
    }
    
    // Also create lead
    lead := &database.Lead{
        Email:    email,
        LeadType: "engineer",
        Source:   "website",
        Status:   "new",
    }
    db.CreateLead(ctx, lead)
    
    // Store scan ID in session/context for later use
    sessionID := scan.ID.String()
    
    // Continue with your existing SBOM processing...
}
```

### 3. Update Scan After Analysis

After your analysis completes:

```go
func analyzeHandler(scanID uuid.UUID) {
    // Your existing analysis logic...
    complianceScore := calculateComplianceScore()
    criticalFindings := countCriticalFindings()
    
    // Update database
    ctx := context.Background()
    err := db.UpdateScanStatus(ctx, scanID, "completed", &complianceScore, criticalFindings)
    if err != nil {
        log.Printf("Failed to update scan: %v", err)
    }
    
    // Save detailed findings
    results := []database.AnalysisResult{
        {
            ScanID:        scanID,
            Framework:     "CRA",
            RequirementID: "ANNEX_I_PART_I_1",
            Status:        "non_compliant",
            Finding:       strPtr("No SBOM metadata found"),
            Remediation:   strPtr("Add cyclonedx-bom.bbclass to your image recipe"),
            Severity:      strPtr("high"),
        },
        // Add more results...
    }
    db.SaveAnalysisResults(ctx, results)
}

func strPtr(s string) *string {
    return &s
}
```

### 4. Payment Integration

When user clicks "Unlock Full Report":

```go
func createPaymentHandler(w http.ResponseWriter, r *http.Request) {
    scanID := r.FormValue("scan_id")
    email := r.FormValue("email")
    tier := r.FormValue("tier") // "tier1", "tier2", "tier3"
    
    // Determine amount based on tier
    amounts := map[string]float64{
        "tier1": 499.00,
        "tier2": 2499.00,
        "tier3": 4900.00,
    }
    amount := amounts[tier]
    
    // Create Mollie payment
    molliePayment, err := mollieClient.Payments.Create(context.Background(), &mollie.PaymentRequest{
        Amount: &mollie.Amount{
            Currency: "EUR",
            Value:    fmt.Sprintf("%.2f", amount),
        },
        Description: "V-Assessor Technical Construction File",
        RedirectURL: "https://leona-cravit.be/success",
        WebhookURL:  "https://leona-cravit.be/webhook/mollie",
        Metadata: map[string]interface{}{
            "scan_id": scanID,
        },
    })
    
    // Store payment in database
    scanUUID, _ := uuid.Parse(scanID)
    payment := &database.Payment{
        ScanID:          scanUUID,
        MolliePaymentID: molliePayment.ID,
        Amount:          amount,
        Currency:        "EUR",
        Status:          string(molliePayment.Status),
        Tier:            tier,
        CustomerEmail:   email,
        RedirectURL:     strPtr("https://leona-cravit.be/success"),
        WebhookURL:      strPtr("https://leona-cravit.be/webhook/mollie"),
    }
    
    if err := db.CreatePayment(context.Background(), payment); err != nil {
        log.Printf("Failed to store payment: %v", err)
    }
    
    // Redirect to Mollie checkout
    http.Redirect(w, r, molliePayment.Links.Checkout.Href, http.StatusSeeOther)
}
```

### 5. Success Page - Check Payment Status

On `/success` page:

```go
func successHandler(w http.ResponseWriter, r *http.Request) {
    scanID := r.URL.Query().Get("scan_id")
    
    // Query database to check payment status
    // (You'll need to add a GetScan method to SupabaseClient)
    
    // If paid, show download links
    // If pending, show "Processing payment..."
    
    tmpl.ExecuteTemplate(w, "success.html", data)
}
```

## Webhook Server Deployment

Run the webhook server separately:

```bash
cd cmd/webhook
go build -o webhook
./webhook
```

Or with systemd:

```ini
[Unit]
Description=LEONA Mollie Webhook Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/leona-scanner
Environment="SUPABASE_URL=https://your-project.supabase.co"
Environment="SUPABASE_SERVICE_KEY=your-service-key"
Environment="MOLLIE_API_KEY=your-mollie-key"
Environment="PORT=8081"
ExecStart=/opt/leona-scanner/webhook
Restart=always

[Install]
WantedBy=multi-user.target
```

## Analytics Queries

### Check conversion rate
```sql
SELECT * FROM scan_analytics WHERE date >= CURRENT_DATE - INTERVAL '7 days';
```

### Revenue tracking
```sql
SELECT * FROM revenue_analytics WHERE date >= CURRENT_DATE - INTERVAL '30 days';
```

### Hot leads (multiple scans, no payment)
```sql
SELECT 
    email, 
    COUNT(*) as scan_count,
    MAX(created_at) as last_scan
FROM scans 
WHERE payment_status = 'free'
GROUP BY email 
HAVING COUNT(*) > 1
ORDER BY scan_count DESC;
```

## Next Steps

1. **Deploy Supabase schema**: Run `migrations/001_create_tables.sql` in your Supabase SQL editor
2. **Add env variables**: Update your production `.env`
3. **Integrate tracking**: Add CreateScan/UpdateScan calls to your existing handlers
4. **Deploy webhook server**: Set up separate service for Mollie webhooks
5. **Test payment flow**: Use Mollie test mode to verify end-to-end
6. **Monitor analytics**: Check Supabase dashboard for conversion metrics
