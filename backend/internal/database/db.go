package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// InitDB initializes database connection
func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// Migrate runs all database migrations
func Migrate(db *sql.DB) error {
	migrations := []string{
		// Scans table
		`CREATE TABLE IF NOT EXISTS scans (
			id VARCHAR(36) PRIMARY KEY,
			cloud_provider VARCHAR(50) NOT NULL,
			regions TEXT[] NOT NULL,
			resource_types TEXT[] NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			drift_count INT DEFAULT 0,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP
		)`,

		// Drift items table
		`CREATE TABLE IF NOT EXISTS drift_items (
			id VARCHAR(36) PRIMARY KEY,
			scan_id VARCHAR(36) NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
			resource_id VARCHAR(255) NOT NULL,
			resource_type VARCHAR(100) NOT NULL,
			cloud_provider VARCHAR(50) NOT NULL,
			region VARCHAR(50),
			severity VARCHAR(50) NOT NULL,
			category VARCHAR(100) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			expected_state JSONB,
			actual_state JSONB,
			difference JSONB,
			remediation_status VARCHAR(50) DEFAULT 'pending',
			is_resolved BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// Create indexes for drift_items
		`CREATE INDEX IF NOT EXISTS idx_drift_items_scan_id ON drift_items(scan_id)`,
		`CREATE INDEX IF NOT EXISTS idx_drift_items_severity ON drift_items(severity)`,
		`CREATE INDEX IF NOT EXISTS idx_drift_items_status ON drift_items(remediation_status)`,
		`CREATE INDEX IF NOT EXISTS idx_drift_items_resource_id ON drift_items(resource_id)`,

		// Remediation requests table
		`CREATE TABLE IF NOT EXISTS remediation_requests (
			id VARCHAR(36) PRIMARY KEY,
			drift_id VARCHAR(36) NOT NULL REFERENCES drift_items(id) ON DELETE CASCADE,
			action_type VARCHAR(100) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			dry_run BOOLEAN DEFAULT FALSE,
			approval_status VARCHAR(50) DEFAULT 'pending',
			approved_by VARCHAR(255),
			approved_at TIMESTAMP,
			executed_at TIMESTAMP,
			result JSONB,
			error TEXT,
			rolled_back_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,

		// Create indexes for remediation_requests
		`CREATE INDEX IF NOT EXISTS idx_remediation_drift_id ON remediation_requests(drift_id)`,
		`CREATE INDEX IF NOT EXISTS idx_remediation_status ON remediation_requests(status)`,

		// Compliance rules table
		`CREATE TABLE IF NOT EXISTS compliance_rules (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			severity VARCHAR(50) NOT NULL,
			category VARCHAR(100) NOT NULL,
			resource_types TEXT[] NOT NULL,
			rules JSONB NOT NULL,
			enabled BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP
		)`,

		// Reports table
		`CREATE TABLE IF NOT EXISTS reports (
			id VARCHAR(36) PRIMARY KEY,
			scan_id VARCHAR(36) NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
			cloud_provider VARCHAR(50) NOT NULL,
			total_resources INT DEFAULT 0,
			drifted_resources INT DEFAULT 0,
			drift_percentage FLOAT DEFAULT 0,
			by_severity JSONB,
			by_category JSONB,
			cost_impact FLOAT DEFAULT 0,
			compliance_status VARCHAR(50),
			recommendations TEXT[],
			created_at TIMESTAMP NOT NULL
		)`,

		// Create indexes for reports
		`CREATE INDEX IF NOT EXISTS idx_reports_scan_id ON reports(scan_id)`,

		// Audit logs table
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			action VARCHAR(255) NOT NULL,
			resource_type VARCHAR(100) NOT NULL,
			resource_id VARCHAR(255),
			details JSONB,
			timestamp TIMESTAMP NOT NULL,
			ip_address VARCHAR(45)
		)`,

		// Create indexes for audit_logs
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
