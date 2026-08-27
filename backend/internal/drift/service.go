package drift

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"coherence.dev/backend/internal/models"
	"coherence.dev/backend/internal/providers"
)

// Service handles drift detection and analysis
type Service struct {
	db            *sql.DB
	awsProvider   providers.CloudProvider
	gcpProvider   providers.CloudProvider
	azureProvider providers.CloudProvider
	logger        *logrus.Logger
}

// NewService creates a new drift service
func NewService(
	db *sql.DB,
	awsProvider providers.CloudProvider,
	gcpProvider providers.CloudProvider,
	azureProvider providers.CloudProvider,
	logger *logrus.Logger,
) *Service {
	return &Service{
		db:            db,
		awsProvider:   awsProvider,
		gcpProvider:   gcpProvider,
		azureProvider: azureProvider,
		logger:        logger,
	}
}

// ScanDrift performs a comprehensive drift scan
func (s *Service) ScanDrift(ctx context.Context, scan *models.Scan) error {
	s.logger.WithField("scan_id", scan.ID).Info("Starting drift scan")

	// Update scan status to running
	if err := s.updateScanStatus(scan.ID, "running"); err != nil {
		return err
	}

	// Get the appropriate cloud provider
	provider := s.getProvider(scan.CloudProvider)
	if provider == nil {
		return fmt.Errorf("unsupported cloud provider: %s", scan.CloudProvider)
	}

	// Fetch actual cloud state
	actualResources, err := provider.GetResources(ctx, scan.Regions, scan.ResourceTypes)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch cloud resources")
		s.updateScanStatus(scan.ID, "failed")
		return err
	}

	// Fetch expected state from IaC
	expectedResources, err := s.getExpectedState(scan.CloudProvider)
	if err != nil {
		s.logger.WithError(err).Error("Failed to fetch expected state")
		s.updateScanStatus(scan.ID, "failed")
		return err
	}

	// Detect drift
	driftItems := s.detectDrift(actualResources, expectedResources, scan)

	// Enrich drift items with audit information
	for _, driftItem := range driftItems {
		if err := s.enrichDriftWithAudit(ctx, driftItem, provider); err != nil {
			s.logger.WithError(err).Warn("Failed to enrich drift with audit info")
		}
	}

	// Save drift items to database
	for _, driftItem := range driftItems {
		if err := s.saveDriftItem(driftItem); err != nil {
			s.logger.WithError(err).Error("Failed to save drift item")
		}
	}

	// Update scan status
	scan.DriftCount = len(driftItems)
	scan.CompletedAt = new(time.Time)
	*scan.CompletedAt = time.Now()
	if err := s.updateScanStatus(scan.ID, "completed"); err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"scan_id":     scan.ID,
		"drift_count": len(driftItems),
	}).Info("Drift scan completed")

	return nil
}

// detectDrift compares actual and expected state to find drift
func (s *Service) detectDrift(
	actualResources map[string]interface{},
	expectedResources map[string]interface{},
	scan *models.Scan,
) []*models.DriftItem {
	var driftItems []*models.DriftItem

	// Check for resources in actual but not in expected (extra resources)
	for resourceID, actualState := range actualResources {
		if expectedState, exists := expectedResources[resourceID]; !exists {
			// Resource exists in cloud but not in IaC
			severity := models.SeverityMedium
			if s.isCriticalResource(resourceID) {
				severity = models.SeverityHigh
			}

			driftItem := models.NewDriftItem(
				scan.ID,
				resourceID,
				s.extractResourceType(resourceID),
				scan.CloudProvider,
				severity,
			)
			driftItem.Title = fmt.Sprintf("Unmanaged resource: %s", resourceID)
			driftItem.Description = "This resource exists in your cloud account but is not defined in your Infrastructure-as-Code"
			driftItem.Category = "unmanaged"
			driftItem.ActualState = actualState.(map[string]interface{})
			driftItem.ExpectedState = nil
			driftItem.FixSuggestions = []models.FixSuggestion{
				{
					ID:         "suggest_1",
					Title:      "Add to IaC",
					Description: "Import or define this resource in your Terraform/CloudFormation",
					Action:     "update_code",
					Risk:       "low",
				},
				{
					ID:         "suggest_2",
					Title:      "Delete resource",
					Description: "Remove this unmanaged resource from your cloud account",
					Action:     "safe_remediate",
					Risk:       "medium",
					IsSafeAuto: false,
				},
			}

			driftItems = append(driftItems, driftItem)
		} else if !s.statesEqual(actualState, expectedState) {
			// Resource exists but configuration differs
			driftItem := models.NewDriftItem(
				scan.ID,
				resourceID,
				s.extractResourceType(resourceID),
				scan.CloudProvider,
				models.SeverityMedium,
			)
			driftItem.Title = fmt.Sprintf("Configuration drift: %s", resourceID)
			driftItem.Description = "Resource configuration differs from Infrastructure-as-Code"
			driftItem.Category = "configuration"
			driftItem.ActualState = actualState.(map[string]interface{})
			driftItem.ExpectedState = expectedState.(map[string]interface{})
			driftItem.Difference = s.calculateDifference(actualState, expectedState)

			// Suggest fixes
			driftItem.FixSuggestions = []models.FixSuggestion{
				{
					ID:         "suggest_1",
					Title:      "Apply IaC changes",
					Description: "Update resources to match Infrastructure-as-Code",
					Action:     "apply_iac",
					Risk:       "low",
					IsSafeAuto: true,
				},
				{
					ID:         "suggest_2",
					Title:      "Update IaC",
					Description: "Update your code to match current cloud configuration",
					Action:     "update_code",
					Risk:       "none",
				},
			}

			driftItems = append(driftItems, driftItem)
		}
	}

	// Check for resources in expected but not in actual (missing resources)
	for resourceID := range expectedResources {
		if _, exists := actualResources[resourceID]; !exists {
			driftItem := models.NewDriftItem(
				scan.ID,
				resourceID,
				s.extractResourceType(resourceID),
				scan.CloudProvider,
				models.SeverityCritical,
			)
			driftItem.Title = fmt.Sprintf("Missing resource: %s", resourceID)
			driftItem.Description = "This resource is defined in your Infrastructure-as-Code but does not exist in your cloud account"
			driftItem.Category = "missing"
			driftItem.FixSuggestions = []models.FixSuggestion{
				{
					ID:         "suggest_1",
					Title:      "Create resource",
					Description: "Apply IaC to create this missing resource",
					Action:     "apply_iac",
					Risk:       "low",
					IsSafeAuto: true,
				},
			}

			driftItems = append(driftItems, driftItem)
		}
	}

	return driftItems
}

// enrichDriftWithAudit adds audit information to drift items
func (s *Service) enrichDriftWithAudit(ctx context.Context, driftItem *models.DriftItem, provider providers.CloudProvider) error {
	// Fetch audit trail from cloud provider
	auditEvents, err := provider.GetAuditTrail(ctx, driftItem.ResourceID)
	if err != nil {
		return err
	}

	if len(auditEvents) > 0 {
		// Use the most recent event
		lastEvent := auditEvents[0]
		driftItem.AuditInfo = &models.AuditInfo{
			ChangeType: lastEvent.ChangeType,
			ChangedBy:  lastEvent.Principal,
			ChangeTime: lastEvent.Timestamp,
			Source:     lastEvent.Source,
			EventID:    lastEvent.EventID,
			SourceIP:   lastEvent.SourceIP,
			UserAgent:  lastEvent.UserAgent,
		}
	}

	return nil
}

// getExpectedState fetches the expected infrastructure state from IaC
func (s *Service) getExpectedState(provider models.CloudProvider) (map[string]interface{}, error) {
	// In a real implementation, this would:
	// 1. Parse Terraform/CloudFormation files
	// 2. Execute terraform show or cloudformation describe-stacks
	// 3. Compare with expected state

	// Mock implementation for now
	return map[string]interface{}{
		"i-1234567890abcdef0": map[string]interface{}{
			"name":            "web-server-01",
			"instance_type":   "t3.medium",
			"state":           "running",
			"availability_zone": "us-east-1a",
		},
		"s3://my-app-bucket": map[string]interface{}{
			"versioning": true,
			"encryption": "AES256",
			"public_access_block": true,
		},
	}, nil
}

// Helper methods

func (s *Service) getProvider(cloudProvider models.CloudProvider) providers.CloudProvider {
	switch cloudProvider {
	case models.ProviderAWS:
		return s.awsProvider
	case models.ProviderGCP:
		return s.gcpProvider
	case models.ProviderAzure:
		return s.azureProvider
	default:
		return nil
	}
}

func (s *Service) statesEqual(actual, expected interface{}) bool {
	// Simplified comparison - in production, use proper deep comparison
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)
	return actualStr == expectedStr
}

func (s *Service) calculateDifference(actual, expected interface{}) map[string]interface{} {
	// Calculate the difference between actual and expected state
	return map[string]interface{}{
		"actual":   actual,
		"expected": expected,
	}
}

func (s *Service) extractResourceType(resourceID string) string {
	// Extract resource type from resource ID (e.g., "i-" prefix for EC2)
	if len(resourceID) > 2 {
		switch resourceID[:2] {
		case "i-":
			return "ec2"
		case "s-":
			return "s3"
		case "db-":
			return "rds"
		default:
			return "unknown"
		}
	}
	return "unknown"
}

func (s *Service) isCriticalResource(resourceID string) bool {
	// Determine if a resource is critical based on naming patterns or tags
	criticalPatterns := []string{"prod", "production", "database", "primary"}
	for _, pattern := range criticalPatterns {
		if len(resourceID) > len(pattern) && resourceID[:len(pattern)] == pattern {
			return true
		}
	}
	return false
}

func (s *Service) updateScanStatus(scanID, status string) error {
	query := `UPDATE scans SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := s.db.Exec(query, status, time.Now(), scanID)
	return err
}

func (s *Service) saveDriftItem(driftItem *models.DriftItem) error {
	query := `
	INSERT INTO drift_items (
		id, scan_id, resource_id, resource_type, cloud_provider, region,
		severity, category, title, description, expected_state, actual_state,
		remediation_status, is_resolved, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	ON CONFLICT (id) DO UPDATE SET
		updated_at = $16
	`
	_, err := s.db.Exec(
		query,
		driftItem.ID, driftItem.ScanID, driftItem.ResourceID, driftItem.ResourceType,
		driftItem.CloudProvider, driftItem.Region, driftItem.Severity, driftItem.Category,
		driftItem.Title, driftItem.Description, driftItem.ExpectedState, driftItem.ActualState,
		driftItem.RemediationStatus, driftItem.IsResolved, driftItem.CreatedAt, driftItem.UpdatedAt,
	)
	return err
}

// GetDriftItems retrieves drift items from database
func (s *Service) GetDriftItems(scanID string) ([]*models.DriftItem, error) {
	query := `SELECT id, scan_id, resource_id, resource_type, cloud_provider, region, 
	         severity, category, title, description, is_resolved, created_at, updated_at 
	         FROM drift_items WHERE scan_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(context.Background(), query, scanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var driftItems []*models.DriftItem
	for rows.Next() {
		var item models.DriftItem
		if err := rows.Scan(
			&item.ID, &item.ScanID, &item.ResourceID, &item.ResourceType,
			&item.CloudProvider, &item.Region, &item.Severity, &item.Category,
			&item.Title, &item.Description, &item.IsResolved, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		driftItems = append(driftItems, &item)
	}

	return driftItems, rows.Err()
}
