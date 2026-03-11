package usecase

import (
	"encoding/json"
	"fmt"

	"leona-scanner/internal/repository"
	"leona-scanner/internal/scanner"
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
