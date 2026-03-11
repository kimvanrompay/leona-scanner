# Email Setup Guide for LEONA Lead Magnets

## Netim SMTP Configuration

Since your email is `support@leona-cravit.be` on Netim, you need to configure SMTP settings.

### Step 1: Get SMTP Credentials from Netim

1. Log in to [Netim control panel](https://www.netim.com)
2. Go to **Email** → **Email Accounts**
3. Find `support@leona-cravit.be`
4. Click **Configure** or **Settings**
5. Look for **SMTP Server** settings

### Step 2: Update `.env` File

Add these settings to your `.env` file:

```bash
# SMTP Configuration (Netim)
SMTP_HOST=mail1.netim.hosting
SMTP_PORT=465
SMTP_USER=support@leona-cravit.be
SMTP_PASS=your_email_password_here
SMTP_FROM=support@leona-cravit.be
```

**Note:** Netim uses SSL/TLS on port 465 (not STARTTLS on 587).

### Step 3: Test Email Delivery

Start the server and test the lead magnet:

```bash
cd /Volumes/1T\ SSD/PRODUCTS/leona-scanner
go run cmd/server/main.go
```

Then open `http://localhost:8080` and scroll to "Voor Engineers en Juristen" section. Enter your email and click "Ontvang Layer".

Check your inbox for the email with subject: **"Jouw meta-leona CRA Validator Layer"**

### Step 4: Troubleshooting

#### Error: "SMTP not configured"
- Check that all SMTP_ variables are set in `.env`
- Restart the server after updating `.env`

#### Error: "535 Authentication failed"
- Wrong username or password
- Check credentials in Netim panel
- Some providers require "app passwords" instead of main password

#### Error: "Connection refused"
- Wrong SMTP host or port
- Try port 465 with SSL instead of 587 with STARTTLS
- Check firewall settings

#### Emails go to spam
- Add SPF record to DNS: `v=spf1 include:netim.com ~all`
- Add DKIM signature (ask Netim support)
- Use consistent "From" address (support@leona-cravit.be)

### Common Netim SMTP Settings

**Official Netim Settings (from your panel):**
```bash
SMTP_HOST=mail1.netim.hosting
SMTP_PORT=465
```

This uses **SSL/TLS** (not STARTTLS). The Go code is already configured for this.

### What Gets Sent?

**Engineer Email:**
- Subject: "Jouw meta-leona CRA Validator Layer"
- Contains: Installation instructions for meta-leona Yocto layer
- CTA: Download link + link to V-Assessor scanner

**Lawyer Email:**
- Subject: "CRA Annex I Mapping Template voor uw cliënten"
- Contains: Excel template for compliance mapping
- CTA: Download link + €2,500/year partnership offer

### Lead Tracking

All email submissions are saved to Supabase `leads` table (if configured):

```sql
SELECT * FROM leads WHERE lead_magnet_requested IS NOT NULL;
```

Fields tracked:
- `email`
- `lead_type` ("engineer" or "lawyer")
- `source` ("website")
- `status` ("new")
- `lead_magnet_requested` ("meta-leona" or "annex-i-template")
- `created_at`

### Production Checklist

Before going live:

- [ ] Test both engineer and lawyer email flows
- [ ] Verify emails don't go to spam (test with Gmail, Outlook, Protonmail)
- [ ] Check "From" address shows as "support@leona-cravit.be"
- [ ] Verify all links in email work (https://leona-cravit.be/...)
- [ ] Set up email monitoring (track delivery rate)
- [ ] Add SPF/DKIM DNS records for better deliverability
- [ ] Create actual downloadable files:
  - `/static/downloads/meta-leona.tar.gz`
  - `/static/downloads/CRA_Annex_I_Mapping_Template.xlsx`

### Next Steps

1. Configure SMTP settings in `.env`
2. Test email delivery locally
3. Create the actual downloadable files (meta-leona layer, Excel template)
4. Deploy to production
5. Monitor lead conversions in Supabase

Need help? Email support@leona-cravit.be (that's you! 😄)
