package assessor

// Framework definieert de verschillende juridische kaders
type Framework string

const (
	CRA  Framework = "Cyber Resilience Act (EU 2024/2847)"
	CER  Framework = "Critical Entities Resilience Directive"
	NIS2 Framework = "Network and Information Security Directive 2"
)

// ComplianceMapping definieert de link tussen techniek en wet
type ComplianceMapping struct {
	RequirementID string    `json:"requirement_id"`
	Description   string    `json:"description"`
	Framework     Framework `json:"framework"`
	Status        string    `json:"status"` // COMPLIANT, PARTIAL, NON_COMPLIANT
	Remediation   string    `json:"remediation,omitempty"`
}

// TechnicalIssue representeert een gedetecteerd probleem in de build
type TechnicalIssue struct {
	IssueID     string
	Severity    string // CRITICAL, HIGH, MEDIUM, LOW
	Component   string
	Description string
}

// CrossFrameworkAssessment bevat alle compliance-mappings voor één build
type CrossFrameworkAssessment struct {
	CRAMappings  []ComplianceMapping `json:"cra_mappings"`
	CERMappings  []ComplianceMapping `json:"cer_mappings"`
	NIS2Mappings []ComplianceMapping `json:"nis2_mappings"`
	OverallScore int                 `json:"overall_score"`
}

// GetCrossFrameworkAdvice vertaalt een technische tekortkoming naar alle relevante kaders
func GetCrossFrameworkAdvice(issue TechnicalIssue) []ComplianceMapping {
	mappings := map[string][]ComplianceMapping{
		"KERNEL_OUTDATED": {
			{
				RequirementID: "CRA Annex I (3.1)",
				Description:   "Security by design & lifetime support. Kernel moet security updates ontvangen voor productlevensduur.",
				Framework:     CRA,
				Status:        "NON_COMPLIANT",
				Remediation:   "Upgrade naar LTS kernel (6.1.x of 6.6.x) met gegarandeerde backporting.",
			},
			{
				RequirementID: "NIS2 Art. 21",
				Description:   "Supply chain security & vulnerability handling. Leveranciers moeten bewijs leveren van patch-management.",
				Framework:     NIS2,
				Status:        "NON_COMPLIANT",
				Remediation:   "Implementeer geautomatiseerde CVE-monitoring met 24h patch-SLA.",
			},
			{
				RequirementID: "CER Art. 13",
				Description:   "Operational resilience van digitale componenten. Kritieke systemen moeten redundantie hebben.",
				Framework:     CER,
				Status:        "PARTIAL",
				Remediation:   "Voeg fallback-mechanisme toe voor kernel panic scenarios.",
			},
		},
		"MISSING_ENCRYPTION": {
			{
				RequirementID: "CRA Annex I (3.3)",
				Description:   "Confidentiality van data at rest. Gevoelige data moet encrypted zijn op filesystem-niveau.",
				Framework:     CRA,
				Status:        "NON_COMPLIANT",
				Remediation:   "Enable dm-crypt met LUKS2 voor data partities. Voeg `DISTRO_FEATURES += \"luks\"` toe.",
			},
			{
				RequirementID: "NIS2 Art. 21 (d)",
				Description:   "Cryptography en encryption standards. Gebruik van FIPS 140-2/3 gecertificeerde algoritmes verplicht.",
				Framework:     NIS2,
				Status:        "NON_COMPLIANT",
				Remediation:   "Integreer OpenSSL FIPS module via `meta-security` layer.",
			},
		},
		"GPL3_TIVOIZATION": {
			{
				RequirementID: "CRA Art. 18",
				Description:   "Technical documentation moet licentie-informatie bevatten. GPL-3.0 vereist disclosure van signing keys.",
				Framework:     CRA,
				Status:        "NON_COMPLIANT",
				Remediation:   "Verwijder GPL-3.0 componenten OF publiceer signing keys conform GPLv3 sectie 6.",
			},
		},
		"MISSING_SECURE_BOOT": {
			{
				RequirementID: "CRA Annex I (3.2)",
				Description:   "Protection against tampering. Boot-proces moet cryptografisch geverifieerd worden.",
				Framework:     CRA,
				Status:        "NON_COMPLIANT",
				Remediation:   "Implementeer UEFI Secure Boot of U-Boot verified boot met TPM 2.0.",
			},
			{
				RequirementID: "CER Art. 13 (b)",
				Description:   "Physical security van kritieke assets. Firmware moet beschermd zijn tegen unauthorized modification.",
				Framework:     CER,
				Status:        "NON_COMPLIANT",
				Remediation:   "Enable hardware root of trust (bijv. ARM TrustZone, Intel Boot Guard).",
			},
		},
		"NO_UPDATE_MECHANISM": {
			{
				RequirementID: "CRA Art. 10",
				Description:   "Vulnerability handling vereist een update-mechanisme voor security patches binnen productlevensduur.",
				Framework:     CRA,
				Status:        "NON_COMPLIANT",
				Remediation:   "Integreer SWUpdate of Mender via `meta-swupdate` layer.",
			},
			{
				RequirementID: "NIS2 Art. 21 (e)",
				Description:   "Incident response capability. Systemen moeten remote patches kunnen ontvangen binnen 24h.",
				Framework:     NIS2,
				Status:        "NON_COMPLIANT",
				Remediation:   "Implementeer A/B partition schema met atomic updates.",
			},
		},
		"TELNET_ENABLED": {
			{
				RequirementID: "CRA Annex I (3.1)",
				Description:   "Secure by default. Onveilige protocollen zoals Telnet moeten disabled zijn in productie-builds.",
				Framework:     CRA,
				Status:        "NON_COMPLIANT",
				Remediation:   "Voeg `PACKAGECONFIG_remove_pn-busybox = \"telnetd\"` toe aan local.conf.",
			},
		},
		"MISSING_SBOM": {
			{
				RequirementID: "CRA Art. 14",
				Description:   "SBOM vereist voor alle componenten. CycloneDX of SPDX formaat verplicht vanaf 2026.",
				Framework:     CRA,
				Status:        "NON_COMPLIANT",
				Remediation:   "Enable `INHERIT += \"create-spdx\"` in Yocto build configuration.",
			},
			{
				RequirementID: "NIS2 Art. 23",
				Description:   "Supply chain transparency. Leveranciers moeten volledige dependency-tree kunnen aantonen.",
				Framework:     NIS2,
				Status:        "NON_COMPLIANT",
				Remediation:   "Genereer SBOM met relationship mapping via `spdx-sbom-generator`.",
			},
		},
	}

	result, exists := mappings[issue.IssueID]
	if !exists {
		// Fallback voor onbekende issues
		return []ComplianceMapping{
			{
				RequirementID: "CRA General",
				Description:   "Onbekend technisch risico gedetecteerd. Verdere analyse vereist.",
				Framework:     CRA,
				Status:        "UNKNOWN",
			},
		}
	}

	return result
}

// GenerateCrossFrameworkReport genereert een volledig multi-framework assessment
func GenerateCrossFrameworkReport(issues []TechnicalIssue) CrossFrameworkAssessment {
	assessment := CrossFrameworkAssessment{
		CRAMappings:  []ComplianceMapping{},
		CERMappings:  []ComplianceMapping{},
		NIS2Mappings: []ComplianceMapping{},
	}

	// Map elk issue naar alle frameworks
	for _, issue := range issues {
		mappings := GetCrossFrameworkAdvice(issue)
		for _, mapping := range mappings {
			switch mapping.Framework {
			case CRA:
				assessment.CRAMappings = append(assessment.CRAMappings, mapping)
			case CER:
				assessment.CERMappings = append(assessment.CERMappings, mapping)
			case NIS2:
				assessment.NIS2Mappings = append(assessment.NIS2Mappings, mapping)
			}
		}
	}

	// Bereken overall score (0-100)
	totalMappings := len(assessment.CRAMappings) + len(assessment.CERMappings) + len(assessment.NIS2Mappings)
	compliantCount := 0

	for _, m := range assessment.CRAMappings {
		if m.Status == "COMPLIANT" {
			compliantCount++
		}
	}
	for _, m := range assessment.CERMappings {
		if m.Status == "COMPLIANT" {
			compliantCount++
		}
	}
	for _, m := range assessment.NIS2Mappings {
		if m.Status == "COMPLIANT" {
			compliantCount++
		}
	}

	if totalMappings > 0 {
		assessment.OverallScore = (compliantCount * 100) / totalMappings
	}

	return assessment
}

// GetFrameworkSummary geeft een menselijke samenvatting per framework
func GetFrameworkSummary(mappings []ComplianceMapping, framework Framework) string {
	compliant := 0
	nonCompliant := 0

	for _, m := range mappings {
		if m.Status == "COMPLIANT" {
			compliant++
		} else if m.Status == "NON_COMPLIANT" {
			nonCompliant++
		}
	}

	total := len(mappings)
	if total == 0 {
		return "Geen bevindingen voor dit framework."
	}

	status := "Kritiek"
	if nonCompliant == 0 {
		status = "Volledig Compliant"
	} else if nonCompliant < total/2 {
		status = "Gedeeltelijk Compliant"
	}

	return status
}
