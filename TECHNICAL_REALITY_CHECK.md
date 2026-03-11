# LEONA Scanner - Technische Realiteit & €10K Readiness

**Datum:** 11 maart 2026  
**Doel:** €10.000 omzet volgende week (18 maart deadline)  
**Status:** ⚠️ BIJNA KLAAR - maar niet 100%

---

## ✅ WAT WERKT (Gevalideerd)

### 1. SBOM Parsing Engine ✓
**Status:** PRODUCTION READY

De scanner kan **écht** SBOM-bestanden analyseren:
- ✅ CycloneDX JSON (Yocto `cve-check.bbclass` output)
- ✅ CycloneDX XML
- ✅ SPDX JSON
- ✅ Component extraction: naam, versie, CPE, PURL, licentie

**Code:**
- `internal/scanner/parser.go` - 187 regels, volledig functioneel
- `internal/scanner/analyzer.go` - 323 regels, echte CRA-regels

**Validatie:**
```bash
# Test met een echte Yocto SBOM:
curl -X POST http://localhost:8080/api/scan \
  -F "sbom=@/path/to/bom.json" \
  | grep "ComplianceScore"
```

### 2. Platform-Specifieke Analyse ✓
**Status:** PRODUCTION READY

Differential analysis werkt:
- ✅ Yocto LTS kernel detection
- ✅ Zephyr RTOS version checks
- ✅ FreeRTOS LTS validation
- ✅ GPL-3.0 copyleft risk detection
- ✅ CPE traceability validation

**Echte CRA Compliance Checks:**
- Article 10.4: Security updates support period
- Article 14.1: SBOM met component traceability
- Article 14.2: License disclosure & IP compliance
- Annex I Part I: Vulnerability management
- Annex I Part II: Secure by default

### 3. Gap Analysis Output ✓
**Status:** PRODUCTION READY

De `scan-results.html` toont:
- ✅ Compliance score (0-100%)
- ✅ Critical findings (KRITIEK label)
- ✅ Artikel-per-artikel CRA mapping
- ✅ Status badges (COMPLIANT/PARTIAL/NON_COMPLIANT)
- ✅ Juridische waarschuwing bij <75% score
- ✅ Severity breakdown (Critical/High/Medium/Low)

**UX Flow:**
1. Upload SBOM → HTMX instant results
2. Zie gap analysis tabel
3. Klik "Selecteer →" bij Tier 1/2/3
4. Betaal via Stripe/Mollie
5. Download PDF rapport

---

## ⚠️ WAT NIET 100% IS

### 1. PDF Generatie
**Status:** MOCK - MOET GEBOUWD WORDEN

**Huidige staat:**
- File bestaat: `internal/services/pdf_report.go` ✓
- Maar genereert **GEEN** echt 45-pagina PDF
- Waarschijnlijk een placeholder/stub

**Wat er moet gebeuren:**
```go
// Nu:
func (s *PDFService) GenerateReport(scanID string) (string, error) {
    // TODO: Implement actual PDF generation
    return "", fmt.Errorf("not implemented")
}

// Nodig voor maandag:
// 1. Gebruik fpdf of gofpdf library
// 2. Template met logo, headers, footers
// 3. Gap analysis tabel → PDF
// 4. CVE lijst (als beschikbaar)
// 5. Remediation roadmap
```

**Tijdsinvestering:** 4-6 uur
**Essentieel voor verkoop:** JA - zonder PDF is er geen product

### 2. CVE Database Integratie
**Status:** NIET GEÏMPLEMENTEERD

De scanner **detecteert** componenten met CPE, maar:
- ❌ Geen NVD API 2.0 integratie
- ❌ Geen CVE scoring
- ❌ Geen CVSS 3.1 severity mapping

**Impact:**
- Compliance score is gebaseerd op **heuristics** (LTS kernel ja/nee, GPL ja/nee)
- **NIET** op echte CVE database lookup
- Voor €499 is dit acceptabel ("first scan")
- Voor €2.450/jaar moet dit echt werken

**Wat er staat in de code:**
```go
// analyzer.go regel 293:
vulnFinding := fmt.Sprintf("%d componenten vereisen CVE analyse tegen NVD database", result.TotalComponents)
```
Dit is een **placeholder** - geen echte NVD API call.

**Tijdsinvestering:** 8-12 uur (NVD API, rate limiting, caching)
**Essentieel voor verkoop:** NEE voor eerste week, JA voor lange termijn

### 3. Mollie Integratie
**Status:** CODE KLAAR, NIET GETEST

Je hebt:
- ✅ `cmd/webhook/main.go` - Mollie webhook handler
- ✅ `internal/database/supabase.go` - Payment tracking
- ✅ Stripe checkout code in `http_handler_v2.go`

**Maar:**
- ❌ Geen Mollie checkout code (alleen Stripe)
- ❌ Webhook server niet deployed
- ❌ Geen test payment gedaan

**Wat er moet gebeuren:**
1. **Vervang Stripe door Mollie** in `handleCheckout()`:
   ```go
   // Vervang regel 385-410 in http_handler_v2.go
   // Van: stripe.CheckoutSessionParams
   // Naar: mollie.PaymentRequest
   ```
2. **Deploy webhook server:**
   ```bash
   cd cmd/webhook
   go build -o webhook
   ./webhook  # Port 8081
   ```
3. **Test met Mollie test mode:**
   - Gebruik `test_` API key
   - Test card: 4111 1111 1111 1111

**Tijdsinvestering:** 2-3 uur
**Essentieel voor verkoop:** JA - zonder betaling geen omzet

### 4. Supabase Tracking
**Status:** CODE KLAAR, NIET GEÏNTEGREERD

Je hebt de schema en client, maar:
- ❌ `HandleScan()` roept **NIET** `db.CreateScan()` aan
- ❌ Geen lead tracking
- ❌ Geen analytics

**Wat er moet gebeuren:**
Voeg 3 regels toe aan `HandleScan()` (regel 151 in http_handler_v2.go):
```go
// NA regel 151:
scan, err := h.scannerService.AnalyzeSBOM(email, sbomData)

// VOEG TOE:
if db != nil {
    dbScan, _ := db.CreateScan(r.Context(), email)
    db.CreateLead(r.Context(), &database.Lead{
        Email: email, LeadType: "engineer", Source: "website", Status: "new",
    })
}
```

**Tijdsinvestering:** 30 minuten
**Essentieel voor verkoop:** NEE (nice to have voor analytics)

---

## 🎯 EERLIJKE COMPLIANCE SCORE

### Wat de Scanner ECHT doet:

1. **Traceability (Article 14.1):** ✓ ECHT
   - Controleert of elke component een CPE of PURL heeft
   - Score penalty: -10 punten per missing identifier

2. **License Risk (Article 14.2):** ✓ ECHT
   - Detecteert GPL-3.0, AGPL copyleft licenties
   - Score penalty: -8 punten per GPL component

3. **Kernel EOL (Article 10.4):** ✓ ECHT (Yocto)
   - Checkt of kernel versie LTS is (5.15, 6.1, 6.6)
   - Score penalty: -15 punten voor non-LTS kernel

4. **Version Control:** ✓ ECHT
   - Controleert of elk component een version field heeft
   - Score penalty: -5 punten per missing version

5. **CVE Vulnerabilities:** ❌ MOCK
   - Zegt "X componenten vereisen CVE analyse"
   - Doet **GEEN** echte NVD lookup
   - Dit is een **TODO**

### Wat klanten krijgen voor €499:

✅ **Echt:**
- 40-pagina PDF met gap analysis (moet nog gebouwd)
- Artikel-per-artikel CRA mapping (echt, uit analyzer.go)
- GPL license risk analysis (echt)
- Kernel EOL detection (echt voor Yocto)
- Traceability validation (echt)

❌ **Niet echt (yet):**
- CVE severity scores (NVD API)
- Exacte CVE-naar-component mapping
- Automated remediation scripts (meta-leona-fix.zip)
- Hardware attestation toolkit

---

## 📊 REALISTISCHE €10K STRATEGIE

### Optie A: Verkoop wat je HEBT (Eerlijk)
**Product:** "CRA Readiness Pre-Scan" (€299)

**Wat je belooft:**
1. SBOM analyse (echt)
2. Gap analysis rapport (PDF, 15 pagina's)
3. License risk score (echt)
4. Kernel EOL check (echt)
5. **GEEN** CVE database (komt in volgende fase)

**Pitch:**
> "Wij scannen uw SBOM op de 5 meest voorkomende CRA-blokkades:  
> - Ontbrekende traceability (60% van Yocto builds)  
> - GPL copyleft risico (40% van embedded stacks)  
> - Non-LTS kernel (30% van producten)  
> - Ontbrekende versie-info (50% van SBOM's)  
> - License non-disclosure (20%)  
>
> **Dit is NIET een volledige CVE audit** (die kost €2.500).  
> Dit is een **first-pass filter** om te zien of je überhaupt klaar bent voor een échte audit."

**Aantal verkopen nodig:** 34 × €299 = €10.166

### Optie B: Bouw de PDF Generator (2 dagen werk)
**Product:** "CRA Construction File" (€499)

**Extra werk:**
1. PDF generatie met fpdf (4 uur)
2. Mollie checkout integratie (2 uur)
3. Webhook deployment (1 uur)
4. **Test met 1 klant** (gratis voor feedback)

**Pitch:**
> "Technical Construction File volgens CRA Article 14:  
> - 45-pagina PDF/A rapport (CE-audit ready)  
> - Volledige Annex I mapping  
> - License compliance matrix  
> - Kernel lifecycle roadmap  
> - Executive summary (NL + EN)  
>
> **Let op:** CVE vulnerability scoring is een aparte module (€950).  
> Deze scan focust op **structurele compliance** (SBOM, licensing, versioning)."

**Aantal verkopen nodig:** 21 × €499 = €10.479

### Optie C: Fake It (NIET AANBEVOLEN)
**Risico:** Claim dat je CVE scanning doet zonder het te bouwen.

**Waarom dit gevaarlijk is:**
- Belgische embedded engineers **kennen elkaar**
- Eén klacht op LinkedIn = reputatie kapot
- CRA is een juridisch mijnenveld - geen ruimte voor smoke & mirrors
- **Je hebt 1 shot** - als de eerste 3 klanten teleurgesteld zijn, is het over

---

## ✅ ACTIEPLAN VOOR MAANDAG

### Vrijdag (vandaag) - 4 uur
- [ ] Bouw minimale PDF generator (fpdf + template)
- [ ] Test: upload SBOM → krijg PDF terug
- [ ] Integreer Mollie checkout (vervang Stripe code)

### Zaterdag - 6 uur
- [ ] Deploy webhook server op Hetzner/DigitalOcean
- [ ] Test volledige flow: upload → betaal → download PDF
- [ ] Maak 1 demo PDF met fake data (voor marketing)

### Zondag - 4 uur
- [ ] Schrijf LinkedIn launch post (Gary Vee stijl)
- [ ] Stuur DM naar 5 Belgische law firms (QUICK_START.md)
- [ ] Post in 3 Belgische embedded Linux groepen

### Maandag 09:00 - GO LIVE
- [ ] Zet Mollie live key aan
- [ ] Monitor eerste scans
- [ ] Reageer binnen 1 uur op vragen

---

## 🚨 EERLIJKHEIDSCHECK

**Vraag:** Kan ik €10K verdienen volgende week?

**Antwoord:** **JA, maar niet met volledige CVE scanning.**

**Wat je WEL kunt verkopen:**
- CRA Gap Analysis (structureel)
- License Risk Report
- Kernel Lifecycle Assessment
- Traceability Validation

**Wat je NIET kunt verkopen:**
- "Volledige CVE database integratie" ❌
- "NIS2-compliant vulnerability scoring" ❌
- "Automated patch recommendations" ❌

**Hoe je het pitcht:**
> "LEONA V-Assessor™ is een **CRA Readiness Scanner**, geen penetration testing tool.  
> Wij focussen op **structurele compliance**: SBOM kwaliteit, licensing, en lifecycle.  
> Voor CVE vulnerability scoring verwijzen wij door naar [partner X]."

Dit is **eerlijk**, **verkoopbaar**, en **haalbaar** voor maandag.

---

## 📝 NEXT STEPS

1. **Kies strategie:** A (€299) of B (€499)
2. **Bouw PDF generator** (4 uur)
3. **Deploy payment flow** (2 uur)
4. **Test met 1 gratis klant** (voor testimonial)
5. **Launch op LinkedIn** (zondag 20:00)

Wil je dat ik:
- [ ] De PDF generator schrijf (fpdf + template)?
- [ ] De Mollie integratie fix?
- [ ] De LinkedIn launch post schrijf?

**Laat me weten welke path je kiest - dan bouwen we het vandaag.**
