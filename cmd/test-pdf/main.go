package main

import (
	"fmt"
	"log"
	"leona-scanner/internal/services"
)

func main() {
	outputPath := "./test-report.pdf"
	
	fmt.Println("🔧 Generating sample CRAVIT PDF report...")
	
	err := services.CreateSampleReport(outputPath)
	if err != nil {
		log.Fatalf("❌ Failed to generate PDF: %v", err)
	}
	
	fmt.Printf("✅ SUCCESS! PDF generated at: %s\n", outputPath)
	fmt.Println("📄 Open it to verify the output looks professional")
}
