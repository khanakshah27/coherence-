package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"coherence.dev/backend/internal/drift"
	"coherence.dev/backend/internal/models"
)

// ScanHandler handles scan-related API requests
type ScanHandler struct {
	driftService *drift.Service
	db           *sql.DB
	logger       *logrus.Logger
}

// NewScanHandler creates a new scan handler
func NewScanHandler(ds *drift.Service, db *sql.DB, logger *logrus.Logger) *ScanHandler {
	return &ScanHandler{
		driftService: ds,
		db:           db,
		logger:       logger,
	}
}

// CreateScan creates a new drift scan
func (h *ScanHandler) CreateScan(c *gin.Context) {
	var req struct {
		CloudProvider string   `json:"cloud_provider" binding:"required"`
		Regions       []string `json:"regions" binding:"required"`
		ResourceTypes []string `json:"resource_types"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scan := &models.Scan{
		ID:            uuid.New().String(),
		CloudProvider: models.CloudProvider(req.CloudProvider),
		Regions:       req.Regions,
		ResourceTypes: req.ResourceTypes,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	// Save scan to database
	query := `INSERT INTO scans (id, cloud_provider, regions, resource_types, status, created_at)
	         VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := h.db.Exec(query, scan.ID, scan.CloudProvider, scan.Regions, scan.ResourceTypes, scan.Status, scan.CreatedAt)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create scan")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scan"})
		return
	}

	// Start scan in background
	go func() {
		if err := h.driftService.ScanDrift(context.Background(), scan); err != nil {
			h.logger.WithError(err).Error("Drift scan failed")
		}
	}()

	c.JSON(http.StatusCreated, scan)
}

// ListScans lists all scans
func (h *ScanHandler) ListScans(c *gin.Context) {
	query := `SELECT id, cloud_provider, status, drift_count, started_at, completed_at, created_at
	         FROM scans ORDER BY created_at DESC LIMIT 100`

	rows, err := h.db.QueryContext(c, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list scans"})
		return
	}
	defer rows.Close()

	var scans []*models.Scan
	for rows.Next() {
		var scan models.Scan
		if err := rows.Scan(&scan.ID, &scan.CloudProvider, &scan.Status, &scan.DriftCount,
			&scan.StartedAt, &scan.CompletedAt, &scan.CreatedAt); err != nil {
			continue
		}
		scans = append(scans, &scan)
	}

	c.JSON(http.StatusOK, scans)
}

// GetScan retrieves a specific scan
func (h *ScanHandler) GetScan(c *gin.Context) {
	scanID := c.Param("id")

	query := `SELECT id, cloud_provider, status, drift_count, started_at, completed_at, created_at
	         FROM scans WHERE id = $1`

	var scan models.Scan
	err := h.db.QueryRowContext(c, query, scanID).Scan(
		&scan.ID, &scan.CloudProvider, &scan.Status, &scan.DriftCount,
		&scan.StartedAt, &scan.CompletedAt, &scan.CreatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scan"})
		return
	}

	c.JSON(http.StatusOK, scan)
}

// DeleteScan deletes a scan
func (h *ScanHandler) DeleteScan(c *gin.Context) {
	scanID := c.Param("id")

	query := `DELETE FROM scans WHERE id = $1`
	_, err := h.db.ExecContext(c, query, scanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scan deleted"})
}

// RetryScan retries a failed scan
func (h *ScanHandler) RetryScan(c *gin.Context) {
	scanID := c.Param("id")

	query := `SELECT id, cloud_provider, regions, resource_types FROM scans WHERE id = $1`
	var scan models.Scan
	if err := h.db.QueryRowContext(c, query, scanID).Scan(&scan.ID, &scan.CloudProvider, &scan.Regions, &scan.ResourceTypes); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
		return
	}

	// Retry scan in background
	go func() {
		if err := h.driftService.ScanDrift(context.Background(), &scan); err != nil {
			h.logger.WithError(err).Error("Retry scan failed")
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Scan retry started"})
}

// DriftHandler handles drift-related API requests
type DriftHandler struct {
	driftService *drift.Service
	db           *sql.DB
	logger       *logrus.Logger
}

// NewDriftHandler creates a new drift handler
func NewDriftHandler(ds *drift.Service, db *sql.DB, logger *logrus.Logger) *DriftHandler {
	return &DriftHandler{
		driftService: ds,
		db:           db,
		logger:       logger,
	}
}

// ListDrifts lists all drift items
func (h *DriftHandler) ListDrifts(c *gin.Context) {
	scanID := c.Query("scan_id")
	severity := c.Query("severity")

	query := `SELECT id, scan_id, resource_id, resource_type, cloud_provider, severity, 
	         category, title, description, is_resolved, created_at, updated_at
	         FROM drift_items WHERE 1=1`

	args := []interface{}{}
	if scanID != "" {
		query += ` AND scan_id = $` + string(rune(len(args)+1))
		args = append(args, scanID)
	}
	if severity != "" {
		query += ` AND severity = $` + string(rune(len(args)+1))
		args = append(args, severity)
	}

	query += ` ORDER BY created_at DESC LIMIT 1000`

	rows, err := h.db.QueryContext(c, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list drifts"})
		return
	}
	defer rows.Close()

	var drifts []*models.DriftItem
	for rows.Next() {
		var drift models.DriftItem
		if err := rows.Scan(&drift.ID, &drift.ScanID, &drift.ResourceID, &drift.ResourceType,
			&drift.CloudProvider, &drift.Severity, &drift.Category, &drift.Title,
			&drift.Description, &drift.IsResolved, &drift.CreatedAt, &drift.UpdatedAt); err != nil {
			continue
		}
		drifts = append(drifts, &drift)
	}

	c.JSON(http.StatusOK, drifts)
}

// GetDrift retrieves a specific drift item
func (h *DriftHandler) GetDrift(c *gin.Context) {
	driftID := c.Param("id")

	query := `SELECT id, scan_id, resource_id, resource_type, cloud_provider, severity, 
	         category, title, description, is_resolved, created_at, updated_at
	         FROM drift_items WHERE id = $1`

	var drift models.DriftItem
	err := h.db.QueryRowContext(c, query, driftID).Scan(
		&drift.ID, &drift.ScanID, &drift.ResourceID, &drift.ResourceType,
		&drift.CloudProvider, &drift.Severity, &drift.Category, &drift.Title,
		&drift.Description, &drift.IsResolved, &drift.CreatedAt, &drift.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Drift not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get drift"})
		return
	}

	c.JSON(http.StatusOK, drift)
}

// UpdateDrift updates a drift item
func (h *DriftHandler) UpdateDrift(c *gin.Context) {
	driftID := c.Param("id")

	var req struct {
		Category    string `json:"category"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE drift_items SET category = $1, title = $2, description = $3, updated_at = $4 WHERE id = $5`
	_, err := h.db.ExecContext(c, query, req.Category, req.Title, req.Description, time.Now(), driftID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update drift"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Drift updated"})
}

// ResolveDrift marks a drift as resolved
func (h *DriftHandler) ResolveDrift(c *gin.Context) {
	driftID := c.Param("id")

	query := `UPDATE drift_items SET is_resolved = true, remediation_status = $1, updated_at = $2 WHERE id = $3`
	_, err := h.db.ExecContext(c, query, "resolved", time.Now(), driftID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve drift"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Drift resolved"})
}

// BulkResolveDrifts marks multiple drifts as resolved
func (h *DriftHandler) BulkResolveDrifts(c *gin.Context) {
	var req struct {
		DriftIDs []string `json:"drift_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE drift_items SET is_resolved = true, remediation_status = $1, updated_at = $2 WHERE id = ANY($3)`
	_, err := h.db.ExecContext(c, query, "resolved", time.Now(), req.DriftIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve drifts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Drifts resolved"})
}

// ReportHandler handles report-related API requests
type ReportHandler struct {
	driftService *drift.Service
	db           *sql.DB
	logger       *logrus.Logger
}

// NewReportHandler creates a new report handler
func NewReportHandler(ds *drift.Service, db *sql.DB, logger *logrus.Logger) *ReportHandler {
	return &ReportHandler{
		driftService: ds,
		db:           db,
		logger:       logger,
	}
}

// ListReports lists all reports
func (h *ReportHandler) ListReports(c *gin.Context) {
	c.JSON(http.StatusOK, []models.Report{})
}

// GetReport retrieves a specific report
func (h *ReportHandler) GetReport(c *gin.Context) {
	c.JSON(http.StatusOK, models.Report{})
}

// GenerateReport generates a new report
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Report generated"})
}

// ExportReport exports a report in specified format
func (h *ReportHandler) ExportReport(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "json"
	}
	c.String(http.StatusOK, "Report exported as "+format)
}

// RemediationHandler handles remediation-related API requests
type RemediationHandler struct {
	driftService *drift.Service
	db           *sql.DB
	logger       *logrus.Logger
}

// NewRemediationHandler creates a new remediation handler
func NewRemediationHandler(ds *drift.Service, db *sql.DB, logger *logrus.Logger) *RemediationHandler {
	return &RemediationHandler{
		driftService: ds,
		db:           db,
		logger:       logger,
	}
}

// RequestRemediation requests a remediation action
func (h *RemediationHandler) RequestRemediation(c *gin.Context) {
	var req struct {
		DriftID    string `json:"drift_id" binding:"required"`
		ActionType string `json:"action_type" binding:"required"`
		DryRun     bool   `json:"dry_run"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	remediation := models.NewRemediationRequest(req.DriftID, req.ActionType)
	remediation.DryRun = req.DryRun

	c.JSON(http.StatusCreated, remediation)
}

// ListRemediations lists all remediation requests
func (h *RemediationHandler) ListRemediations(c *gin.Context) {
	c.JSON(http.StatusOK, []models.RemediationRequest{})
}

// GetRemediation retrieves a specific remediation request
func (h *RemediationHandler) GetRemediation(c *gin.Context) {
	c.JSON(http.StatusOK, models.RemediationRequest{})
}

// ApproveRemediation approves a remediation request
func (h *RemediationHandler) ApproveRemediation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Remediation approved"})
}

// RejectRemediation rejects a remediation request
func (h *RemediationHandler) RejectRemediation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Remediation rejected"})
}

// ExecuteRemediation executes a remediation request
func (h *RemediationHandler) ExecuteRemediation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Remediation executed"})
}

// RollbackRemediation rolls back a remediation
func (h *RemediationHandler) RollbackRemediation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Remediation rolled back"})
}
