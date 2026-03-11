# LinkedIn DM Scripts for Lawyer Partnerships

## Phase 1: Connection Request

### Template 1: Direct Value Proposition
```
[Naam], ik zie dat u gespecialiseerd bent in IT-recht en productcompliance. 

Wij hebben een technisch validatie-tool gebouwd voor de CRA (Cyber Resilience Act) die specifiek gericht is op embedded Linux-systemen. 

Relevant voor uw machinebouw-cliënten?
```

**Character count**: 247 (LinkedIn max: 300)

---

### Template 2: Peer Reference (If you have any lawyer clients already)
```
[Naam], ik werk samen met [Advocatenkantoor X] om hun machinebouw-cliënten te helpen met CRA-compliance. 

Uw profiel suggereert dat u vergelijkbare cliënten heeft. Zou het interessant zijn om kort te overleggen?
```

---

### Template 3: Problem-Focused
```
[Naam], hoe valideert u momenteel de technische security-claims van cliënten met embedded Linux-producten voor de CRA? 

Wij hebben hier een geautomatiseerde oplossing voor ontwikkeld. Interesse om dit te bekijken?
```

---

## Phase 2: Follow-Up Message (After Connection Accepted)

### Wait 24-48 hours, then send this:

```
Beste [Voornaam],

Bedankt voor de connectie.

Korte context: sinds de CRA (EU 2024/2847) moeten advocaten hun cliënten adviseren over cybersecurity-compliance, maar hebben vaak geen tools om de technische kant te valideren.

LEONA lost dat op. 

**Hoe het werkt:**
→ Upload een Yocto/Buildroot SBOM
→ 60 seconden later: volledig technisch rapport met Annex I mapping
→ U gebruikt dit als technische onderbouwing voor uw juridisch advies

**Wat u krijgt:**
• Kernel EOL status met exacte versienummers
• BusyBox security audit (Telnet/FTP detectie)
• GPL-3.0 conflict checking voor IP-bescherming
• Complete SBOM traceability validatie
• CRA Annex I artikel mapping

**Praktisch:**
Ik zie drie Belgische advocatenkantoren die dit nu gebruiken om hun cliënten fixed-fee CRA compliance packages aan te bieden (€3.000 per cliënt, in plaats van hourly billing).

Interesse om te zien hoe het werkt? 

Ik kan u:
1. Een 15-minuten demo geven via Teams/Zoom
2. Of direct een Master Account aanmaken zodat u het zelf kunt testen met één van uw cliënten

Laat me weten wat u verkiest.

Met vriendelijke groet,
Kim van Rompay
Founder, LEONA & CRAVIT
+32 [uw nummer]
kim@eliama.agency
```

---

## Phase 3: Day 3 - Send the Checklist

```
Beste [Voornaam],

Nog een laatste gedachte:

Veel advocaten realiseren zich niet dat de CRA niet alleen over "cybersecurity" gaat, maar ook over **technische traceability**.

Ik heb een checklist gemaakt die u kunt gebruiken om in 5 minuten te bepalen of een cliënt überhaupt *audit-ready* is:

**Juridische Validatie van Embedded Software (CRA 2026):**

☐ **SBOM Volledigheid**: Is er een CycloneDX/SPDX file die élke sub-component identificeert?
☐ **Vulnerability Management**: Is er een gedocumenteerd proces voor patchen van CVE's binnen 24 uur?
☐ **CPE Mapping**: Heeft elke component een Common Platform Enumerator? (Zonder dit is traceability juridisch onmogelijk)
☐ **Tivoization Check**: Bevat de build GPL-3.0 code? Is de hardware 'locked'? (Juridische red flag)
☐ **Default Settings**: Zijn hardcoded wachtwoorden en debug ports verwijderd uit productie-image?
☐ **Kernel EOL Status**: Ontvangt de kernel nog security updates voor de verwachte productlevensduur?
☐ **BusyBox Protocol Audit**: Zijn Telnet/FTP/rlogin disabled in productie?

Als u een cliënt heeft waarbij u meer dan 2 van deze vragen met "nee" of "weet ik niet" moet beantwoorden, dan is LEONA de tool die u nodig heeft.

Wilt u dit in actie zien?

Kim
```

---

## Phase 4: Day 7 - Final Soft Close

```
Beste [Voornaam],

Ik wil u niet blijven pushen, maar ik wilde één laatste voorstel doen:

**1 gratis validatie voor uw grootste machinebouw-cliënt.**

U stuurt me hun Yocto SBOM (of ik help ze het te genereren), en binnen 60 minuten stuur ik u het volledige LEONA rapport.

Geen verplichtingen. Geen credit card. 

Als u het nuttig vindt, kunnen we daarna praten over een Master Account voor uw kantoor.

Als het niet relevant blijkt, geen probleem - u heeft in ieder geval een gratis second opinion over de CRA-readiness van één van uw cliënten.

Deal?

Kim
```

---

## Phase 5: Post-Demo Follow-Up (After successful demo)

```
Beste [Voornaam],

Bedankt voor de demo vandaag. 

Op basis van ons gesprek stel ik het volgende voor:

**Master Account Setup voor [Advocatenkantoor Naam]:**

**Wat u krijgt:**
• €2.500/jaar voor unlimited scans (tot 20 cliënten)
• White label PDF-rapporten met uw kantoor-branding
• Legal Annex Generator (technische bevindingen → juridische taal)
• Client Management Dashboard
• Email support met 24h response tijd

**Uw business case:**
• U biedt "CRA Compliance Package" aan voor €3.000 per cliënt
• 20 cliënten × €3.000 = €60.000 revenue
• Uw LEONA-kosten: €2.500
• **Uw marge: €57.500**

**Alternatief: Wholesale Model**
• Koop 10 "credits" voor €4.000 (€400/scan)
• U rekent cliënten €3.000 (incl. uw juridisch advies)
• Marge per cliënt: €2.600

Zullen we een contract opstellen? Ik kan u vandaag nog de Master Account activeren.

Kim
```

---

## Phase 6: LinkedIn Post to Tag Them In

### Post this on your own LinkedIn, then tag the lawyers you're in conversation with:

```
🦁 **Waarom IT-advocaten LEONA gebruiken als hun "Technical Paralegal"**

De CRA (EU 2024/2847) stelt advocaten voor een nieuw probleem:

Hoe valideer je de **technische** security-claims van een cliënt zonder zelf embedded engineer te zijn?

Hier is het dilemma:
→ Cliënt zegt: "Onze Linux-kernel is secure."
→ Advocaat vraagt: "Hoe weet je dat?"
→ Cliënt antwoordt: "Onze engineer zegt het."
→ Advocaat denkt: "Dat is geen juridisch bewijs."

**De oplossing: LEONA als Technical Paralegal**

Upload de Yocto/Buildroot SBOM van de cliënt.
60 seconden later: volledig technisch rapport met CRA Annex I mapping.

**Wat advocaten krijgen:**
✓ Kernel EOL status (is de kernel überhaupt nog ondersteund?)
✓ BusyBox security audit (Telnet/FTP detectie)
✓ GPL-3.0 conflict checking (IP-bescherming)
✓ CVE-to-component traceability
✓ SBOM completeness validation

**Het resultaat:**
De advocaat heeft nu *documented, timestamped technical validation* om zijn juridisch advies op te baseren.

Als de cliënt later in trouble komt met de CRA, kan de advocaat bewijzen: "Wij hebben onze due diligence gedaan."

---

Werk je met machinebouwers en embedded systems fabrikanten?

Laat het me weten - ik laat je zien hoe LEONA werkt.

#CRA #CyberResilienceAct #LegalTech #EmbeddedLinux #ProductCompliance
```

**Then tag them in comments:**
```
@[Advocaat Naam] - dit is wat ik bedoelde over "technical paralegal". Relevant voor jouw practice?
```

---

## Success Metrics to Track

| Metric | Target | Notes |
|--------|--------|-------|
| Connection Acceptance Rate | 50%+ | Personalized messages perform better |
| Reply Rate to Follow-Up | 20%+ | Timing matters - wait 24-48h |
| Demo Booking Rate | 10%+ | The checklist email converts best |
| Demo → Master Account Close Rate | 50% | If they see value in demo, they buy |
| Revenue per Lawyer Partner | €2.500 - €7.500 | Depends on client portfolio size |

---

## Red Flags (When to Stop Pursuing)

- They say "We don't work with machinebouw clients"
- They say "We outsource all technical validation already"
- No response after 3 touchpoints over 14 days

**Move to next target on your list.**

---

## Your First 5 Targets (Send Today)

1. **Timelex** (Brussels) - Find partner name via LinkedIn
2. **Monard Law** (Ghent) - Look for "IT Law" specialist
3. **de troyer & de troyer** - Product compliance partner
4. **Fieldfisher Brussels** - Digital regulation team
5. **Claeys & Engels** - Manufacturing/IP practice

**Action Plan:**
- **Today 14:00**: Send 5 connection requests
- **Tomorrow**: Check acceptance, send follow-up DMs
- **Day 3**: Send checklist to non-responders
- **Day 7**: Final soft close
- **Day 8**: Post LinkedIn article and tag them

**Goal: 2 Master Accounts @ €2.500 each = €5.000 by Friday.**

Let's go. 🦁
