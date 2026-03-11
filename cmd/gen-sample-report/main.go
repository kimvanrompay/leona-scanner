package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"leona-scanner/internal/services"
)

func main() {
	// Get output path from arguments or use default
	outputPath := "./static/sample_cravit_report.pdf"
	if len(os.Args) > 1 {
		outputPath = os.Args[1]
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// Generate the sample report
	fmt.Println("Generating LEONA & CRAVIT Sample CRA Compliance Report...")
	err := services.CreateSampleReport(outputPath)
	if err != nil {
		log.Fatalf("Failed to generate report: %v", err)
	}

	fmt.Printf("✅ Sample report generated successfully: %s\n", outputPath)
	fmt.Println("📄 This report demonstrates:")
	fmt.Println("   - Annex I Compliance Mapping")
	fmt.Println("   - 68% compliance score (realistic for non-compliant systems)")
	fmt.Println("   - Formal Liability Shield Statement")
	fmt.Println("   - Professional formatting for enterprise sales")
}
