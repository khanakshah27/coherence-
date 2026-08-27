"""Data models shared across the drift detection, compliance, and remediation services."""
from __future__ import annotations

import uuid
from dataclasses import dataclass, field
from datetime import datetime, timezone
from enum import Enum
from typing import Any, Optional


def utcnow() -> datetime:
    return datetime.now(timezone.utc)


class DriftSeverity(str, Enum):
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    INFO = "info"


class CloudProvider(str, Enum):
    AWS = "aws"
    GCP = "gcp"
    AZURE = "azure"


class RemediationStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    EXECUTING = "executing"
    SUCCESS = "success"
    FAILED = "failed"
    ROLLED_BACK = "rolled_back"


@dataclass
class ImpactAnalysis:
    cost_impact: float = 0.0
    performance_impact: str = "none"
    security_impact: str = "none"
    compliance_rules: list[str] = field(default_factory=list)
    affected_services: list[str] = field(default_factory=list)


@dataclass
class AuditInfo:
    change_type: str = ""
    changed_by: str = ""
    change_time: Optional[datetime] = None
    source: str = ""
    event_id: str = ""
    source_ip: str = ""
    user_agent: str = ""


@dataclass
class FixSuggestion:
    id: str
    title: str
    description: str
    action: str  # apply_iac, update_code, safe_remediate, manual_fix
    is_safe_auto: bool = False
    steps: list[str] = field(default_factory=list)
    risk: str = "none"
    time_to_fix_seconds: int = 0
    estimated_cost_savings: float = 0.0


@dataclass
class Scan:
    id: str
    cloud_provider: CloudProvider
    regions: list[str]
    resource_types: list[str]
    status: str = "pending"
    drift_count: int = 0
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    created_at: datetime = field(default_factory=utcnow)

    @classmethod
    def new(cls, provider: CloudProvider, regions: list[str], resource_types: list[str]) -> "Scan":
        return cls(
            id=str(uuid.uuid4()),
            cloud_provider=provider,
            regions=regions,
            resource_types=resource_types,
        )


@dataclass
class DriftItem:
    id: str
    scan_id: str
    resource_id: str
    resource_type: str
    cloud_provider: CloudProvider
    severity: DriftSeverity
    region: str = ""
    category: str = ""
    title: str = ""
    description: str = ""
    expected_state: Optional[dict[str, Any]] = None
    actual_state: Optional[dict[str, Any]] = None
    difference: Optional[dict[str, Any]] = None
    impact_analysis: Optional[ImpactAnalysis] = None
    audit_info: Optional[AuditInfo] = None
    fix_suggestions: list[FixSuggestion] = field(default_factory=list)
    remediation_status: RemediationStatus = RemediationStatus.PENDING
    is_resolved: bool = False
    created_at: datetime = field(default_factory=utcnow)
    updated_at: datetime = field(default_factory=utcnow)

    @classmethod
    def new(
        cls,
        scan_id: str,
        resource_id: str,
        resource_type: str,
        provider: CloudProvider,
        severity: DriftSeverity,
    ) -> "DriftItem":
        return cls(
            id=str(uuid.uuid4()),
            scan_id=scan_id,
            resource_id=resource_id,
            resource_type=resource_type,
            cloud_provider=provider,
            severity=severity,
        )


@dataclass
class RemediationRequest:
    id: str
    drift_id: str
    action_type: str
    status: RemediationStatus = RemediationStatus.PENDING
    dry_run: bool = False
    approval_status: str = "pending"
    approved_by: str = ""
    approved_at: Optional[datetime] = None
    executed_at: Optional[datetime] = None
    result: Optional[dict[str, Any]] = None
    error: str = ""
    rolled_back_at: Optional[datetime] = None
    created_at: datetime = field(default_factory=utcnow)
    updated_at: datetime = field(default_factory=utcnow)

    @classmethod
    def new(cls, drift_id: str, action_type: str) -> "RemediationRequest":
        return cls(id=str(uuid.uuid4()), drift_id=drift_id, action_type=action_type)


@dataclass
class ComplianceRule:
    id: str
    name: str
    description: str
    severity: DriftSeverity
    category: str
    resource_types: list[str] = field(default_factory=list)
    rules: dict[str, Any] = field(default_factory=dict)
    enabled: bool = True
    created_at: datetime = field(default_factory=utcnow)


@dataclass
class Report:
    id: str
    scan_id: str
    cloud_provider: CloudProvider
    total_resources: int = 0
    drifted_resources: int = 0
    drift_percentage: float = 0.0
    by_severity: dict[str, int] = field(default_factory=dict)
    by_category: dict[str, int] = field(default_factory=dict)
    cost_impact: float = 0.0
    compliance_status: str = ""
    recommendations: list[str] = field(default_factory=list)
    created_at: datetime = field(default_factory=utcnow)
