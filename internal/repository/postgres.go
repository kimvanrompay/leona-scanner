package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Lead struct {
	ID        int64
	Email     string
	CreatedAt time.Time
}

type Scan struct {
	ID         string
	LeadID     int64
	Platform   string
	Score      int
	Status     string // FREE or PAID
	RawData    []byte
	ResultJSON string
	CreatedAt  time.Time
	PaidAt     *time.Time
}

type Repository struct {
	db     *sql.DB
	dbType string // "sqlite" or "postgres"
}

func NewRepository(connStr string) (*Repository, error) {
	var db *sql.DB
	var err error
	var dbType string

	// Determine database type from connection string
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		dbType = "postgres"
		db, err = sql.Open("postgres", connStr)
	} else {
		// Default to SQLite for file paths or sqlite:// URLs
		dbType = "sqlite"
		if strings.HasPrefix(connStr, "sqlite://") {
			connStr = strings.TrimPrefix(connStr, "sqlite://")
		}
		db, err = sql.Open("sqlite3", connStr)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &Repository{db: db, dbType: dbType}

	// Initialize schema for SQLite
	if dbType == "sqlite" {
		if err := repo.initSQLiteSchema(); err != nil {
			return nil, fmt.Errorf("failed to initialize SQLite schema: %w", err)
		}
	}

	return repo, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

// CreateLead inserts a new lead or returns existing one by email
func (r *Repository) CreateLead(email string) (*Lead, error) {
	// First check if lead exists
	var lead Lead
	err := r.db.QueryRow(`
		SELECT id, email, created_at 
		FROM leads 
		WHERE email = $1
	`, email).Scan(&lead.ID, &lead.Email, &lead.CreatedAt)

	if err == nil {
		return &lead, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing lead: %w", err)
	}

	// Create new lead
	err = r.db.QueryRow(`
		INSERT INTO leads (email, created_at)
		VALUES ($1, $2)
		RETURNING id, email, created_at
	`, email, time.Now()).Scan(&lead.ID, &lead.Email, &lead.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create lead: %w", err)
	}

	return &lead, nil
}

// CreateScan inserts a new scan record
func (r *Repository) CreateScan(leadID int64, platform string, score int, rawData []byte, resultJSON string) (*Scan, error) {
	scan := &Scan{
		ID:         uuid.New().String(),
		LeadID:     leadID,
		Platform:   platform,
		Score:      score,
		Status:     "FREE",
		RawData:    rawData,
		ResultJSON: resultJSON,
		CreatedAt:  time.Now(),
	}

	_, err := r.db.Exec(`
		INSERT INTO scans (id, lead_id, platform, score, status, raw_data, result_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, scan.ID, scan.LeadID, scan.Platform, scan.Score, scan.Status, scan.RawData, scan.ResultJSON, scan.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create scan: %w", err)
	}

	return scan, nil
}

// GetScanByID retrieves a scan by its ID
func (r *Repository) GetScanByID(scanID string) (*Scan, error) {
	var scan Scan
	err := r.db.QueryRow(`
		SELECT id, lead_id, platform, score, status, raw_data, result_json, created_at, paid_at
		FROM scans
		WHERE id = $1
	`, scanID).Scan(&scan.ID, &scan.LeadID, &scan.Platform, &scan.Score, &scan.Status, &scan.RawData, &scan.ResultJSON, &scan.CreatedAt, &scan.PaidAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("scan not found: %s", scanID)
		}
		return nil, fmt.Errorf("failed to get scan: %w", err)
	}

	return &scan, nil
}

// MarkScanAsPaid updates scan status to PAID
func (r *Repository) MarkScanAsPaid(scanID string) error {
	now := time.Now()
	result, err := r.db.Exec(`
		UPDATE scans
		SET status = 'PAID', paid_at = $1
		WHERE id = $2
	`, now, scanID)

	if err != nil {
		return fmt.Errorf("failed to mark scan as paid: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("scan not found: %s", scanID)
	}

	return nil
}

// GetLeadByEmail retrieves a lead by email
func (r *Repository) GetLeadByEmail(email string) (*Lead, error) {
	var lead Lead
	err := r.db.QueryRow(`
		SELECT id, email, created_at
		FROM leads
		WHERE email = $1
	`, email).Scan(&lead.ID, &lead.Email, &lead.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("lead not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}

	return &lead, nil
}

// GetScansByLeadID retrieves all scans for a lead
func (r *Repository) GetScansByLeadID(leadID int64) ([]*Scan, error) {
	rows, err := r.db.Query(`
		SELECT id, lead_id, platform, score, status, raw_data, result_json, created_at, paid_at
		FROM scans
		WHERE lead_id = $1
		ORDER BY created_at DESC
	`, leadID)

	if err != nil {
		return nil, fmt.Errorf("failed to get scans: %w", err)
	}
	defer rows.Close()

	var scans []*Scan
	for rows.Next() {
		var scan Scan
		err := rows.Scan(&scan.ID, &scan.LeadID, &scan.Platform, &scan.Score, &scan.Status, &scan.RawData, &scan.ResultJSON, &scan.CreatedAt, &scan.PaidAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		scans = append(scans, &scan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return scans, nil
}

// initSQLiteSchema creates the necessary tables for SQLite
func (r *Repository) initSQLiteSchema() error {
	// Create leads table
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS leads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create leads table: %w", err)
	}

	// Create scans table
	_, err = r.db.Exec(`
		CREATE TABLE IF NOT EXISTS scans (
			id TEXT PRIMARY KEY,
			lead_id INTEGER NOT NULL,
			platform TEXT NOT NULL,
			score INTEGER NOT NULL,
			status TEXT NOT NULL,
			raw_data BLOB,
			result_json TEXT,
			created_at TIMESTAMP NOT NULL,
			paid_at TIMESTAMP,
			FOREIGN KEY (lead_id) REFERENCES leads(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create scans table: %w", err)
	}

	return nil
}
