package usecase

import (
	"encoding/json"
	"fmt"

	"leona-scanner/internal/repository"
	"leona-scanner/internal/scanner"
	"leona-scanner/internal/services"
)

type ScannerService struct {
	repo *repository.Repository
}

func NewScannerService(repo *repository.Repository) *ScannerService {
	return &ScannerService{repo: repo}
}

// AnalyzeSBOM performs complete CRA compliance analysis
func (s *ScannerService) AnalyzeSBOM(email string, sbomData []byte) (*scanner.AnalysisResult, *repository.Scan, error) {
	// Parse SBOM file
	components, err := scanner.ParseSBOM(sbomData)
	if err != nil {
		return nil, nil, fmt.Errorf("SBOM parse fout: %w", err)
	}

	// Detect platform
	platform := scanner.DetectPlatform(components)

	// Analyze components
	result := scanner.AnalyzeComponents(components, platform)

	// Create or get lead
	lead, err := s.repo.CreateLead(email)
	if err != nil {
		return nil, nil, fmt.Errorf("lead aanmaken fout: %w", err)
	}

	// Serialize result to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("result serialisatie fout: %w", err)
	}

	// Store scan in database
	scan, err := s.repo.CreateScan(lead.ID, platform, result.OverallScore, sbomData, string(resultJSON))
	if err != nil {
		return nil, nil, fmt.Errorf("scan opslaan fout: %w", err)
	}

	return &result, scan, nil
}

// GetScanResult retrieves a previous scan result
func (s *ScannerService) GetScanResult(scanID string) (*scanner.AnalysisResult, *repository.Scan, error) {
	scan, err := s.repo.GetScanByID(scanID)
	if err != nil {
		return nil, nil, fmt.Errorf("scan ophalen fout: %w", err)
	}

	var result scanner.AnalysisResult
	if err := json.Unmarshal([]byte(scan.ResultJSON), &result); err != nil {
		return nil, nil, fmt.Errorf("result deserialisatie fout: %w", err)
	}

	return &result, scan, nil
}

// MarkScanPaid marks a scan as paid after successful payment
func (s *ScannerService) MarkScanPaid(scanID string) error {
	return s.repo.MarkScanAsPaid(scanID)
}

// GeneratePDFReport converts an AnalysisResult into a formal CRAVIT PDF
func (s *ScannerService) GeneratePDFReport(scan *repository.Scan, result *scanner.AnalysisResult, outputPath string) error {
	// Convert AnalysisResult to ComplianceReport format
	violations := s.buildViolations(result)
	
	riskLevel := "Laag"
	if result.OverallScore < 60 {
		riskLevel = "Hoog"
	} else if result.OverallScore < 80 {
		riskLevel = "Midden"
	}
	
	report := services.ComplianceReport{
		ProductName:     "Embedded Linux Product",
		ProductVersion:  "1.0",
		ScanDate:        scan.CreatedAt,
		ComplianceScore: result.OverallScore,
		RiskLevel:       riskLevel,
		Violations:      violations,
		KernelVersion:   s.detectKernel(result),
		Distribution:    scan.Platform,
	}
	
	return services.GenerateCRAVITReport(report, outputPath)
}

func (s *ScannerService) buildViolations(result *scanner.AnalysisResult) []services.ComplianceViolation {
	var violations []services.ComplianceViolation
	
	for _, issue := range result.Issues {
		status := "FAIL"
		if issue.Severity == "LOW" {
			status = "WARN"
		}
		
		article := "Annex I.II.1"
		if issue.Category == "TRACEABILITY" {
			article = "Art. 14.1 (SBOM)"
		} else if issue.Category == "IP_RISK" {
			article = "Art. 14.2 (License)"
		} else if issue.Category == "SECURITY" {
			article = "Art. 10.4 (Updates)"
		}
		
		violations = append(violations, services.ComplianceViolation{
			CRAArticle:       article,
			Status:           status,
			TechnicalFinding: issue.Description,
			Remediation:      issue.Recommendation,
		})
	}
	
	return violations
}

func (s *ScannerService) detectKernel(result *scanner.AnalysisResult) string {
	for _, issue := range result.Issues {
		if issue.Category == "SECURITY" && issue.Component != "" {
			return issue.Component
		}
	}
	return "Unknown"
}
