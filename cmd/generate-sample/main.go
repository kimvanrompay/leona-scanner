package main

import (
	"fmt"
	"log"
	"time"
	
	"leona-scanner/internal/services"
)

func main() {
	// Create a realistic, convincing sample report for TÜV/NoBo
	report := services.ComplianceReport{
		ProductName:     "Industriële Gateway MG-2400X",
		ProductVersion:  "3.2.1",
		ScanDate:        time.Now(),
		ComplianceScore: 78, // Good but not perfect - realistic
		RiskLevel:       "Midden",
		KernelVersion:   "Linux 6.6.15 LTS (support t/m dec 2029)",
		Distribution:    "Yocto Scarthgap (5.0)",
		Violations: []services.ComplianceViolation{
			{
				CRAArticle:       "Art. 14.1 (SBOM Traceability)",
				Status:           "PASS",
				TechnicalFinding: "Alle 347 componenten hebben volledige CPE/PURL traceability volgens NIST IR 8278A standaard",
				Remediation:      "✓ Voldoet aan CRA Annex I Part II.1 - geen actie vereist",
			},
			{
				CRAArticle:       "Art. 10.4 (Security Update Lifecycle)",
				Status:           "PASS",
				TechnicalFinding: "Kernel 6.6 LTS met support tot december 2029 (ruim na producteinde verwachting 2027)",
				Remediation:      "✓ Voldoet - 2 jaar margin boven minimale CRA vereiste",
			},
			{
				CRAArticle:       "Annex I.II.1 (Secure Boot)",
				Status:           "WARN",
				TechnicalFinding: "U-Boot secure boot geïmplementeerd, maar geen hardware root-of-trust (TPM/TEE)",
				Remediation:      "Optioneel: Voeg TPM 2.0 module toe voor volledige hardware-backed chain of trust",
			},
			{
				CRAArticle:       "Art. 14.2 (License Compliance)",
				Status:           "FAIL",
				TechnicalFinding: "Qt 5.15.2 (GPL-3.0) gelinkt met proprietary applicatie - Anti-Tivoization clause actief",
				Remediation:      "KRITIEK: Schakel over naar commerciële Qt licentie OF isoleer via IPC boundary (D-Bus)",
			},
			{
				CRAArticle:       "Annex I.II.2 (Default Credentials)",
				Status:           "PASS",
				TechnicalFinding: "Geen hardcoded credentials gevonden. Per-device unique SSH keys via provisioning script",
				Remediation:      "✓ Voldoet - best practice geïmplementeerd",
			},
			{
				CRAArticle:       "Art. 10.1 (Vulnerability Management)",
				Status:           "WARN",
				TechnicalFinding: "OpenSSL 3.0.8 bevat 2 LOW severity CVEs (CVE-2023-5363, CVE-2023-6129)",
				Remediation:      "Niet-kritiek maar patch naar 3.0.13 aanbevolen voor volledige compliance attestatie",
			},
			{
				CRAArticle:       "Annex I.III.1 (Telemetry & Logging)",
				Status:           "PASS",
				TechnicalFinding: "Security event logging via systemd-journald met remote syslog naar SIEM",
				Remediation:      "✓ Voldoet - geconfigureerd conform IEC 62443-4-2",
			},
			{
				CRAArticle:       "Art. 3.3 (Integrity Protection)",
				Status:           "WARN",
				TechnicalFinding: "dm-verity op rootfs actief, maar /var partition beschrijfbaar zonder integrity check",
				Remediation:      "Implementeer IMA/EVM voor runtime file integrity monitoring op /var",
			},
			{
				CRAArticle:       "Annex I.II.3 (Least Privilege)",
				Status:           "PASS",
				TechnicalFinding: "Alle applicaties draaien non-root met SELinux enforcing mode (strict policy)",
				Remediation:      "✓ Voldoet - defense-in-depth geïmplementeerd",
			},
			{
				CRAArticle:       "Art. 10.2 (CVE Disclosure)",
				Status:           "PASS",
				TechnicalFinding: "SBOM bevat NVD CVE mapping voor alle componenten met bekende vulnerabilities",
				Remediation:      "✓ Voldoet - CVE database up-to-date (laatste sync: vandaag)",
			},
		},
	}
	
	outputPath := "./static/sample-cra-report.pdf"
	
	fmt.Println("🏭 Generating TÜV-quality sample CRA compliance report...")
	fmt.Printf("   Product: %s v%s\n", report.ProductName, report.ProductVersion)
	fmt.Printf("   Platform: %s\n", report.Distribution)
	fmt.Printf("   Score: %d%% (%s risico)\n", report.ComplianceScore, report.RiskLevel)
	fmt.Printf("   Violations: %d items checked\n\n", len(report.Violations))
	
	if err := services.GenerateCRAVITReport(report, outputPath); err != nil {
		log.Fatalf("❌ Failed to generate PDF: %v", err)
	}
	
	fmt.Printf("✅ SUCCESS! Sample report generated: %s\n", outputPath)
	fmt.Println("📄 This PDF is designed to convince TÜV/NoBo auditors:")
	fmt.Println("   • Professional CRAVIT branding")
	fmt.Println("   • Realistic product details (industrial gateway)")
	fmt.Println("   • 78% compliance score (credible, not perfect)")
	fmt.Println("   • Mix of PASS/WARN/FAIL (shows thorough analysis)")
	fmt.Println("   • Specific remediation steps")
	fmt.Println("   • Legal liability shield statement")
	fmt.Println("   • Certificate attestation")
}
