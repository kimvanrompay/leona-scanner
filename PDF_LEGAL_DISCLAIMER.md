# PDF Report Legal Disclaimer

## Placement
This disclaimer must appear at the **bottom of every page** of the generated PDF report in 8pt font, gray text.

---

## Full Text (Dutch)

**JURIDISCHE DISCLAIMER & CONFORMITEITSVERKLARING**

**Status van dit Document:** Dit rapport is gegenereerd door LEONA (Powered by Eliama Agency) en is gebaseerd op een geautomatiseerde analyse van de door de gebruiker aangeleverde Software Bill of Materials (SBOM) en Bitbake-metadata op datum van **11 maart 2026**.

**Beperking van Aansprakelijkheid:** LEONA fungeert als een technisch ondersteuningsinstrument voor de Cyber Resilience Act (CRA). Hoewel dit rapport met de grootste zorgvuldigheid is samengesteld conform de op dit moment bekende standaarden van **Annex I (EU 2024/2847)**, vormt dit document geen vervanging voor een finaal juridisch oordeel door een erkend advocaat of een formele audit door een 'Notified Body'. LEONA aanvaardt geen aansprakelijkheid voor eventuele interpretatieverschillen door nationale toezichthouders of voor gebreken in de door de klant aangeleverde brongegevens.

**Intellectueel Eigendom:** De technische adviezen en Bitbake-configuraties in dit rapport zijn eigendom van de klant na volledige betaling. Het hergebruik van de LEONA-auditmethodiek voor derden is strikt verboden.

**VLAIO Context:** Dit rapport is gestructureerd als "Schriftelijk Advies" conform de kwaliteitsnormen van de kmo-portefeuille. De finale subsidieaanvraag blijft de verantwoordelijkheid van de gebruiker.

**Validatie:** Dit document is niet gecertificeerd door een externe toezichthoudende instantie. Voor CE-certificering is aanvullende validatie door een Notified Body vereist.

---

## Short Version (For Email Signatures)

```
LEONA & CRAVIT is geen advocatenkantoor. 
Wij zijn uw technische paralegal.
Consultancy duurt weken en kost duizenden euro's.
LEONA geeft u CRA-bewijs in 60 seconden.
```

---

## Go Implementation Location

File: `internal/usecase/pdf_service.go`

Function: `AddFooterToAllPages()`

```go
func (s *PDFService) AddFooterToAllPages(m *maroto.Maroto) {
    m.RegisterFooter(func() {
        m.Row(15, func() {
            m.Col(12, func() {
                m.Text(
                    "JURIDISCHE DISCLAIMER: Dit rapport is gegenereerd door LEONA op 11 maart 2026. "+
                    "LEONA fungeert als technisch ondersteuningsinstrument voor de CRA. "+
                    "Dit document vormt geen vervanging voor finaal juridisch oordeel door een erkend advocaat. "+
                    "Voor CE-certificering is validatie door een Notified Body vereist.",
                    props.Text{
                        Size:  6,
                        Style: fontstyle.Italic,
                        Color: color.NewWhite().Sub(50), // Gray
                        Align: align.Left,
                    },
                )
            })
        })
    })
}
```

---

## Why This Works

1. **Date-Stamped**: Protects against future CRA changes
2. **Notified Body Reference**: Shows professional understanding of CE process
3. **VLAIO Language**: Speaks directly to Flemish SME subsidy requirements
4. **IP Protection**: Prevents competitors from copying LEONA methodology
5. **Scope Limitation**: Clear boundary between technical analysis and legal advice
