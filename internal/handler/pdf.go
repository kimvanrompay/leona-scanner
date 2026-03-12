package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"leona-scanner/internal/usecase"

	"github.com/gorilla/mux"
)

type PDFHandler struct {
	scannerService *usecase.ScannerService
	pdfDir         string
}

func NewPDFHandler(scannerService *usecase.ScannerService, pdfDir string) *PDFHandler {
	// Create PDF directory if it doesn't exist
	if err := os.MkdirAll(pdfDir, 0755); err != nil {
		log.Printf("Warning: failed to create PDF directory: %v", err)
	}

	return &PDFHandler{
		scannerService: scannerService,
		pdfDir:         pdfDir,
	}
}

// HandleDownloadPDF serves the PDF report for a paid scan
func (h *PDFHandler) HandleDownloadPDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID := vars["scan_id"]

	if scanID == "" {
		http.Error(w, "Scan ID is required", http.StatusBadRequest)
		return
	}

	// Get scan from database
	result, scan, err := h.scannerService.GetScanResult(scanID)
	if err != nil {
		log.Printf("Failed to get scan: %v", err)
		http.Error(w, "Scan not found", http.StatusNotFound)
		return
	}

	// Check if scan is paid
	if scan.Status != "PAID" {
		http.Error(w, "This scan has not been paid for. Please complete payment first.", http.StatusPaymentRequired)
		return
	}

	// PDF file path
	pdfPath := filepath.Join(h.pdfDir, fmt.Sprintf("%s.pdf", scanID))

	// Generate PDF if it doesn't exist
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Printf("Generating PDF for scan %s...", scanID)

		if err := h.scannerService.GeneratePDFReport(scan, result, pdfPath); err != nil {
			log.Printf("Failed to generate PDF: %v", err)
			http.Error(w, "Failed to generate PDF report", http.StatusInternalServerError)
			return
		}

		log.Printf("PDF generated successfully: %s", pdfPath)
	}

	// Serve the PDF file
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"CRAVIT-Report-%s.pdf\"", scanID[:8]))

	http.ServeFile(w, r, pdfPath)
}

// HandleGeneratePDF - Alternative endpoint that generates PDF immediately after scan (for testing)
func (h *PDFHandler) HandleGeneratePDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scanID := vars["scan_id"]

	if scanID == "" {
		http.Error(w, "Scan ID is required", http.StatusBadRequest)
		return
	}

	// Get scan from database
	result, scan, err := h.scannerService.GetScanResult(scanID)
	if err != nil {
		log.Printf("Failed to get scan: %v", err)
		http.Error(w, "Scan not found", http.StatusNotFound)
		return
	}

	pdfPath := filepath.Join(h.pdfDir, fmt.Sprintf("%s.pdf", scanID))

	log.Printf("Generating PDF for scan %s...", scanID)
	if err := h.scannerService.GeneratePDFReport(scan, result, pdfPath); err != nil {
		log.Printf("Failed to generate PDF: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate PDF: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success with download link
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success": true, "download_url": "/api/pdf/download/%s"}`, scanID)
}
