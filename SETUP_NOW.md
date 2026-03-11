# ⚡ SETUP NOW (2 Minutes)

## ✅ What's Already Done

Your `.env` file is **90% configured** with the correct Netim settings:

```bash
SMTP_HOST=mail1.netim.hosting  ✅
SMTP_PORT=465                   ✅
SMTP_USER=support@leona-cravit.be  ✅
SMTP_FROM=support@leona-cravit.be  ✅
SMTP_PASS=your-app-specific-password  ⚠️ NEEDS YOUR PASSWORD
```

## 🔑 One Missing Piece: Your Email Password

### Step 1: Get Your Password (30 seconds)

Your email password is the **same password** you use to log in to:
**https://mail1.netim.hosting/webmail/**

If you don't remember it:
1. Go to [Netim control panel](https://www.netim.com)
2. Email → support@leona-cravit.be → Reset password
3. Copy the new password

### Step 2: Update `.env` (30 seconds)

Open `.env` and replace this line:
```bash
SMTP_PASS=your-app-specific-password
```

With your actual password:
```bash
SMTP_PASS=your_real_password_here
```

**Save the file.**

### Step 3: Test It! (1 minute)

Run the test script:

```bash
cd "/Volumes/1T SSD/PRODUCTS/leona-scanner"
./test-email.sh
```

This will:
1. ✅ Verify your `.env` settings
2. 🚀 Start the server
3. 📧 Prompt you to test the email form

Then:
1. Open http://localhost:8080 in your browser
2. Scroll to **"Voor Engineers en Juristen"**
3. Enter your email (e.g., `kim@youremail.com`)
4. Click **"Ontvang Layer"**

**Expected result:**
- ✅ Green message: "Check je inbox! We hebben de meta-leona layer naar kim@youremail.com gestuurd."
- 📧 Email arrives within 30 seconds
- Subject: **"Jouw meta-leona CRA Validator Layer"**

### Step 4: Verify Email Arrived

Check your inbox for:

**Subject:** 🚀 Jouw meta-leona Layer  
**From:** support@leona-cravit.be  
**Content:**
- Installation guide for meta-leona Yocto layer
- CTA buttons to download + link to scanner
- Professional HTML design

## ❌ If Email Doesn't Arrive

### Check 1: Server Logs
Look for errors in the terminal where you ran `./test-email.sh`:

**Error: "535 Authentication failed"**  
→ Wrong password in `.env`

**Error: "SMTP not configured"**  
→ Missing SMTP_ variables in `.env`

**Error: "Connection refused"**  
→ Firewall blocking port 465

### Check 2: Spam Folder
The email might be in spam (first time sending).

### Check 3: Test with Different Email
Try with a Gmail or Outlook address to rule out recipient issues.

## ✅ Success Checklist

Once you get the first test email:

- [x] Email works from localhost
- [ ] Test lawyer form too (scroll to same section, other card)
- [ ] Deploy to production
- [ ] Test from production URL
- [ ] Monitor first real leads in Supabase

## 🚀 Deploy to Production

Once localhost works, deploy with:

```bash
# Build
go build -o leona-server cmd/server/main.go

# Run on server
./leona-server
```

Make sure production `.env` has:
- `BASE_URL=https://leona-cravit.be`
- All SMTP_ variables (same as localhost)

## 📊 Track Your Leads

If you set up Supabase, check leads:

```sql
SELECT 
    email,
    lead_type,
    lead_magnet_requested,
    created_at
FROM leads
ORDER BY created_at DESC
LIMIT 10;
```

**Lead types:**
- `engineer` = meta-leona downloads
- `lawyer` = Annex I template downloads

## 💡 Marketing Your Lead Magnets

**For Engineers:**
- Post on Yocto mailing list
- Share in embedded Linux forums
- LinkedIn: "Free CRA compliance layer for Yocto"

**For Lawyers:**
- Email Belgian compliance law firms
- LinkedIn: "Free CRA Annex I template"
- Partner with industry associations

## 🎯 Next Steps After First Email

1. ✅ Verify email delivery (you're doing this now!)
2. 📦 Create actual downloadable files (optional):
   - `static/downloads/meta-leona.tar.gz`
   - `static/downloads/CRA_Annex_I_Mapping_Template.xlsx`
3. 🚀 Deploy to production
4. 📈 Launch marketing campaign
5. 💰 Convert leads to €499 scans

---

**Need help?**  
Check `EMAIL_SETUP.md` for detailed troubleshooting.

**Ready to go?**  
Run `./test-email.sh` now! 🚀
