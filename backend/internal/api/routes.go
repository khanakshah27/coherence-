package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"coherence.dev/backend/internal/drift"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	router *gin.Engine,
	db *sql.DB,
	driftService *drift.Service,
	logger *logrus.Logger,
) {
	// Initialize handlers
	scanHandler := NewScanHandler(driftService, db, logger)
	driftHandler := NewDriftHandler(driftService, db, logger)
	reportHandler := NewReportHandler(driftService, db, logger)
	remediationHandler := NewRemediationHandler(driftService, db, logger)

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Scan endpoints
	v1.POST("/scans", scanHandler.CreateScan)
	v1.GET("/scans", scanHandler.ListScans)
	v1.GET("/scans/:id", scanHandler.GetScan)
	v1.DELETE("/scans/:id", scanHandler.DeleteScan)
	v1.POST("/scans/:id/retry", scanHandler.RetryScan)

	// Drift endpoints
	v1.GET("/drifts", driftHandler.ListDrifts)
	v1.GET("/drifts/:id", driftHandler.GetDrift)
	v1.PUT("/drifts/:id", driftHandler.UpdateDrift)
	v1.POST("/drifts/:id/resolve", driftHandler.ResolveDrift)
	v1.POST("/drifts/bulk-resolve", driftHandler.BulkResolveDrifts)

	// Report endpoints
	v1.GET("/reports", reportHandler.ListReports)
	v1.GET("/reports/:id", reportHandler.GetReport)
	v1.POST("/reports/generate", reportHandler.GenerateReport)
	v1.GET("/reports/:id/export", reportHandler.ExportReport)

	// Remediation endpoints
	v1.POST("/remediations", remediationHandler.RequestRemediation)
	v1.GET("/remediations", remediationHandler.ListRemediations)
	v1.GET("/remediations/:id", remediationHandler.GetRemediation)
	v1.POST("/remediations/:id/approve", remediationHandler.ApproveRemediation)
	v1.POST("/remediations/:id/reject", remediationHandler.RejectRemediation)
	v1.POST("/remediations/:id/execute", remediationHandler.ExecuteRemediation)
	v1.POST("/remediations/:id/rollback", remediationHandler.RollbackRemediation)

	// Health and metrics
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	v1.GET("/metrics", func(c *gin.Context) {
		// Prometheus metrics endpoint
		c.String(200, "# Coherence Metrics\n")
	})
}
