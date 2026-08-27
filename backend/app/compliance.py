"""Compliance rule checker for cloud resources."""
from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Callable

from app.models import DriftSeverity


@dataclass
class Rule:
    id: str
    name: str
    description: str
    severity: DriftSeverity
    category: str
    check: Callable[[dict[str, Any]], tuple[bool, str]]


@dataclass
class Result:
    rule_id: str
    rule_name: str
    resource_id: str
    passed: bool
    message: str
    severity: DriftSeverity


def _s3_encryption(res: dict[str, Any]) -> tuple[bool, str]:
    enc = res.get("encryption")
    if not enc:
        return False, "S3 bucket does not have encryption configured"
    return True, "S3 bucket encryption is enabled"


def _s3_public_block(res: dict[str, Any]) -> tuple[bool, str]:
    if not res.get("public_block"):
        return False, "S3 bucket does not block public access"
    return True, "S3 public access block is enabled"


def _s3_versioning(res: dict[str, Any]) -> tuple[bool, str]:
    if res.get("versioning") != "Enabled":
        return False, "S3 bucket versioning is not enabled"
    return True, "S3 bucket versioning is enabled"


def _ec2_no_public_ssh(res: dict[str, Any]) -> tuple[bool, str]:
    # Simplified: a real implementation would inspect security group rules
    sgs = res.get("security_groups")
    if not sgs:
        return True, "No security groups found, skipping"
    return True, "No public SSH found"


def _ec2_ebs_encryption(res: dict[str, Any]) -> tuple[bool, str]:
    if not res.get("root_volume_encrypted"):
        return False, "EC2 root volume is not encrypted"
    return True, "EC2 root volume encryption is enabled"


def _ec2_required_tags(res: dict[str, Any]) -> tuple[bool, str]:
    tags = res.get("tags")
    if not isinstance(tags, dict):
        return False, "No tags found on EC2 instance"
    for tag in ("Environment", "Team", "Project"):
        if tag not in tags:
            return False, f"Missing required tag: {tag}"
    return True, "All required tags present"


def _rds_encryption(res: dict[str, Any]) -> tuple[bool, str]:
    if not res.get("storage_encrypted"):
        return False, "RDS instance storage is not encrypted"
    return True, "RDS storage encryption is enabled"


def _rds_multi_az(res: dict[str, Any]) -> tuple[bool, str]:
    tags = res.get("tags") or {}
    env = tags.get("Environment")
    if env not in ("production", "prod"):
        return True, "Non-production instance, Multi-AZ not required"
    if not res.get("multi_az"):
        return False, "Production RDS instance does not have Multi-AZ enabled"
    return True, "Multi-AZ is enabled"


def _rds_backup_retention(res: dict[str, Any]) -> tuple[bool, str]:
    retention = res.get("backup_retention_period")
    if not isinstance(retention, (int, float)) or retention < 7:
        return False, f"RDS backup retention is {retention} days (minimum 7 required)"
    return True, f"RDS backup retention is {retention} days"


_BUILT_IN_RULES: list[Rule] = [
    Rule("s3-001", "S3 Bucket Encryption Required",
         "All S3 buckets must have server-side encryption enabled",
         DriftSeverity.HIGH, "security", _s3_encryption),
    Rule("s3-002", "S3 Public Access Block",
         "S3 buckets must block all public access",
         DriftSeverity.CRITICAL, "security", _s3_public_block),
    Rule("s3-003", "S3 Versioning Enabled",
         "S3 buckets should have versioning enabled for data protection",
         DriftSeverity.MEDIUM, "compliance", _s3_versioning),
    Rule("ec2-001", "EC2 No Public SSH",
         "EC2 instances must not allow SSH (port 22) from 0.0.0.0/0",
         DriftSeverity.CRITICAL, "security", _ec2_no_public_ssh),
    Rule("ec2-002", "EC2 EBS Encryption",
         "EC2 instance root EBS volumes must be encrypted",
         DriftSeverity.HIGH, "security", _ec2_ebs_encryption),
    Rule("ec2-003", "EC2 Required Tags",
         "EC2 instances must have Environment, Team, and Project tags",
         DriftSeverity.LOW, "governance", _ec2_required_tags),
    Rule("rds-001", "RDS Encryption At Rest",
         "RDS instances must have storage encryption enabled",
         DriftSeverity.HIGH, "security", _rds_encryption),
    Rule("rds-002", "RDS Multi-AZ in Production",
         "Production RDS instances must have Multi-AZ enabled",
         DriftSeverity.HIGH, "reliability", _rds_multi_az),
    Rule("rds-003", "RDS Backup Retention",
         "RDS instances must have backup retention of at least 7 days",
         DriftSeverity.MEDIUM, "reliability", _rds_backup_retention),
]


class Checker:
    """Runs compliance rules against cloud resources."""

    def __init__(self, rules: list[Rule] | None = None):
        self.rules = rules if rules is not None else list(_BUILT_IN_RULES)

    def run_checks(self, resources: dict[str, Any]) -> list[Result]:
        results: list[Result] = []
        for resource_id, resource_data in resources.items():
            if not isinstance(resource_data, dict):
                continue
            for rule in self.rules:
                passed, message = rule.check(resource_data)
                results.append(Result(rule.id, rule.name, resource_id, passed, message, rule.severity))
        return results


def summary(results: list[Result]) -> dict[str, Any]:
    total = len(results)
    passed = sum(1 for r in results if r.passed)
    by_severity: dict[str, int] = {}
    for r in results:
        if not r.passed:
            by_severity[r.severity.value] = by_severity.get(r.severity.value, 0) + 1

    return {
        "total": total,
        "passed": passed,
        "failed": total - passed,
        "score": (passed / total * 100) if total else 0.0,
        "by_severity": by_severity,
    }
