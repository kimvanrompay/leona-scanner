# LEONA Compliance Email Setup

## ✅ Configuration Complete

Your system has been updated to use **support@leonacompliance.be**

### What Was Changed

1. **Environment Variables** (`.env`)
   - Updated SMTP_USER to `support@leonacompliance.be`
   - Updated SMTP_FROM to `support@leonacompliance.be`
   - Server: `mail1.netim.hosting:465` (SSL/TLS)

2. **Code Updates**
   - All handler files updated to use new email
   - Domain changed from `leona-cravit.be` to `leonacompliance.be`
   - Email footers updated with new branding

### Email Configuration

```env
SMTP_HOST=mail1.netim.hosting
SMTP_PORT=465
SMTP_USER=support@leonacompliance.be
SMTP_PASS=YOUR_PASSWORD_HERE  # ⚠️  ADD YOUR PASSWORD!
SMTP_FROM=support@leonacompliance.be
```

## 🚀 Next Steps

### 1. Add Your Password to .env

```bash
# Open .env file
nano .env

# Update this line with your actual password:
SMTP_PASS=your-actual-password-here
```

### 2. Test Email Configuration

```bash
# Run the test script
cd /Volumes/1T\ SSD/PRODUCTS/leona-scanner
go run scripts/test-email.go
```

This will send a test email to kim@eliama.agency to verify the configuration.

### 3. Start the Server

```bash
# Build and run
go build -o leona-server cmd/server/main.go
./leona-server
```

### 4. Test Email Features

Visit these pages to test email functionality:

1. **CRA Assessment** (with email results)
   ```
   http://localhost:8080/cra-assessment
   ```

2. **Contact Form** (with confirmation email)
   ```
   http://localhost:8080/contact
   ```

3. **Free Audit Form**
   ```
   http://localhost:8080/free-audit
   ```

## 📧 Email Features Active

All these features will now use `support@leonacompliance.be`:

✅ CRA Assessment results emails
✅ Contact form confirmations
✅ Lead magnet downloads
✅ Demo request notifications
✅ Sample report delivery
✅ Risk assessment results
✅ Admin notifications (to kim@eliama.agency)

## 🔒 IMAP/Webmail Access

If you need to check incoming emails:

- **Webmail:** https://mail1.netim.hosting
- **IMAP Server:** mail1.netim.hosting:993 (SSL/TLS)
- **Username:** support@leonacompliance.be
- **Password:** [your password]

## 🧪 Testing Checklist

- [ ] Add password to `.env` file
- [ ] Run `go run scripts/test-email.go`
- [ ] Check kim@eliama.agency inbox
- [ ] Start server: `./leona-server`
- [ ] Test CRA assessment form
- [ ] Test contact form
- [ ] Verify email delivery

## 🆘 Troubleshooting

### Email Not Sending

```bash
# Check SMTP configuration
cat .env | grep SMTP

# Test connection manually
telnet mail1.netim.hosting 465
```

### Check Server Logs

```bash
# Run with verbose logging
LOG_VERBOSE=true ./leona-server
```

### Common Issues

1. **"SMTP not configured"** → Check `.env` file has all SMTP variables
2. **"Authentication failed"** → Verify password is correct
3. **"Connection timeout"** → Check firewall/network settings

## 📊 Email Stats Tracking

Admin notifications are sent to **kim@eliama.agency** for:

- New CRA assessments
- Contact form submissions
- Demo requests
- Lead magnet downloads
- Sample report downloads

---

**Ready to test?** Run: `go run scripts/test-email.go`
