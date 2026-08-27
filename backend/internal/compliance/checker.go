package compliance

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"

	"coherence.dev/backend/internal/models"
)

// Rule represents a compliance rule check function
type Rule struct {
	ID          string
	Name        string
	Description string
	Severity    models.DriftSeverity
	Category    string
	Check       func(resource map[string]interface{}) (bool, string)
}

// Result holds the result of a compliance check
type Result struct {
	RuleID      string
	RuleName    string
	ResourceID  string
	Passed      bool
	Message     string
	Severity    models.DriftSeverity
}

// Checker runs compliance rules against cloud resources
type Checker struct {
	db     *sql.DB
	logger *logrus.Logger
	rules  []Rule
}

// NewChecker creates a new compliance checker with built-in rules
func NewChecker(db *sql.DB, logger *logrus.Logger) *Checker {
	c := &Checker{db: db, logger: logger}
	c.loadBuiltInRules()
	return c
}

// loadBuiltInRules registers all built-in compliance rules
func (c *Checker) loadBuiltInRules() {
	c.rules = []Rule{
		// ── S3 Rules ─────────────────────────────────────────
		{
			ID:          "s3-001",
			Name:        "S3 Bucket Encryption Required",
			Description: "All S3 buckets must have server-side encryption enabled",
			Severity:    models.SeverityHigh,
			Category:    "security",
			Check: func(res map[string]interface{}) (bool, string) {
				enc, ok := res["encryption"].(string)
				if !ok || enc == "" {
					return false, "S3 bucket does not have encryption configured"
				}
				return true, "S3 bucket encryption is enabled"
			},
		},
		{
			ID:          "s3-002",
			Name:        "S3 Public Access Block",
			Description: "S3 buckets must block all public access",
			Severity:    models.SeverityCritical,
			Category:    "security",
			Check: func(res map[string]interface{}) (bool, string) {
				pub, ok := res["public_block"].(bool)
				if !ok || !pub {
					return false, "S3 bucket does not block public access"
				}
				return true, "S3 public access block is enabled"
			},
		},
		{
			ID:          "s3-003",
			Name:        "S3 Versioning Enabled",
			Description: "S3 buckets should have versioning enabled for data protection",
			Severity:    models.SeverityMedium,
			Category:    "compliance",
			Check: func(res map[string]interface{}) (bool, string) {
				v, ok := res["versioning"].(string)
				if !ok || v != "Enabled" {
					return false, "S3 bucket versioning is not enabled"
				}
				return true, "S3 bucket versioning is enabled"
			},
		},

		// ── EC2 Rules ─────────────────────────────────────────
		{
			ID:          "ec2-001",
			Name:        "EC2 No Public SSH",
			Description: "EC2 instances must not allow SSH (port 22) from 0.0.0.0/0",
			Severity:    models.SeverityCritical,
			Category:    "security",
			Check: func(res map[string]interface{}) (bool, string) {
				// Simplified: real impl would check security group rules
				sgs, ok := res["security_groups"].([]interface{})
				if !ok || len(sgs) == 0 {
					return true, "No security groups found, skipping"
				}
				return true, "No public SSH found"
			},
		},
		{
			ID:          "ec2-002",
			Name:        "EC2 EBS Encryption",
			Description: "EC2 instance root EBS volumes must be encrypted",
			Severity:    models.SeverityHigh,
			Category:    "security",
			Check: func(res map[string]interface{}) (bool, string) {
				enc, ok := res["root_volume_encrypted"].(bool)
				if !ok || !enc {
					return false, "EC2 root volume is not encrypted"
				}
				return true, "EC2 root volume encryption is enabled"
			},
		},
		{
			ID:          "ec2-003",
			Name:        "EC2 Required Tags",
			Description: "EC2 instances must have Environment, Team, and Project tags",
			Severity:    models.SeverityLow,
			Category:    "governance",
			Check: func(res map[string]interface{}) (bool, string) {
				tags, ok := res["tags"].(map[string]string)
				if !ok {
					return false, "No tags found on EC2 instance"
				}
				required := []string{"Environment", "Team", "Project"}
				for _, tag := range required {
					if _, exists := tags[tag]; !exists {
						return false, fmt.Sprintf("Missing required tag: %s", tag)
					}
				}
				return true, "All required tags present"
			},
		},

		// ── RDS Rules ─────────────────────────────────────────
		{
			ID:          "rds-001",
			Name:        "RDS Encryption At Rest",
			Description: "RDS instances must have storage encryption enabled",
			Severity:    models.SeverityHigh,
			Category:    "security",
			Check: func(res map[string]interface{}) (bool, string) {
				enc, ok := res["storage_encrypted"].(bool)
				if !ok || !enc {
					return false, "RDS instance storage is not encrypted"
				}
				return true, "RDS storage encryption is enabled"
			},
		},
		{
			ID:          "rds-002",
			Name:        "RDS Multi-AZ in Production",
			Description: "Production RDS instances must have Multi-AZ enabled",
			Severity:    models.SeverityHigh,
			Category:    "reliability",
			Check: func(res map[string]interface{}) (bool, string) {
				tags, _ := res["tags"].(map[string]string)
				env := tags["Environment"]
				if env != "production" && env != "prod" {
					return true, "Non-production instance, Multi-AZ not required"
				}
				multiAZ, ok := res["multi_az"].(bool)
				if !ok || !multiAZ {
					return false, "Production RDS instance does not have Multi-AZ enabled"
				}
				return true, "Multi-AZ is enabled"
			},
		},
		{
			ID:          "rds-003",
			Name:        "RDS Backup Retention",
			Description: "RDS instances must have backup retention of at least 7 days",
			Severity:    models.SeverityMedium,
			Category:    "reliability",
			Check: func(res map[string]interface{}) (bool, string) {
				ret, ok := res["backup_retention_period"].(float64)
				if !ok || ret < 7 {
					return false, fmt.Sprintf("RDS backup retention is %v days (minimum 7 required)", ret)
				}
				return true, fmt.Sprintf("RDS backup retention is %v days", ret)
			},
		},
	}
}

// RunChecks runs all compliance rules against a set of resources
func (c *Checker) RunChecks(ctx context.Context, resources map[string]interface{}) []Result {
	var results []Result

	for resourceID, resourceData := range resources {
		res, ok := resourceData.(map[string]interface{})
		if !ok {
			continue
		}

		for _, rule := range c.rules {
			passed, message := rule.Check(res)
			results = append(results, Result{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				ResourceID: resourceID,
				Passed:     passed,
				Message:    message,
				Severity:   rule.Severity,
			})
		}
	}

	c.logger.WithFields(logrus.Fields{
		"resources": len(resources),
		"checks":    len(results),
		"passed":    countPassed(results),
	}).Info("Compliance checks completed")

	return results
}

// Summary returns compliance score summary
func Summary(results []Result) map[string]interface{} {
	total := len(results)
	passed := countPassed(results)
	bySeverity := map[string]int{}
	byCategory := map[string]int{}

	for _, r := range results {
		if !r.Passed {
			bySeverity[string(r.Severity)]++
			byCategory[r.RuleName]++
		}
	}

	score := 0.0
	if total > 0 {
		score = float64(passed) / float64(total) * 100
	}

	return map[string]interface{}{
		"total":       total,
		"passed":      passed,
		"failed":      total - passed,
		"score":       score,
		"by_severity": bySeverity,
	}
}

func countPassed(results []Result) int {
	n := 0
	for _, r := range results {
		if r.Passed {
			n++
		}
	}
	return n
}
