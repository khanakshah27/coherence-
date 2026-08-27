package remediation

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"coherence.dev/backend/internal/models"
)

// Engine handles the execution of remediation actions
type Engine struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewEngine creates a new remediation engine
func NewEngine(db *sql.DB, logger *logrus.Logger) *Engine {
	return &Engine{db: db, logger: logger}
}

// Execute runs a remediation action for a given drift item
func (e *Engine) Execute(ctx context.Context, req *models.RemediationRequest, drift *models.DriftItem) error {
	e.logger.WithFields(logrus.Fields{
		"remediation_id": req.ID,
		"drift_id":       drift.ID,
		"action":         req.ActionType,
		"dry_run":        req.DryRun,
	}).Info("Starting remediation execution")

	// Update status to executing
	if err := e.updateStatus(req.ID, "executing", ""); err != nil {
		return err
	}

	var execErr error
	switch req.ActionType {
	case "apply_iac":
		execErr = e.applyIaC(ctx, drift, req.DryRun)
	case "safe_remediate":
		execErr = e.safeRemediate(ctx, drift, req.DryRun)
	case "update_code":
		execErr = e.updateCode(ctx, drift, req.DryRun)
	case "delete_resource":
		execErr = e.deleteResource(ctx, drift, req.DryRun)
	default:
		execErr = fmt.Errorf("unknown action type: %s", req.ActionType)
	}

	if execErr != nil {
		e.logger.WithError(execErr).Error("Remediation execution failed")
		if err := e.updateStatus(req.ID, "failed", execErr.Error()); err != nil {
			e.logger.WithError(err).Error("Failed to update remediation status")
		}
		return execErr
	}

	// Mark as success
	now := time.Now()
	query := `UPDATE remediation_requests SET status = $1, executed_at = $2, updated_at = $3 WHERE id = $4`
	if _, err := e.db.ExecContext(ctx, query, "success", now, now, req.ID); err != nil {
		return err
	}

	// Mark drift as resolved (unless dry run)
	if !req.DryRun {
		if err := e.markDriftResolved(ctx, drift.ID); err != nil {
			e.logger.WithError(err).Warn("Failed to mark drift as resolved after remediation")
		}
	}

	e.logger.WithField("remediation_id", req.ID).Info("Remediation completed successfully")
	return nil
}

// Rollback reverts a previously applied remediation
func (e *Engine) Rollback(ctx context.Context, req *models.RemediationRequest, drift *models.DriftItem) error {
	e.logger.WithField("remediation_id", req.ID).Info("Starting remediation rollback")

	// In a real implementation:
	// 1. Fetch the previous state snapshot taken before execution
	// 2. Re-apply the previous state via cloud API or Terraform
	// 3. Update the drift item status

	now := time.Now()
	query := `UPDATE remediation_requests SET rolled_back_at = $1, status = $2, updated_at = $3 WHERE id = $4`
	if _, err := e.db.ExecContext(ctx, query, now, "rolled_back", now, req.ID); err != nil {
		return err
	}

	// Re-open the drift item
	driftQuery := `UPDATE drift_items SET is_resolved = false, remediation_status = $1, updated_at = $2 WHERE id = $3`
	if _, err := e.db.ExecContext(ctx, driftQuery, "rolled_back", now, drift.ID); err != nil {
		return err
	}

	return nil
}

// DryRun previews what would happen during remediation
func (e *Engine) DryRun(ctx context.Context, actionType string, drift *models.DriftItem) (map[string]interface{}, error) {
	preview := map[string]interface{}{
		"action":      actionType,
		"resource_id": drift.ResourceID,
		"changes":     []map[string]interface{}{},
		"risk":        "low",
		"estimated_duration_seconds": 30,
	}

	switch actionType {
	case "apply_iac":
		preview["changes"] = []map[string]interface{}{
			{
				"attribute": "instance_type",
				"from":      drift.ActualState["instance_type"],
				"to":        drift.ExpectedState["instance_type"],
				"impact":    "Resource restart required",
			},
		}
		preview["risk"] = "medium"
	case "safe_remediate":
		preview["changes"] = []map[string]interface{}{
			{
				"attribute": "tags",
				"from":      drift.ActualState["tags"],
				"to":        drift.ExpectedState["tags"],
				"impact":    "No downtime",
			},
		}
		preview["risk"] = "low"
	}

	return preview, nil
}

// ─── Private action handlers ────────────────────────────────────────────────

// applyIaC re-applies the IaC definition to the drifted resource
func (e *Engine) applyIaC(ctx context.Context, drift *models.DriftItem, dryRun bool) error {
	e.logger.WithFields(logrus.Fields{
		"resource_id": drift.ResourceID,
		"dry_run":     dryRun,
	}).Info("Applying IaC to resource")

	// In production this would:
	// 1. Identify the Terraform module / CF stack owning the resource
	// 2. Run `terraform apply -target=<resource>` or `aws cloudformation update-stack`
	// 3. Monitor the apply until completion
	// 4. Verify the new state matches expected

	if dryRun {
		e.logger.Info("[DRY RUN] Would apply IaC changes to resource:", drift.ResourceID)
		return nil
	}

	// Simulate apply delay
	time.Sleep(2 * time.Second)
	e.logger.Info("IaC apply completed successfully for resource:", drift.ResourceID)
	return nil
}

// safeRemediate applies non-disruptive fixes (tags, minor config changes)
func (e *Engine) safeRemediate(ctx context.Context, drift *models.DriftItem, dryRun bool) error {
	e.logger.WithField("resource_id", drift.ResourceID).Info("Applying safe remediation")

	// Safe remediations include:
	// - Adding/updating tags
	// - Enabling versioning on S3
	// - Enabling encryption on S3
	// - Adding lifecycle rules

	if dryRun {
		e.logger.Info("[DRY RUN] Would apply safe remediation to:", drift.ResourceID)
		return nil
	}

	time.Sleep(500 * time.Millisecond)
	return nil
}

// updateCode creates a PR / commit to update IaC to match actual state
func (e *Engine) updateCode(ctx context.Context, drift *models.DriftItem, dryRun bool) error {
	e.logger.WithField("resource_id", drift.ResourceID).Info("Updating IaC code")

	// In production:
	// 1. Clone or checkout the IaC repository
	// 2. Find the resource definition
	// 3. Update the definition to match actual state
	// 4. Create a pull request via GitHub/GitLab API

	if dryRun {
		e.logger.Info("[DRY RUN] Would open PR to update IaC for:", drift.ResourceID)
		return nil
	}

	return nil
}

// deleteResource removes an unmanaged resource from the cloud
func (e *Engine) deleteResource(ctx context.Context, drift *models.DriftItem, dryRun bool) error {
	e.logger.WithField("resource_id", drift.ResourceID).Warn("Deleting unmanaged resource")

	if dryRun {
		e.logger.Info("[DRY RUN] Would delete resource:", drift.ResourceID)
		return nil
	}

	// In production: call the cloud provider API to delete the resource
	return nil
}

// ─── Database helpers ────────────────────────────────────────────────────────

func (e *Engine) updateStatus(id, status, errMsg string) error {
	query := `UPDATE remediation_requests SET status = $1, error = $2, updated_at = $3 WHERE id = $4`
	_, err := e.db.Exec(query, status, errMsg, time.Now(), id)
	return err
}

func (e *Engine) markDriftResolved(ctx context.Context, driftID string) error {
	query := `UPDATE drift_items SET is_resolved = true, remediation_status = $1, updated_at = $2 WHERE id = $3`
	_, err := e.db.ExecContext(ctx, query, "success", time.Now(), driftID)
	return err
}
