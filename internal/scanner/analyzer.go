package scanner

import (
	"fmt"
	"strings"

	"leona-scanner/internal/services"
)

// Component represents a software component from an SBOM
type Component struct {
	Name    string
	Version string
	License string
	CPE     string
	PURL    string
}

// AnalysisResult contains the compliance analysis results
type AnalysisResult struct {
	DetectedPlatform string
	TotalComponents  int
	OverallScore     int
	Status           string // CONFORM, VOORWAARDELIJK CONFORM, NIET-CONFORM
	Issues           []Issue
	CriticalCount    int
	HighCount        int
	MediumCount      int
	LowCount         int
	Recommendations  []string
	Vulnerabilities  []ComponentVulnerability // CVE vulnerability data
}

// Issue represents a compliance issue found during analysis
type Issue struct {
	Severity       string // CRITICAL, HIGH, MEDIUM, LOW
	Category       string // TRACEABILITY, IP_RISK, VERSION, SECURITY
	Component      string
	Description    string
	Recommendation string
}

// ComponentVulnerability represents CVE vulnerabilities found in a component
type ComponentVulnerability struct {
	Component string
	CPE       string
	CVECount  int
	Critical  int
	High      int
	Medium    int
	Low       int
	CVEs      []services.CVEResult
}

// Platform types
const (
	PlatformYocto    = "YOCTO"
	PlatformZephyr   = "ZEPHYR"
	PlatformFreeRTOS = "FREERTOS"
	PlatformGeneric  = "GENERIC"
)

// DetectPlatform identifies the embedded platform from SBOM components
func DetectPlatform(components []Component) string {
	for _, c := range components {
		name := strings.ToLower(c.Name)

		// Yocto detection
		if strings.Contains(name, "yocto") || strings.Contains(name, "poky") ||
			strings.Contains(name, "meta-") {
			return PlatformYocto
		}

		// Zephyr detection
		if strings.Contains(name, "zephyr") || strings.Contains(name, "zephyr-kernel") {
			return PlatformZephyr
		}

		// FreeRTOS detection
		if strings.Contains(name, "freertos") || strings.Contains(name, "free-rtos") {
			return PlatformFreeRTOS
		}
	}

	return PlatformGeneric
}

// AnalyzeComponents executes differential analysis based on detected platform
func AnalyzeComponents(components []Component, platform string) AnalysisResult {
	return AnalyzeComponentsWithCVE(components, platform, nil)
}

// AnalyzeComponentsWithCVE executes differential analysis with optional CVE vulnerability lookup
func AnalyzeComponentsWithCVE(components []Component, platform string, cveService *services.CVEService) AnalysisResult {
	score := 100
	var issues []Issue
	var recommendations []string

	// Universal CRA Rules (apply to all platforms)
	for _, c := range components {
		// Article 11: Software Bill of Materials - CPE Traceability
		if c.CPE == "" && c.PURL == "" {
			score -= 10
			issues = append(issues, Issue{
				Severity:    "HIGH",
				Category:    "TRACEABILITY",
				Component:   c.Name,
				Description: fmt.Sprintf("ONTBREKENDE TRACEERBAARHEID: %s heeft geen CPE of PURL identifier (CRA Article 11 vereiste).", c.Name),
			})
		}

		// Version control check
		if c.Version == "" {
			score -= 5
			issues = append(issues, Issue{
				Severity:    "MEDIUM",
				Category:    "VERSION",
				Component:   c.Name,
				Description: fmt.Sprintf("Geen versie-informatie voor %s. Dit bemoeilijkt vulnerability tracking.", c.Name),
			})
		}

		// License validation
		if c.License == "" {
			score -= 3
			issues = append(issues, Issue{
				Severity:    "LOW",
				Category:    "IP_RISK",
				Component:   c.Name,
				Description: fmt.Sprintf("Ontbrekende licentie-informatie voor %s.", c.Name),
			})
		}

		// GPL-3.0 copyleft detection
		if strings.Contains(strings.ToUpper(c.License), "GPL-3") ||
			strings.Contains(strings.ToUpper(c.License), "AGPL") {
			score -= 8
			issues = append(issues, Issue{
				Severity:    "HIGH",
				Category:    "IP_RISK",
				Component:   c.Name,
				Description: fmt.Sprintf("IP-RISICO: %s gebruikt GPL-3.0/AGPL (Copyleft). Dit kan distributieverplichtingen opleggen.", c.Name),
			})
			recommendations = append(recommendations, fmt.Sprintf("Overweeg een alternatief voor %s met een permissieve licentie (MIT, Apache-2.0, BSD).", c.Name))
		}
	}

	// Platform-specific rules
	switch platform {
	case PlatformYocto:
		analyzeYocto(components, &score, &issues, &recommendations)
	case PlatformZephyr:
		analyzeZephyr(components, &score, &issues, &recommendations)
	case PlatformFreeRTOS:
		analyzeFreeRTOS(components, &score, &issues, &recommendations)
	}

	// CVE Vulnerability Analysis (if CVE service is available)
	var vulnerabilities []ComponentVulnerability
	if cveService != nil {
		for _, c := range components {
			if c.CPE != "" {
				cves, err := cveService.LookupByCPE(c.CPE)
				if err == nil && len(cves) > 0 {
					// Count vulnerabilities by severity
					critical, high, medium, low := 0, 0, 0, 0
					for _, cve := range cves {
						switch cve.Severity {
						case "CRITICAL":
							critical++
						case "HIGH":
							high++
						case "MEDIUM":
							medium++
						case "LOW":
							low++
						}
					}

					vulnerabilities = append(vulnerabilities, ComponentVulnerability{
						Component: c.Name,
						CPE:       c.CPE,
						CVECount:  len(cves),
						Critical:  critical,
						High:      high,
						Medium:    medium,
						Low:       low,
						CVEs:      cves,
					})

					// Add CVE issues to the overall issue list
					if critical > 0 || high > 0 {
						severity := "HIGH"
						if critical > 0 {
							severity = "CRITICAL"
							score -= 15 // Critical CVEs significantly impact score
						} else {
							score -= 10 // High CVEs moderately impact score
						}

						issues = append(issues, Issue{
							Severity:    severity,
							Category:    "SECURITY",
							Component:   c.Name,
							Description: fmt.Sprintf("BEVEILIGINGSRISICO: %s heeft %d CVE(s) gevonden - %d critical, %d high, %d medium, %d low", c.Name, len(cves), critical, high, medium, low),
						})
						recommendations = append(recommendations, fmt.Sprintf("Update %s om %d bekende beveiligingslekken op te lossen.", c.Name, len(cves)))
					}
				}
			}
		}
	}

	// Determine CRA status
	status := "NIET-CONFORM"
	if score >= 90 {
		status = "CONFORM"
	} else if score >= 75 {
		status = "VOORWAARDELIJK CONFORM"
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	// Calculate severity counts
	criticalCount, highCount, mediumCount, lowCount := 0, 0, 0, 0
	for _, issue := range issues {
		switch issue.Severity {
		case "CRITICAL":
			criticalCount++
		case "HIGH":
			highCount++
		case "MEDIUM":
			mediumCount++
		case "LOW":
			lowCount++
		}
	}

	return AnalysisResult{
		DetectedPlatform: platform,
		TotalComponents:  len(components),
		OverallScore:     score,
		Status:           status,
		Issues:           issues,
		CriticalCount:    criticalCount,
		HighCount:        highCount,
		MediumCount:      mediumCount,
		LowCount:         lowCount,
		Recommendations:  recommendations,
		Vulnerabilities:  vulnerabilities,
	}
}

// analyzeYocto applies Yocto-specific compliance rules
func analyzeYocto(components []Component, score *int, issues *[]Issue, recommendations *[]string) {
	hasLTSKernel := false

	for _, c := range components {
		name := strings.ToLower(c.Name)

		// Check for LTS kernel usage
		if strings.Contains(name, "linux-kernel") || strings.Contains(name, "linux-yocto") {
			// Check if it's an LTS version (simplified check)
			if strings.Contains(c.Version, "lts") ||
				strings.Contains(c.Version, "5.15") ||
				strings.Contains(c.Version, "6.1") ||
				strings.Contains(c.Version, "6.6") {
				hasLTSKernel = true
			} else {
				*score -= 15
				*issues = append(*issues, Issue{
					Severity:    "CRITICAL",
					Category:    "SECURITY",
					Component:   c.Name,
					Description: "Yocto kernel is niet LTS. Dit verhoogt security patch management risico's.",
				})
				*recommendations = append(*recommendations, "Gebruik een LTS kernel versie (5.15, 6.1, of 6.6) voor langdurige ondersteuning.")
			}
		}

		// Check for meta-security layer
		if strings.Contains(name, "meta-security") {
			*recommendations = append(*recommendations, "✓ Meta-security layer gedetecteerd - uitstekend voor hardening.")
		}
	}

	if !hasLTSKernel {
		*recommendations = append(*recommendations, "Overweeg upgrade naar Yocto LTS kernel voor compliance met security update vereisten.")
	}
}

// analyzeZephyr applies Zephyr RTOS-specific compliance rules
func analyzeZephyr(components []Component, score *int, issues *[]Issue, recommendations *[]string) {
	hasZephyrKernel := false

	for _, c := range components {
		name := strings.ToLower(c.Name)

		// Check for Zephyr kernel
		if strings.Contains(name, "zephyr") && strings.Contains(name, "kernel") {
			hasZephyrKernel = true

			// Version must be specified for RTOS kernel
			if c.Version == "" {
				*score -= 20
				*issues = append(*issues, Issue{
					Severity:    "CRITICAL",
					Category:    "VERSION",
					Component:   c.Name,
					Description: "KRITIEK: Geen versiebeheer op Zephyr RTOS kernel gedetecteerd. Dit is essentieel voor traceerbaarheid.",
				})
			}

			// Check if it's a recent stable version
			if strings.HasPrefix(c.Version, "2.") || strings.HasPrefix(c.Version, "1.") {
				*issues = append(*issues, Issue{
					Severity:    "MEDIUM",
					Category:    "SECURITY",
					Component:   c.Name,
					Description: "Verouderde Zephyr kernel versie. Overweeg upgrade naar Zephyr 3.x voor actuele security patches.",
				})
			}
		}

		// Check for common Zephyr vulnerabilities
		if strings.Contains(name, "mbedtls") || strings.Contains(name, "tinycrypt") {
			if c.CPE == "" {
				*score -= 10
				*issues = append(*issues, Issue{
					Severity:    "HIGH",
					Category:    "SECURITY",
					Component:   c.Name,
					Description: fmt.Sprintf("Crypto library %s heeft geen CPE voor CVE-tracking.", c.Name),
				})
			}
		}
	}

	if hasZephyrKernel {
		*recommendations = append(*recommendations, "Valideer Zephyr kernel configuratie (prj.conf) voor security hardening opties.")
	}
}

// analyzeFreeRTOS applies FreeRTOS-specific compliance rules
func analyzeFreeRTOS(components []Component, score *int, issues *[]Issue, recommendations *[]string) {
	hasFreeRTOSKernel := false

	for _, c := range components {
		name := strings.ToLower(c.Name)

		// Check for FreeRTOS kernel
		if strings.Contains(name, "freertos") {
			hasFreeRTOSKernel = true

			// Check for LTS version
			if !strings.Contains(c.Version, "LTS") && !strings.Contains(c.Version, "10.") && !strings.Contains(c.Version, "11.") {
				*score -= 12
				*issues = append(*issues, Issue{
					Severity:    "HIGH",
					Category:    "SECURITY",
					Component:   c.Name,
					Description: "FreeRTOS versie is niet LTS. Gebruik FreeRTOS LTS voor guaranteed security updates.",
				})
				*recommendations = append(*recommendations, "Upgrade naar FreeRTOS LTS (202210.01 of nieuwer) voor commerciële deployments.")
			}

			// Check for kernel version tracking
			if c.Version == "" {
				*score -= 15
				*issues = append(*issues, Issue{
					Severity:    "CRITICAL",
					Category:    "VERSION",
					Component:   c.Name,
					Description: "KRITIEK: Geen FreeRTOS kernel versie gespecificeerd.",
				})
			}
		}

		// Check for AWS IoT integration (common in FreeRTOS deployments)
		if strings.Contains(name, "aws") || strings.Contains(name, "iot") {
			if c.CPE == "" {
				*score -= 8
				*issues = append(*issues, Issue{
					Severity:    "MEDIUM",
					Category:    "TRACEABILITY",
					Component:   c.Name,
					Description: fmt.Sprintf("AWS/IoT component %s mist CPE voor vulnerability tracking.", c.Name),
				})
			}
		}
	}

	if hasFreeRTOSKernel {
		*recommendations = append(*recommendations, "Controleer FreeRTOSConfig.h voor security best practices (stack overflow detection, MPU settings).")
	}
}
