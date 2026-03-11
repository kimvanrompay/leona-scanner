# LEONA - CRA Compliance Scanner

**LEONA** is a professional CRA (Cyber Resilience Act) compliance scanner for embedded systems. It analyzes SBOM (Software Bill of Materials) files and generates detailed compliance reports for Yocto, Zephyr, FreeRTOS, and generic Linux systems.

## 🎯 Business Model

- **Bait & Sell**: Free compliance score display, €99 for full PDF report
- **Target Market**: Embedded systems companies facing September 2026 CRA deadline
- **Revenue Goal**: €10,000+ from compliance reports
- **Payment**: Stripe integration with iDEAL and card payments
- **Delivery**: Automated PDF generation and email delivery via SMTP

## 🎨 Branding

- **Primary Color**: Royal Blue (#1428A0) with gradients to Davis Blue (#0033A0)
- **Accent Color**: Davis Orange (#FF6B35)
- **Theme**: Professional law firm aesthetic, compliance-focused, urgency-driven
- **Style**: Inspired by Davis Polk - royal blue backgrounds with white text and orange accents

## 🏗️ Architecture

```
leona-scanner/
├── cmd/server/          # Application entrypoint
├── internal/
│   ├── scanner/         # Core CRA analysis engine
│   │   ├── analyzer.go  # Platform-specific compliance rules
│   │   └── parser.go    # SBOM parser (CycloneDX, SPDX)
│   ├── handler/         # HTTP handlers
│   ├── usecase/         # Business logic services
│   │   ├── scanner_service.go
│   │   └── pdf_service.go
│   └── repository/      # PostgreSQL data layer
├── templates/           # HTML templates
│   ├── index.html       # Landing page
│   └── results.html     # Analysis results page
├── migrations/          # Database migrations
├── Dockerfile           # Multi-stage Docker build
└── docker-compose.yml   # Full stack orchestration
```

## 🚀 Quick Start

### Prerequisites

- Go 1.22+ (for local development)
- Docker & Docker Compose (optional, for PostgreSQL)
- Stripe account (for payments)
- SMTP credentials (Gmail, SendGrid, etc.)
- Supabase account (for production PostgreSQL)

### 1. Clone and Configure

```bash
cd /Volumes/1T\ SSD/PRODUCTS/leona-scanner
cp .env.example .env
```

Edit `.env` with your credentials:

```bash
# Database (SQLite for dev, PostgreSQL for production)
DATABASE_URL=./leona.db

# Stripe
STRIPE_API_KEY=sk_test_your_key_here
STRIPE_WEBHOOK_SECRET=whsec_your_secret_here

# SMTP (Gmail example)
SMTP_USER=support@leonatech.com
SMTP_PASS=your-app-password
SMTP_FROM=support@leonatech.com
```

### 2. Start with Docker Compose

```bash
docker-compose up --build
```

The application will be available at `http://localhost:8080`

### 3. Run Locally (Development with SQLite)

```bash
# Install dependencies
go mod download

# Run application (uses SQLite by default)
go run cmd/server/main.go
```

The SQLite database (`leona.db`) will be created automatically in the project root.

### 4. Production with Supabase

1. Create a Supabase project at https://supabase.com
2. Get your PostgreSQL connection string from Settings → Database
3. Update `.env` with your Supabase connection:

```bash
DATABASE_URL=postgresql://postgres:[YOUR-PASSWORD]@db.[YOUR-PROJECT-REF].supabase.co:5432/postgres
```

4. Run migrations in Supabase SQL Editor:

```sql
CREATE TABLE leads (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE scans (
    id UUID PRIMARY KEY,
    lead_id INTEGER NOT NULL REFERENCES leads(id),
    platform VARCHAR(50) NOT NULL,
    score INTEGER NOT NULL,
    status VARCHAR(10) NOT NULL,
    raw_data BYTEA,
    result_json TEXT,
    created_at TIMESTAMP NOT NULL,
    paid_at TIMESTAMP
);
```

## 📊 Features

### Universal CRA Rules (All Platforms)

- **CPE Traceability** (Article 11): Checks for CPE/PURL identifiers
- **Version Control**: Validates version information presence
- **License Compliance**: Detects missing licenses and GPL-3.0 copyleft risks
- **IP Risk Detection**: Flags GPL-3.0/AGPL licenses

### Platform-Specific Analysis

#### Yocto
- LTS kernel detection (5.15, 6.1, 6.6)
- Meta-security layer validation
- Security patch management assessment

#### Zephyr RTOS
- Kernel version validation
- LTS version recommendations
- Real-time compliance checks

#### FreeRTOS
- LTS version detection (202012.00+)
- MIT license validation
- IoT security best practices

### Scoring System

- **100 points base score**
- Deductions:
  - Missing CPE/PURL: -10 points
  - Missing version: -5 points
  - Missing license: -3 points
  - GPL-3.0 detected: -8 points
  - Platform-specific penalties

- **Status Determination**:
  - ≥90: CONFORM
  - 75-89: VOORWAARDELIJK CONFORM
  - <75: NIET-CONFORM

## 💳 Payment Flow

1. User uploads SBOM file
2. Free analysis shows score and top 5 issues (blurred)
3. Click "Koop Volledig Rapport" → Stripe Checkout (€99)
4. Payment success → Webhook triggers PDF generation
5. PDF emailed automatically to customer

## 🗄️ Database Schema

### `leads` table
```sql
id           SERIAL PRIMARY KEY
email        VARCHAR(255) UNIQUE
created_at   TIMESTAMP
```

### `scans` table
```sql
id           UUID PRIMARY KEY
lead_id      INTEGER REFERENCES leads(id)
platform     VARCHAR(50)
score        INTEGER
status       VARCHAR(10) -- FREE or PAID
raw_data     BYTEA
result_json  TEXT
created_at   TIMESTAMP
paid_at      TIMESTAMP NULL
```

## 📄 Supported SBOM Formats

- **CycloneDX**: JSON and XML
- **SPDX**: JSON

Auto-detection based on file content.

## 🔧 Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@localhost:5432/leona` |
| `PORT` | HTTP server port | `8080` |
| `BASE_URL` | Application base URL | `http://localhost:8080` |
| `STRIPE_API_KEY` | Stripe secret key | `sk_test_...` |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret | `whsec_...` |
| `SMTP_HOST` | SMTP server hostname | `smtp.gmail.com` |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USER` | SMTP username | `your-email@gmail.com` |
| `SMTP_PASS` | SMTP password | `app-password` |
| `SMTP_FROM` | Sender email address | `noreply@leona.io` |

## 🧪 Testing Locally

### Upload Test SBOM

Create `test-sbom.json`:

```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.4",
  "components": [
    {
      "name": "linux-yocto",
      "version": "6.1.0",
      "cpe": "cpe:2.3:o:linux:linux_kernel:6.1.0",
      "licenses": [{"license": {"id": "GPL-2.0"}}]
    }
  ]
}
```

Upload via web interface at `http://localhost:8080`

## 🚢 Production Deployment

### Stripe Webhook Setup

1. Go to Stripe Dashboard → Webhooks
2. Add endpoint: `https://yourdomain.com/api/webhook`
3. Select event: `checkout.session.completed`
4. Copy webhook secret to `STRIPE_WEBHOOK_SECRET`

### SMTP Setup (Gmail)

1. Enable 2FA on Google Account
2. Generate App Password: https://myaccount.google.com/apppasswords
3. Use App Password as `SMTP_PASS`

### Dockerfile Production Build

```bash
docker build -t leona-scanner:latest .
docker run -p 8080:8080 --env-file .env leona-scanner:latest
```

## 📈 Revenue Tracking

Monitor via PostgreSQL:

```sql
-- Total revenue
SELECT COUNT(*) * 99 as total_revenue_eur 
FROM scans 
WHERE status = 'PAID';

-- Daily conversions
SELECT DATE(paid_at), COUNT(*) as paid_scans
FROM scans
WHERE status = 'PAID'
GROUP BY DATE(paid_at);

-- Conversion rate
SELECT 
  COUNT(CASE WHEN status = 'PAID' THEN 1 END)::float / COUNT(*) * 100 as conversion_rate
FROM scans;
```

## 🛡️ Security Considerations

- Stripe webhook signature verification implemented
- SQL injection protection via parameterized queries
- Environment variables for sensitive data
- Non-root Docker container execution
- 10MB file upload limit

## 📝 License

Proprietary - All rights reserved

## 🤝 Support

For issues or questions, contact: support@leona.io

---

**CRA Deadline: September 2026** - Help embedded systems companies achieve compliance today!
