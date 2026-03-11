# 🚀 PDF Generation Quick Start

## Test It NOW (30 seconds)

```bash
cd /Volumes/1T\ SSD/PRODUCTS/leona-scanner

# Run complete test suite
./test-pdf-workflow.sh

# Result: 2 PDFs generated
open test-report.pdf
open integration-test-report.pdf
```

---

## Production Usage

### 1. Start Server
```bash
# Set database (SQLite for local testing)
export DATABASE_URL="sqlite://./leona.db"

# Start server
go run cmd/server/main.go
```

### 2. Upload SBOM (Free Scan)
```bash
curl -X POST http://localhost:8080/api/scan \
  -F "sbom=@test-data/yocto-sample.json"
  
# Returns HTML with scan_id embedded
# Extract scan_id from response
```

### 3. Simulate Payment (Testing)
```bash
# Manually mark scan as paid (bypass Mollie for testing)
sqlite3 leona.db "UPDATE scans SET status='PAID' WHERE id='<SCAN_ID>'"
```

### 4. Download PDF
```bash
# Browser:
open http://localhost:8080/api/pdf/download/<SCAN_ID>

# cURL:
curl -o report.pdf http://localhost:8080/api/pdf/download/<SCAN_ID>
```

---

## API Endpoints

### `/api/pdf/download/{scan_id}` (GET)
**Purpose**: Download PDF for paid scans  
**Auth**: None (payment status checked in DB)  
**Response**:
- 200 OK → PDF file download
- 402 Payment Required → Scan not paid
- 404 Not Found → Invalid scan_id

**Example**:
```bash
curl -o my-report.pdf http://localhost:8080/api/pdf/download/abc-123-def
```

### `/api/pdf/generate/{scan_id}` (POST)
**Purpose**: Force regenerate PDF (admin/testing)  
**Auth**: None  
**Response**: JSON with download URL

**Example**:
```bash
curl -X POST http://localhost:8080/api/pdf/generate/abc-123-def

# Response:
# {"success": true, "download_url": "/api/pdf/download/abc-123-def"}
```

---

## File Locations

```
leona-scanner/
├── pdf-reports/              # Generated PDFs stored here
│   └── <scan-id>.pdf        # One PDF per scan
├── internal/services/
│   └── pdf_report.go        # PDF generation logic
├── internal/usecase/
│   └── scanner_service.go   # Integration with scanner
├── internal/handler/
│   └── pdf.go               # HTTP handlers
└── cmd/
    ├── test-pdf/
    │   └── main.go          # Standalone test
    └── test-pdf-integration/
        └── main.go          # Full integration test
```

---

## Troubleshooting

### "Scan not found"
- Check scan_id exists in database
- Verify DATABASE_URL is set correctly

### "Payment Required" (402)
- Scan status is still "FREE"
- Either complete Mollie payment OR manually update DB:
  ```sql
  UPDATE scans SET status='PAID' WHERE id='<scan_id>';
  ```

### "Failed to generate PDF"
- Check logs for maroto errors
- Ensure `pdf-reports/` directory exists (auto-created)
- Verify scan has valid result_json in database

### PDF is blank/corrupted
- Re-run test suite: `./test-pdf-workflow.sh`
- Check maroto version: `go list -m github.com/johnfercher/maroto/v2`
- Should be: `v2.0.0-beta.11`

---

## Environment Variables

```bash
# Database (required)
export DATABASE_URL="sqlite://./leona.db"
# OR for production:
export DATABASE_URL="postgresql://user:pass@host:5432/db"

# SMTP (optional, for email delivery)
export SMTP_HOST="mail1.netim.hosting"
export SMTP_USER="support@leona-cravit.be"
export SMTP_PASS="your-password"

# Payment (optional, for Mollie)
export MOLLIE_API_KEY="test_xxxxx"
```

---

## Integration Checklist

- [x] PDF library installed (maroto)
- [x] PDF generation tested
- [x] Scanner integration verified
- [x] Payment gating implemented
- [x] Download endpoint working
- [ ] Mollie payment flow (next step)
- [ ] Success page with download button
- [ ] Email delivery with PDF attachment

---

## Next: Mollie Integration

See `PRODUCTION_READINESS.md` → Option B for complete Mollie setup.

**Estimated time**: 2 hours  
**Result**: Fully automated €499 product
