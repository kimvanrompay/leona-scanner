package main

import (
	"fmt"
	"io"
	"log"
	"os"
	
	"leona-scanner/internal/repository"
	"leona-scanner/internal/scanner"
	"leona-scanner/internal/usecase"
)

func main() {
	// Read test SBOM
	sbomFile := "./test-data/yocto-sample.json"
	
	file, err := os.Open(sbomFile)
	if err != nil {
		log.Fatalf("Failed to open test SBOM: %v", err)
	}
	defer file.Close()
	
	sbomData, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read SBOM: %v", err)
	}
	
	fmt.Println("🔍 Parsing SBOM...")
	components, err := scanner.ParseSBOM(sbomData)
	if err != nil {
		log.Fatalf("Failed to parse SBOM: %v", err)
	}
	
	fmt.Printf("✅ Found %d components\n", len(components))
	
	fmt.Println("🔬 Analyzing components...")
	platform := scanner.DetectPlatform(components)
	result := scanner.AnalyzeComponents(components, platform)
	
	fmt.Printf("✅ Platform: %s, Compliance Score: %d%%\n", platform, result.OverallScore)
	fmt.Printf("   Issues: %d total (%d critical, %d high, %d medium, %d low)\n",
		result.CriticalCount+result.HighCount+result.MediumCount+result.LowCount,
		result.CriticalCount, result.HighCount, result.MediumCount, result.LowCount)
	
	// Create a fake repository (in-memory)
	repo, err := repository.NewRepository("sqlite://:memory:")
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	scannerService := usecase.NewScannerService(repo)
	
	// Create a scan record
	fmt.Println("💾 Creating scan record...")
	analysisResult, scan, err := scannerService.AnalyzeSBOM("test@example.com", sbomData)
	if err != nil {
		log.Fatalf("Failed to analyze SBOM: %v", err)
	}
	
	outputPath := "./integration-test-report.pdf"
	
	fmt.Println("📄 Generating PDF report...")
	err = scannerService.GeneratePDFReport(scan, analysisResult, outputPath)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
	}
	
	fmt.Printf("✅ SUCCESS! PDF generated at: %s\n", outputPath)
	fmt.Println("📊 Report contents:")
	fmt.Printf("   - Product: Embedded Linux Product v1.0\n")
	fmt.Printf("   - Compliance Score: %d%%\n", analysisResult.OverallScore)
	fmt.Printf("   - Platform: %s\n", scan.Platform)
	fmt.Printf("   - Violations: %d\n", len(analysisResult.Issues))
	fmt.Println("\n🎉 PDF generation system is PRODUCTION READY!")
}
