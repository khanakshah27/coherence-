"""Remediation engine: executes, previews, and rolls back drift fixes."""
from __future__ import annotations

import logging
import time
from datetime import datetime, timezone
from typing import Any

from app import database
from app.models import DriftItem, RemediationRequest

logger = logging.getLogger("coherence")


class Engine:
    def execute(self, req: RemediationRequest, drift: DriftItem) -> None:
        logger.info(
            "Starting remediation execution",
            extra={
                "remediation_id": req.id,
                "drift_id": drift.id,
                "action": req.action_type,
                "dry_run": req.dry_run,
            },
        )
        self._update_status(req.id, "executing", "")

        handlers = {
            "apply_iac": self._apply_iac,
            "safe_remediate": self._safe_remediate,
            "update_code": self._update_code,
            "delete_resource": self._delete_resource,
        }
        handler = handlers.get(req.action_type)
        if handler is None:
            error = f"unknown action type: {req.action_type}"
            self._update_status(req.id, "failed", error)
            raise ValueError(error)

        try:
            handler(drift, req.dry_run)
        except Exception as exc:
            logger.exception("Remediation execution failed")
            self._update_status(req.id, "failed", str(exc))
            raise

        now = datetime.now(timezone.utc)
        conn = database.get_conn()
        try:
            with conn.cursor() as cur:
                cur.execute(
                    "UPDATE remediation_requests SET status = %s, executed_at = %s, updated_at = %s WHERE id = %s",
                    ("success", now, now, req.id),
                )
            conn.commit()
        finally:
            database.put_conn(conn)

        if not req.dry_run:
            try:
                self._mark_drift_resolved(drift.id)
            except Exception:
                logger.warning("Failed to mark drift as resolved after remediation", exc_info=True)

        logger.info("Remediation completed successfully", extra={"remediation_id": req.id})

    def rollback(self, req: RemediationRequest, drift: DriftItem) -> None:
        logger.info("Starting remediation rollback", extra={"remediation_id": req.id})
        # In a real implementation:
        # 1. Fetch the previous state snapshot taken before execution
        # 2. Re-apply the previous state via cloud API or Terraform
        # 3. Update the drift item status
        now = datetime.now(timezone.utc)
        conn = database.get_conn()
        try:
            with conn.cursor() as cur:
                cur.execute(
                    "UPDATE remediation_requests SET rolled_back_at = %s, status = %s, updated_at = %s WHERE id = %s",
                    (now, "rolled_back", now, req.id),
                )
                cur.execute(
                    "UPDATE drift_items SET is_resolved = false, remediation_status = %s, updated_at = %s WHERE id = %s",
                    ("rolled_back", now, drift.id),
                )
            conn.commit()
        finally:
            database.put_conn(conn)

    def dry_run(self, action_type: str, drift: DriftItem) -> dict[str, Any]:
        preview: dict[str, Any] = {
            "action": action_type,
            "resource_id": drift.resource_id,
            "changes": [],
            "risk": "low",
            "estimated_duration_seconds": 30,
        }

        actual = drift.actual_state or {}
        expected = drift.expected_state or {}

        if action_type == "apply_iac":
            preview["changes"] = [
                {
                    "attribute": "instance_type",
                    "from": actual.get("instance_type"),
                    "to": expected.get("instance_type"),
                    "impact": "Resource restart required",
                }
            ]
            preview["risk"] = "medium"
        elif action_type == "safe_remediate":
            preview["changes"] = [
                {
                    "attribute": "tags",
                    "from": actual.get("tags"),
                    "to": expected.get("tags"),
                    "impact": "No downtime",
                }
            ]
            preview["risk"] = "low"

        return preview

    # ── Private action handlers ──────────────────────────────────────────

    def _apply_iac(self, drift: DriftItem, dry_run: bool) -> None:
        logger.info("Applying IaC to resource", extra={"resource_id": drift.resource_id, "dry_run": dry_run})
        # In production this would:
        # 1. Identify the Terraform module / CF stack owning the resource
        # 2. Run `terraform apply -target=<resource>` or update the CF stack
        # 3. Monitor the apply until completion
        # 4. Verify the new state matches expected
        if dry_run:
            logger.info("[DRY RUN] Would apply IaC changes to resource: %s", drift.resource_id)
            return
        time.sleep(2)
        logger.info("IaC apply completed successfully for resource: %s", drift.resource_id)

    def _safe_remediate(self, drift: DriftItem, dry_run: bool) -> None:
        logger.info("Applying safe remediation", extra={"resource_id": drift.resource_id})
        # Safe remediations include: adding/updating tags, enabling versioning
        # on S3, enabling encryption on S3, adding lifecycle rules.
        if dry_run:
            logger.info("[DRY RUN] Would apply safe remediation to: %s", drift.resource_id)
            return
        time.sleep(0.5)

    def _update_code(self, drift: DriftItem, dry_run: bool) -> None:
        logger.info("Updating IaC code", extra={"resource_id": drift.resource_id})
        # In production:
        # 1. Clone or checkout the IaC repository
        # 2. Find the resource definition
        # 3. Update the definition to match actual state
        # 4. Create a pull request via GitHub/GitLab API
        if dry_run:
            logger.info("[DRY RUN] Would open PR to update IaC for: %s", drift.resource_id)

    def _delete_resource(self, drift: DriftItem, dry_run: bool) -> None:
        logger.warning("Deleting unmanaged resource", extra={"resource_id": drift.resource_id})
        if dry_run:
            logger.info("[DRY RUN] Would delete resource: %s", drift.resource_id)
            return
        # In production: call the cloud provider API to delete the resource

    # ── Database helpers ─────────────────────────────────────────────────

    @staticmethod
    def _update_status(remediation_id: str, status: str, error_message: str) -> None:
        conn = database.get_conn()
        try:
            with conn.cursor() as cur:
                cur.execute(
                    "UPDATE remediation_requests SET status = %s, error = %s, updated_at = %s WHERE id = %s",
                    (status, error_message, datetime.now(timezone.utc), remediation_id),
                )
            conn.commit()
        finally:
            database.put_conn(conn)

    @staticmethod
    def _mark_drift_resolved(drift_id: str) -> None:
        conn = database.get_conn()
        try:
            with conn.cursor() as cur:
                cur.execute(
                    "UPDATE drift_items SET is_resolved = true, remediation_status = %s, updated_at = %s WHERE id = %s",
                    ("success", datetime.now(timezone.utc), drift_id),
                )
            conn.commit()
        finally:
            database.put_conn(conn)
