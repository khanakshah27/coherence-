package models

import (
	"time"

	"github.com/google/uuid"
)

// DriftSeverity represents the severity level of drift
type DriftSeverity string

const (
	SeverityCritical DriftSeverity = "critical"
	SeverityHigh     DriftSeverity = "high"
	SeverityMedium   DriftSeverity = "medium"
	SeverityLow      DriftSeverity = "low"
	SeverityInfo     DriftSeverity = "info"
)

// CloudProvider represents supported cloud providers
type CloudProvider string

const (
	ProviderAWS   CloudProvider = "aws"
	ProviderGCP   CloudProvider = "gcp"
	ProviderAzure CloudProvider = "azure"
)

// RemediationStatus tracks the state of a remediation action
type RemediationStatus string

const (
	RemediationPending   RemediationStatus = "pending"
	RemediationApproved  RemediationStatus = "approved"
	RemediationExecuting RemediationStatus = "executing"
	RemediationSuccess   RemediationStatus = "success"
	RemediationFailed    RemediationStatus = "failed"
	RemediationRolledBack RemediationStatus = "rolled_back"
)

// Scan represents an infrastructure scan
type Scan struct {
	ID            string        `json:"id" db:"id"`
	CloudProvider CloudProvider `json:"cloud_provider" db:"cloud_provider"`
	Regions       []string      `json:"regions" db:"regions"`
	ResourceTypes []string      `json:"resource_types" db:"resource_types"`
	Status        string        `json:"status" db:"status"` // pending, running, completed, failed
	DriftCount    int           `json:"drift_count" db:"drift_count"`
	StartedAt     time.Time     `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time    `json:"completed_at" db:"completed_at"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
}

// DriftItem represents a single piece of detected drift
type DriftItem struct {
	ID                string            `json:"id" db:"id"`
	ScanID            string            `json:"scan_id" db:"scan_id"`
	ResourceID        string            `json:"resource_id" db:"resource_id"`
	ResourceType      string            `json:"resource_type" db:"resource_type"`
	CloudProvider     CloudProvider     `json:"cloud_provider" db:"cloud_provider"`
	Region            string            `json:"region" db:"region"`
	Severity          DriftSeverity     `json:"severity" db:"severity"`
	Category          string            `json:"category" db:"category"` // breaking, compliance, performance, cost, etc
	Title             string            `json:"title" db:"title"`
	Description       string            `json:"description" db:"description"`
	ExpectedState     map[string]interface{} `json:"expected_state" db:"expected_state"`
	ActualState       map[string]interface{} `json:"actual_state" db:"actual_state"`
	Difference        map[string]interface{} `json:"difference" db:"difference"`
	ImpactAnalysis    *ImpactAnalysis   `json:"impact_analysis" db:"impact_analysis"`
	AuditInfo         *AuditInfo        `json:"audit_info" db:"audit_info"`
	FixSuggestions    []FixSuggestion   `json:"fix_suggestions" db:"fix_suggestions"`
	RemediationStatus RemediationStatus `json:"remediation_status" db:"remediation_status"`
	IsResolved        bool              `json:"is_resolved" db:"is_resolved"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
}

// ImpactAnalysis contains analysis of the drift's impact
type ImpactAnalysis struct {
	CostImpact        float64  `json:"cost_impact"`
	PerformanceImpact string   `json:"performance_impact"` // none, low, medium, high, critical
	SecurityImpact    string   `json:"security_impact"`
	ComplianceRules   []string `json:"compliance_rules"`
	AffectedServices  []string `json:"affected_services"`
}

// AuditInfo contains who made the change and when
type AuditInfo struct {
	ChangeType    string    `json:"change_type"` // create, update, delete
	ChangedBy     string    `json:"changed_by"`  // IAM principal ARN or user email
	ChangeTime    time.Time `json:"change_time"`
	Source        string    `json:"source"` // console, api, sdk, terraform
	EventID       string    `json:"event_id"` // CloudTrail event ID
	SourceIP      string    `json:"source_ip"`
	UserAgent     string    `json:"user_agent"`
}

// FixSuggestion represents an automatic fix suggestion
type FixSuggestion struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Action       string `json:"action"` // apply_iac, update_code, safe_remediate, manual_fix
	IsSafeAuto   bool   `json:"is_safe_auto"`
	Steps        []string `json:"steps"`
	Risk         string `json:"risk"` // none, low, medium, high
	TimeToFix    int    `json:"time_to_fix_seconds"`
	EstimatedCost float64 `json:"estimated_cost_savings"`
}

// RemediationRequest represents a request to remediate drift
type RemediationRequest struct {
	ID              string            `json:"id" db:"id"`
	DriftID         string            `json:"drift_id" db:"drift_id"`
	ActionType      string            `json:"action_type" db:"action_type"` // apply_iac, safe_remediate, etc
	Status          RemediationStatus `json:"status" db:"status"`
	DryRun          bool              `json:"dry_run" db:"dry_run"`
	ApprovalStatus  string            `json:"approval_status" db:"approval_status"` // pending, approved, rejected
	ApprovedBy      string            `json:"approved_by" db:"approved_by"`
	ApprovedAt      *time.Time        `json:"approved_at" db:"approved_at"`
	ExecutedAt      *time.Time        `json:"executed_at" db:"executed_at"`
	Result          map[string]interface{} `json:"result" db:"result"`
	Error           string            `json:"error" db:"error"`
	RolledBackAt    *time.Time        `json:"rolled_back_at" db:"rolled_back_at"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

// ComplianceRule represents a compliance rule to check
type ComplianceRule struct {
	ID          string   `json:"id" db:"id"`
	Name        string   `json:"name" db:"name"`
	Description string   `json:"description" db:"description"`
	Severity    DriftSeverity `json:"severity" db:"severity"`
	Category    string   `json:"category" db:"category"` // security, cost, performance, compliance
	ResourceTypes []string `json:"resource_types" db:"resource_types"`
	Rules       map[string]interface{} `json:"rules" db:"rules"`
	Enabled     bool     `json:"enabled" db:"enabled"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Report represents a drift report
type Report struct {
	ID              string        `json:"id" db:"id"`
	ScanID          string        `json:"scan_id" db:"scan_id"`
	CloudProvider   CloudProvider `json:"cloud_provider" db:"cloud_provider"`
	TotalResources  int           `json:"total_resources" db:"total_resources"`
	DriftedResources int          `json:"drifted_resources" db:"drifted_resources"`
	DriftPercentage float64       `json:"drift_percentage" db:"drift_percentage"`
	BySeverity      map[string]int `json:"by_severity" db:"by_severity"`
	ByCategory      map[string]int `json:"by_category" db:"by_category"`
	CostImpact      float64       `json:"cost_impact" db:"cost_impact"`
	ComplianceStatus string       `json:"compliance_status" db:"compliance_status"`
	Recommendations []string      `json:"recommendations" db:"recommendations"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
}

// NewScan creates a new scan
func NewScan(provider CloudProvider, regions, resourceTypes []string) *Scan {
	return &Scan{
		ID:            uuid.New().String(),
		CloudProvider: provider,
		Regions:       regions,
		ResourceTypes: resourceTypes,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}
}

// NewDriftItem creates a new drift item
func NewDriftItem(scanID, resourceID, resourceType string, provider CloudProvider, severity DriftSeverity) *DriftItem {
	return &DriftItem{
		ID:                uuid.New().String(),
		ScanID:            scanID,
		ResourceID:        resourceID,
		ResourceType:      resourceType,
		CloudProvider:     provider,
		Severity:          severity,
		RemediationStatus: RemediationPending,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// NewRemediationRequest creates a new remediation request
func NewRemediationRequest(driftID, actionType string) *RemediationRequest {
	return &RemediationRequest{
		ID:             uuid.New().String(),
		DriftID:        driftID,
		ActionType:     actionType,
		Status:         RemediationPending,
		ApprovalStatus: "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}
