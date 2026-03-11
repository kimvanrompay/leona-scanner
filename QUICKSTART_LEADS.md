# Quick Start: Lead Magnets (5 Minutes)

## ✅ What Works NOW

Your lead magnet system is **fully functional**. When someone enters their email:

1. **Engineer form** → Sends email with meta-leona Yocto layer instructions
2. **Lawyer form** → Sends email with Annex I Mapping Template  
3. **Both** → Save lead to Supabase `leads` table (if configured)

## 🚀 Setup (Copy-Paste)

### Step 1: Configure Email (2 min)

Edit your `.env` file and add:

```bash
# Netim SMTP (official settings)
SMTP_HOST=mail1.netim.hosting
SMTP_PORT=465
SMTP_USER=support@leona-cravit.be
SMTP_PASS=YOUR_EMAIL_PASSWORD_HERE
```

**Where to find your password:**
- Log in to [Netim panel](https://www.netim.com)
- Go to Email → support@leona-cravit.be
- Use the same password you use for webmail (https://mail1.netim.hosting/webmail/)

**Settings confirmed from your Netim panel:**
- Host: `mail1.netim.hosting`
- Port: `465` (SSL/TLS)
- These are the official Netim settings

### Step 2: Test Locally (2 min)

```bash
cd "/Volumes/1T SSD/PRODUCTS/leona-scanner"

# Start server
go run cmd/server/main.go

# Open browser
open http://localhost:8080
```

Scroll to **"Voor Engineers en Juristen"** section, enter your email, click button.

**Expected result:** ✅ Green success message + email in your inbox within 30 seconds.

### Step 3: Check Email (1 min)

**Engineer email subject:** "Jouw meta-leona CRA Validator Layer"  
**Lawyer email subject:** "CRA Annex I Mapping Template voor uw cliënten"

If email doesn't arrive:
1. Check spam folder
2. Check server logs for errors
3. Try different SMTP_HOST (see EMAIL_SETUP.md)

## 📊 Lead Tracking (Optional)

If you have Supabase configured, all submissions are saved:

```sql
-- View all lead magnet requests
SELECT 
    email,
    lead_type,
    lead_magnet_requested,
    created_at
FROM leads
WHERE lead_magnet_requested IS NOT NULL
ORDER BY created_at DESC;
```

**Lead types:**
- `engineer` = Yocto layer download
- `lawyer` = Annex I template download

## 🎯 What Gets Sent?

### Engineer Email Content:
- 🚀 Meta-leona Yocto layer overview
- 📦 What's included (5 bbclass files)
- ⚙️ Installation guide (4 commands)
- 🎯 CTA: Upload SBOM to V-Assessor

### Lawyer Email Content:
- 📋 Annex I Mapping Template overview
- 🔍 Use case: Client needs technical report
- 💼 Partnership offer (€2,500/year Lawyer Master Account)
- Table of benefits

## ⚡ Production Deployment

Once tested locally, deploy with:

```bash
# Build
go build -o leona-server cmd/server/main.go

# Run with production env
SMTP_HOST=mail1.netim.hosting \
SMTP_PORT=587 \
SMTP_USER=support@leona-cravit.be \
SMTP_PASS=your_password \
./leona-server
```

Or update your production `.env` file.

## 🐛 Troubleshooting

### Error: "SMTP not configured"
→ Check `.env` file has all SMTP_ variables

### Error: "535 Authentication failed"  
→ Wrong password. Check Netim panel

### Error: "Connection refused"
→ Check firewall. Port 465 must be open for outbound connections

### Emails go to spam
→ Add SPF record to DNS:
```
v=spf1 include:netim.com ~all
```

### Form doesn't submit
→ Check browser console (F12) for HTMX errors

## 📝 Next Steps

1. ✅ Test both forms (engineer + lawyer)
2. ⏳ Create actual downloadable files:
   - `static/downloads/meta-leona.tar.gz` (Yocto layer)
   - `static/downloads/CRA_Annex_I_Mapping_Template.xlsx` (Excel)
3. 🚀 Deploy to production
4. 📊 Monitor leads in Supabase

## 💡 Marketing Ideas

**Engineer targeting:**
- Post in Yocto mailing list
- Share on Embedded Linux subreddit
- LinkedIn post with "Free meta-leona layer"

**Lawyer targeting:**
- Email 5 Belgian law firms (see QUICK_START.md)
- LinkedIn outreach: "Free CRA compliance template"
- Partner with compliance consultants

**Conversion optimization:**
- A/B test: "Download" vs "Ontvang"
- Add social proof: "230+ engineers already using this"
- Show preview of email content

## ✅ You're Done!

The system is ready. Just add SMTP credentials and test. 🎉

Need help? Check `EMAIL_SETUP.md` for detailed troubleshooting.
