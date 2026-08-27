"""Drift detection service: compares actual cloud state against expected IaC state."""
from __future__ import annotations

import json
import logging
from datetime import datetime, timezone

import psycopg2.extras

from app import database
from app.models import CloudProvider, DriftItem, DriftSeverity, FixSuggestion, Scan
from app.providers import CloudProvider as CloudProviderAdapter

logger = logging.getLogger("coherence")

_CRITICAL_PATTERNS = ("prod", "production", "database", "primary")
_RESOURCE_TYPE_PREFIXES = {"i-": "ec2", "s-": "s3", "db-": "rds"}


class DriftService:
    def __init__(
        self,
        aws_provider: CloudProviderAdapter,
        gcp_provider: CloudProviderAdapter,
        azure_provider: CloudProviderAdapter,
    ):
        self.providers = {
            CloudProvider.AWS: aws_provider,
            CloudProvider.GCP: gcp_provider,
            CloudProvider.AZURE: azure_provider,
        }

    def scan_drift(self, scan: Scan) -> None:
        logger.info("Starting drift scan", extra={"scan_id": scan.id})
        self._update_scan_status(scan.id, "running")

        provider = self.providers.get(scan.cloud_provider)
        if provider is None:
            self._update_scan_status(scan.id, "failed")
            raise ValueError(f"unsupported cloud provider: {scan.cloud_provider}")

        try:
            actual_resources = provider.get_resources(scan.regions, scan.resource_types)
        except Exception:
            logger.exception("Failed to fetch cloud resources")
            self._update_scan_status(scan.id, "failed")
            raise

        expected_resources = self._get_expected_state()

        drift_items = self._detect_drift(actual_resources, expected_resources, scan)

        for item in drift_items:
            try:
                self._enrich_drift_with_audit(item, provider)
            except Exception:
                logger.warning("Failed to enrich drift with audit info", exc_info=True)

        for item in drift_items:
            try:
                self._save_drift_item(item)
            except Exception:
                logger.exception("Failed to save drift item")

        scan.drift_count = len(drift_items)
        scan.completed_at = datetime.now(timezone.utc)
        self._update_scan_status(scan.id, "completed")

        logger.info(
            "Drift scan completed",
            extra={"scan_id": scan.id, "drift_count": len(drift_items)},
        )

    def _detect_drift(
        self,
        actual_resources: dict,
        expected_resources: dict,
        scan: Scan,
    ) -> list[DriftItem]:
        drift_items: list[DriftItem] = []

        for resource_id, actual_state in actual_resources.items():
            if resource_id not in expected_resources:
                severity = (
                    DriftSeverity.HIGH
                    if self._is_critical_resource(resource_id)
                    else DriftSeverity.MEDIUM
                )
                item = DriftItem.new(
                    scan.id, resource_id, self._extract_resource_type(resource_id),
                    scan.cloud_provider, severity,
                )
                item.title = f"Unmanaged resource: {resource_id}"
                item.description = (
                    "This resource exists in your cloud account but is not defined "
                    "in your Infrastructure-as-Code"
                )
                item.category = "unmanaged"
                item.actual_state = actual_state
                item.expected_state = None
                item.fix_suggestions = [
                    FixSuggestion(
                        id="suggest_1",
                        title="Add to IaC",
                        description="Import or define this resource in your Terraform/CloudFormation",
                        action="update_code",
                        risk="low",
                    ),
                    FixSuggestion(
                        id="suggest_2",
                        title="Delete resource",
                        description="Remove this unmanaged resource from your cloud account",
                        action="safe_remediate",
                        risk="medium",
                        is_safe_auto=False,
                    ),
                ]
                drift_items.append(item)
            elif not self._states_equal(actual_state, expected_resources[resource_id]):
                item = DriftItem.new(
                    scan.id, resource_id, self._extract_resource_type(resource_id),
                    scan.cloud_provider, DriftSeverity.MEDIUM,
                )
                item.title = f"Configuration drift: {resource_id}"
                item.description = "Resource configuration differs from Infrastructure-as-Code"
                item.category = "configuration"
                item.actual_state = actual_state
                item.expected_state = expected_resources[resource_id]
                item.difference = self._calculate_difference(actual_state, expected_resources[resource_id])
                item.fix_suggestions = [
                    FixSuggestion(
                        id="suggest_1",
                        title="Apply IaC changes",
                        description="Update resources to match Infrastructure-as-Code",
                        action="apply_iac",
                        risk="low",
                        is_safe_auto=True,
                    ),
                    FixSuggestion(
                        id="suggest_2",
                        title="Update IaC",
                        description="Update your code to match current cloud configuration",
                        action="update_code",
                        risk="none",
                    ),
                ]
                drift_items.append(item)

        for resource_id in expected_resources:
            if resource_id not in actual_resources:
                item = DriftItem.new(
                    scan.id, resource_id, self._extract_resource_type(resource_id),
                    scan.cloud_provider, DriftSeverity.CRITICAL,
                )
                item.title = f"Missing resource: {resource_id}"
                item.description = (
                    "This resource is defined in your Infrastructure-as-Code but "
                    "does not exist in your cloud account"
                )
                item.category = "missing"
                item.fix_suggestions = [
                    FixSuggestion(
                        id="suggest_1",
                        title="Create resource",
                        description="Apply IaC to create this missing resource",
                        action="apply_iac",
                        risk="low",
                        is_safe_auto=True,
                    ),
                ]
                drift_items.append(item)

        return drift_items

    def _enrich_drift_with_audit(self, item: DriftItem, provider: CloudProviderAdapter) -> None:
        events = provider.get_audit_trail(item.resource_id)
        if events:
            last_event = events[0]
            from app.models import AuditInfo

            item.audit_info = AuditInfo(
                change_type=last_event.change_type,
                changed_by=last_event.principal,
                change_time=last_event.timestamp,
                source=last_event.source,
                event_id=last_event.event_id,
                source_ip=last_event.source_ip,
                user_agent=last_event.user_agent,
            )

    def _get_expected_state(self) -> dict:
        # In a real implementation this would parse Terraform/CloudFormation
        # (e.g. `terraform show -json`) and compare against expected state.
        return {
            "i-1234567890abcdef0": {
                "name": "web-server-01",
                "instance_type": "t3.medium",
                "state": "running",
                "availability_zone": "us-east-1a",
            },
            "s3://my-app-bucket": {
                "versioning": True,
                "encryption": "AES256",
                "public_access_block": True,
            },
        }

    @staticmethod
    def _states_equal(actual, expected) -> bool:
        return json.dumps(actual, sort_keys=True, default=str) == json.dumps(
            expected, sort_keys=True, default=str
        )

    @staticmethod
    def _calculate_difference(actual, expected) -> dict:
        return {"actual": actual, "expected": expected}

    @staticmethod
    def _extract_resource_type(resource_id: str) -> str:
        for prefix, resource_type in _RESOURCE_TYPE_PREFIXES.items():
            if resource_id.startswith(prefix):
                return resource_type
        return "unknown"

    @staticmethod
    def _is_critical_resource(resource_id: str) -> bool:
        return any(resource_id.startswith(pattern) for pattern in _CRITICAL_PATTERNS)

    @staticmethod
    def _update_scan_status(scan_id: str, status: str) -> None:
        conn = database.get_conn()
        try:
            with conn.cursor() as cur:
                cur.execute(
                    "UPDATE scans SET status = %s, updated_at = %s WHERE id = %s",
                    (status, datetime.now(timezone.utc), scan_id),
                )
            conn.commit()
        finally:
            database.put_conn(conn)

    @staticmethod
    def _save_drift_item(item: DriftItem) -> None:
        conn = database.get_conn()
        try:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO drift_items (
                        id, scan_id, resource_id, resource_type, cloud_provider, region,
                        severity, category, title, description, expected_state, actual_state,
                        difference, remediation_status, is_resolved, created_at, updated_at
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at
                    """,
                    (
                        item.id, item.scan_id, item.resource_id, item.resource_type,
                        item.cloud_provider.value, item.region, item.severity.value, item.category,
                        item.title, item.description,
                        psycopg2.extras.Json(item.expected_state) if item.expected_state is not None else None,
                        psycopg2.extras.Json(item.actual_state) if item.actual_state is not None else None,
                        psycopg2.extras.Json(item.difference) if item.difference is not None else None,
                        item.remediation_status.value, item.is_resolved,
                        item.created_at, item.updated_at,
                    ),
                )
            conn.commit()
        finally:
            database.put_conn(conn)

    @staticmethod
    def get_drift_items(scan_id: str) -> list[dict]:
        conn = database.get_conn()
        try:
            with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
                cur.execute(
                    """
                    SELECT id, scan_id, resource_id, resource_type, cloud_provider, region,
                           severity, category, title, description, is_resolved, created_at, updated_at
                    FROM drift_items WHERE scan_id = %s ORDER BY created_at DESC
                    """,
                    (scan_id,),
                )
                return [dict(row) for row in cur.fetchall()]
        finally:
            database.put_conn(conn)
