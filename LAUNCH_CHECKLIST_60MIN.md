# 🦁 LEONA Launch Checklist - Hit €10K in 7 Days

**Date**: 11 March 2026  
**Target**: €10.000 revenue in 7 days  
**Strategy**: 2 lawyer Master Accounts (€5.000) + 10 direct scans (€4.990)

---

## ✅ COMPLETED (Already Done by Oz)

- [x] Landing page with Davis Polk design at leona-cravit.be
- [x] HTMX scanner with theatrical loading effect
- [x] Free report lead magnet page
- [x] Insights/Kennisbank with LinkedIn-ready content
- [x] Scan results partial with upsell pricing tiers
- [x] Legal disclaimer document for PDFs
- [x] Lawyer partnership strategy document
- [x] LinkedIn DM scripts (6-phase approach)
- [x] Technical paralegal positioning
- [x] Dual lead magnets (engineers + lawyers)

---

## 🚀 HOUR 1: Deploy & Verify (14:00 - 15:00)

### Minute 0-15: Push to Production

```bash
cd /Volumes/1T\ SSD/PRODUCTS/leona-scanner
git push origin master
```

**Wait for Coolify to deploy** (check leona-cravit.be)

### Minute 15-30: Smoke Test

- [ ] Visit https://leona-cravit.be
- [ ] Test file upload on scanner (use test SBOM)
- [ ] Verify HTMX loading animation appears
- [ ] Check that scan results page renders
- [ ] Test all navbar links (/, /insights, /kennisbank, /free-report)
- [ ] Verify mobile responsive design

### Minute 30-45: SSL & Payment Check

- [ ] Confirm HTTPS certificate is valid
- [ ] Test Mollie payment link (if implemented)
- [ ] Verify mailto: links work for Tier 1 pricing
- [ ] Check Open Graph tags for LinkedIn sharing

### Minute 45-60: Backup & Monitoring

- [ ] Create database backup
- [ ] Set up uptime monitoring (UptimeRobot or similar)
- [ ] Test contact email: support@craleona.be receives mail

---

## 🎯 HOUR 2: LinkedIn Lawyer Outreach (15:00 - 16:00)

### Minute 0-20: Research Target Lawyers

Use LinkedIn to find:

1. **Timelex (Brussels)** - Partner in IT Law
2. **Monard Law (Ghent)** - IP/IT specialist
3. **de troyer & de troyer** - Product compliance lawyer
4. **Fieldfisher Brussels** - Digital regulation partner
5. **Claeys & Engels** - Manufacturing/IP practice

**Record names in spreadsheet:**

| Firm | Lawyer Name | LinkedIn URL | Connection Sent | Status |
|------|-------------|--------------|-----------------|--------|
| Timelex | [TBD] | [URL] | [ ] | Pending |
| Monard Law | [TBD] | [URL] | [ ] | Pending |
| ... | ... | ... | [ ] | ... |

### Minute 20-40: Send 5 Connection Requests

**Use Template 1 from LINKEDIN_LAWYER_DM_SCRIPTS.md:**

```
[Naam], ik zie dat u gespecialiseerd bent in IT-recht en productcompliance. 

Wij hebben een technisch validatie-tool gebouwd voor de CRA (Cyber Resilience Act) die specifiek gericht is op embedded Linux-systemen. 

Relevant voor uw machinebouw-cliënten?
```

**Action:**
- [ ] Send connection request to Lawyer 1
- [ ] Send connection request to Lawyer 2
- [ ] Send connection request to Lawyer 3
- [ ] Send connection request to Lawyer 4
- [ ] Send connection request to Lawyer 5

### Minute 40-60: LinkedIn Profile Optimization

- [ ] Update your LinkedIn headline: "Founder @ LEONA | CRA Compliance Automation for Embedded Linux"
- [ ] Update profile summary with "Technical Paralegal" positioning
- [ ] Add leona-cravit.be to website field
- [ ] Set profile to "Open to Business" (for lawyer inquiries)

---

## 📱 DAY 2-7: Follow-Up Sequence

### Day 2 (Tomorrow)
- [ ] Check LinkedIn for connection acceptances
- [ ] Send Phase 2 DM to accepted connections (see LINKEDIN_LAWYER_DM_SCRIPTS.md)
- [ ] Post first LinkedIn article: "CRA is Coming. Ready or Not."
- [ ] Share /insights page content as LinkedIn post

### Day 3 (Wednesday)
- [ ] Send "Checklist Email" to non-responders (Phase 3)
- [ ] Post second LinkedIn article: "Technical Paralegal" concept
- [ ] Tag lawyers you've connected with in comments

### Day 4 (Thursday)
- [ ] Email lead magnet subscribers from free-report page
- [ ] Post LinkedIn article about BusyBox security risks
- [ ] Respond to any LinkedIn replies

### Day 5 (Friday)
- [ ] Send "Final Soft Close" DM to non-responders (Phase 4)
- [ ] Post LinkedIn success story (if you have any demo)
- [ ] Prepare weekend follow-ups

### Day 7 (Monday)
- [ ] **Revenue check**: Are you at €10K?
- [ ] Follow up with interested lawyers for Master Account close
- [ ] Send invoices for any paid scans

---

## 💰 Revenue Tracking Spreadsheet

| Date | Source | Client/Lawyer | Type | Amount | Status |
|------|--------|---------------|------|--------|--------|
| 11-Mar | Direct | [Company] | Single Scan | €499 | Pending |
| 12-Mar | Lawyer | [Law Firm] | Master Account | €2.500 | Negotiation |
| ... | ... | ... | ... | ... | ... |
| **TOTAL** | | | | **€X.XXX** | |

**Target: €10.000 by 18 March**

---

## 📊 Conversion Funnel Tracking

### Lead Magnet Performance
- **Free Report Downloads**: [X]
- **Email Capture Rate**: [X%]
- **Free → Paid Conversion**: [X%]

### Lawyer Outreach Performance
- **Connection Requests Sent**: 5
- **Connection Acceptance Rate**: [X%]
- **DM Reply Rate**: [X%]
- **Demo Bookings**: [X]
- **Master Accounts Closed**: [X]

### Direct Scanner Usage
- **Unique Visitors**: [X]
- **SBOM Uploads**: [X]
- **Scan Completions**: [X]
- **Payment Initiated**: [X]
- **Payment Completed**: [X]

---

## 🔧 Technical Todos (If Needed)

### Lead Capture Endpoints (Backend)
If not already implemented:

- [ ] `/api/lead/engineer` - Capture engineer emails for BitBake layer
- [ ] `/api/lead/lawyer` - Capture lawyer emails for Annex I template
- [ ] Set up email automation (SendGrid/Postmark)
- [ ] Create actual downloadable lead magnets

### Payment Flow
- [ ] Implement Mollie checkout for Tier 1 (€499)
- [ ] Implement Mollie checkout for Tier 2 (€2.499)
- [ ] Implement Mollie checkout for Tier 3 (€6.999)
- [ ] Add Mollie webhook handler for payment confirmation
- [ ] Generate invoice PDF after payment

### PDF Generation
- [ ] Implement PDF report generation with Maroto
- [ ] Add legal disclaimer footer to all pages (see PDF_LEGAL_DISCLAIMER.md)
- [ ] Include CRA Annex I mapping table
- [ ] Add Remediation Roadmap section
- [ ] Generate "Golden Package" ZIP with .bbappend fixes

---

## 🚨 Emergency Fallback Plan

**If lawyer outreach is slow:**

### Plan B: Direct Engineer Outreach
- Post on Reddit: r/embedded, r/linux, r/yocto
- Post on Hacker News: "Show HN: CRA Compliance Scanner for Embedded Linux"
- Email direct to Flemish machinebouwers (find via Graydon)

### Plan C: Offer Free Audits for Testimonials
- "First 3 companies get free CRA scan in exchange for LinkedIn testimonial"
- Use testimonials to close lawyer partnerships faster

---

## 📞 Sales Script (If Someone Calls)

**Opening:**
> "Bedankt voor uw interesse in LEONA. Mag ik vragen: werkt u met embedded Linux in uw producten?"

**Discovery:**
> "Gebruikt u Yocto, Buildroot of Debian voor uw Linux-builds?"

**Pain Identification:**
> "Heeft u al een SBOM gegenereerd voor de CRA? Weet u of uw kernel end-of-life is?"

**Demo Offer:**
> "Ik kan u nu direct een gratis scan doen. Heeft u een manifest.json bij de hand?"

**Close:**
> "Als het rapport nuttig blijkt, kunnen we direct overstappen naar het volledige pakket. Akkoord?"

---

## 🎉 Success Criteria

### Week 1 Goals:
- [ ] **2 Master Accounts** signed (€2.500 × 2 = €5.000)
- [ ] **10 paid scans** completed (€499 × 10 = €4.990)
- [ ] **50+ email leads** captured from free report
- [ ] **3 LinkedIn posts** with 1000+ impressions each
- [ ] **1 testimonial** from satisfied client

**If you hit these numbers: €9.990 revenue in 7 days. You're there.**

---

## 📝 Post-Launch Learnings (Fill this out daily)

### What Worked:
- 
- 
- 

### What Didn't Work:
- 
- 
- 

### Adjustments for Tomorrow:
- 
- 
- 

---

## 🦁 Motivational Close

**You've built:**
- A professional SaaS product
- A gap in the market (CRA + Embedded Linux)
- A clear value proposition (60 seconds vs 10 hours)
- A dual revenue stream (lawyers + engineers)

**The only thing left to do: EXECUTE.**

No more coding. No more tweaking. 

**Send those LinkedIn messages NOW.**

One lawyer with 20 clients = Your €10K target.

Let's go. 🦁🚀
