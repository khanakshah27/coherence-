"""REST API routes: /api/v1/scans, /drifts, /reports, /remediations."""
from __future__ import annotations

import logging
import threading
import uuid
from datetime import datetime, timezone
from typing import Any, Optional

import psycopg2.extras
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from app import database
from app.drift import DriftService
from app.models import RemediationRequest

logger = logging.getLogger("coherence")

router = APIRouter(prefix="/api/v1")

_drift_service: DriftService | None = None


def set_drift_service(service: DriftService) -> None:
    global _drift_service
    _drift_service = service


# ── Scans ────────────────────────────────────────────────────────────────

class CreateScanRequest(BaseModel):
    cloud_provider: str
    regions: list[str]
    resource_types: list[str] = []


@router.post("/scans", status_code=201)
def create_scan(req: CreateScanRequest):
    scan_id = str(uuid.uuid4())
    created_at = datetime.now(timezone.utc)

    conn = database.get_conn()
    try:
        with conn.cursor() as cur:
            cur.execute(
                """INSERT INTO scans (id, cloud_provider, regions, resource_types, status, created_at)
                   VALUES (%s, %s, %s, %s, %s, %s)""",
                (scan_id, req.cloud_provider, req.regions, req.resource_types, "pending", created_at),
            )
        conn.commit()
    except Exception:
        logger.exception("Failed to create scan")
        raise HTTPException(status_code=500, detail="Failed to create scan")
    finally:
        database.put_conn(conn)

    from app.models import CloudProvider, Scan

    scan = Scan(
        id=scan_id,
        cloud_provider=CloudProvider(req.cloud_provider),
        regions=req.regions,
        resource_types=req.resource_types,
        status="pending",
        created_at=created_at,
    )

    def _run_scan():
        try:
            assert _drift_service is not None
            _drift_service.scan_drift(scan)
        except Exception:
            logger.exception("Drift scan failed")

    threading.Thread(target=_run_scan, daemon=True).start()

    return {
        "id": scan.id,
        "cloud_provider": scan.cloud_provider.value,
        "regions": scan.regions,
        "resource_types": scan.resource_types,
        "status": scan.status,
        "created_at": scan.created_at.isoformat(),
    }


@router.get("/scans")
def list_scans():
    conn = database.get_conn()
    try:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                """SELECT id, cloud_provider, status, drift_count, started_at, completed_at, created_at
                   FROM scans ORDER BY created_at DESC LIMIT 100"""
            )
            return [dict(row) for row in cur.fetchall()]
    except Exception:
        logger.exception("Failed to list scans")
        raise HTTPException(status_code=500, detail="Failed to list scans")
    finally:
        database.put_conn(conn)


@router.get("/scans/{scan_id}")
def get_scan(scan_id: str):
    conn = database.get_conn()
    try:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                """SELECT id, cloud_provider, status, drift_count, started_at, completed_at, created_at
                   FROM scans WHERE id = %s""",
                (scan_id,),
            )
            row = cur.fetchone()
            if row is None:
                raise HTTPException(status_code=404, detail="Scan not found")
            return dict(row)
    finally:
        database.put_conn(conn)


@router.delete("/scans/{scan_id}")
def delete_scan(scan_id: str):
    conn = database.get_conn()
    try:
        with conn.cursor() as cur:
            cur.execute("DELETE FROM scans WHERE id = %s", (scan_id,))
        conn.commit()
        return {"message": "Scan deleted"}
    except Exception:
        logger.exception("Failed to delete scan")
        raise HTTPException(status_code=500, detail="Failed to delete scan")
    finally:
        database.put_conn(conn)


@router.post("/scans/{scan_id}/retry")
def retry_scan(scan_id: str):
    conn = database.get_conn()
    try:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                "SELECT id, cloud_provider, regions, resource_types FROM scans WHERE id = %s",
                (scan_id,),
            )
            row = cur.fetchone()
            if row is None:
                raise HTTPException(status_code=404, detail="Scan not found")
    finally:
        database.put_conn(conn)

    from app.models import CloudProvider, Scan

    scan = Scan(
        id=row["id"],
        cloud_provider=CloudProvider(row["cloud_provider"]),
        regions=row["regions"],
        resource_types=row["resource_types"],
    )

    def _run_scan():
        try:
            assert _drift_service is not None
            _drift_service.scan_drift(scan)
        except Exception:
            logger.exception("Retry scan failed")

    threading.Thread(target=_run_scan, daemon=True).start()
    return {"message": "Scan retry started"}


# ── Drifts ───────────────────────────────────────────────────────────────

@router.get("/drifts")
def list_drifts(scan_id: Optional[str] = None, severity: Optional[str] = None):
    query = """SELECT id, scan_id, resource_id, resource_type, cloud_provider, severity,
                      category, title, description, is_resolved, created_at, updated_at
               FROM drift_items WHERE 1=1"""
    args: list[Any] = []
    if scan_id:
        query += " AND scan_id = %s"
        args.append(scan_id)
    if severity:
        query += " AND severity = %s"
        args.append(severity)
    query += " ORDER BY created_at DESC LIMIT 1000"

    conn = database.get_conn()
    try:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(query, args)
            return [dict(row) for row in cur.fetchall()]
    except Exception:
        logger.exception("Failed to list drifts")
        raise HTTPException(status_code=500, detail="Failed to list drifts")
    finally:
        database.put_conn(conn)


@router.get("/drifts/{drift_id}")
def get_drift(drift_id: str):
    conn = database.get_conn()
    try:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                """SELECT id, scan_id, resource_id, resource_type, cloud_provider, severity,
                          category, title, description, is_resolved, created_at, updated_at
                   FROM drift_items WHERE id = %s""",
                (drift_id,),
            )
            row = cur.fetchone()
            if row is None:
                raise HTTPException(status_code=404, detail="Drift not found")
            return dict(row)
    finally:
        database.put_conn(conn)


class UpdateDriftRequest(BaseModel):
    category: str = ""
    title: str = ""
    description: str = ""


@router.put("/drifts/{drift_id}")
def update_drift(drift_id: str, req: UpdateDriftRequest):
    conn = database.get_conn()
    try:
        with conn.cursor() as cur:
            cur.execute(
                "UPDATE drift_items SET category = %s, title = %s, description = %s, updated_at = %s WHERE id = %s",
                (req.category, req.title, req.description, datetime.now(timezone.utc), drift_id),
            )
        conn.commit()
        return {"message": "Drift updated"}
    except Exception:
        logger.exception("Failed to update drift")
        raise HTTPException(status_code=500, detail="Failed to update drift")
    finally:
        database.put_conn(conn)


@router.post("/drifts/{drift_id}/resolve")
def resolve_drift(drift_id: str):
    conn = database.get_conn()
    try:
        with conn.cursor() as cur:
            cur.execute(
                "UPDATE drift_items SET is_resolved = true, remediation_status = %s, updated_at = %s WHERE id = %s",
                ("resolved", datetime.now(timezone.utc), drift_id),
            )
        conn.commit()
        return {"message": "Drift resolved"}
    except Exception:
        logger.exception("Failed to resolve drift")
        raise HTTPException(status_code=500, detail="Failed to resolve drift")
    finally:
        database.put_conn(conn)


class BulkResolveRequest(BaseModel):
    drift_ids: list[str]


@router.post("/drifts/bulk-resolve")
def bulk_resolve_drifts(req: BulkResolveRequest):
    conn = database.get_conn()
    try:
        with conn.cursor() as cur:
            cur.execute(
                "UPDATE drift_items SET is_resolved = true, remediation_status = %s, updated_at = %s WHERE id = ANY(%s)",
                ("resolved", datetime.now(timezone.utc), req.drift_ids),
            )
        conn.commit()
        return {"message": "Drifts resolved"}
    except Exception:
        logger.exception("Failed to resolve drifts")
        raise HTTPException(status_code=500, detail="Failed to resolve drifts")
    finally:
        database.put_conn(conn)


# ── Reports ──────────────────────────────────────────────────────────────

@router.get("/reports")
def list_reports():
    return []


@router.get("/reports/{report_id}")
def get_report(report_id: str):
    return {}


@router.post("/reports/generate", status_code=201)
def generate_report():
    return {"message": "Report generated"}


@router.get("/reports/{report_id}/export")
def export_report(report_id: str, format: str = "json"):
    return f"Report exported as {format}"


# ── Remediations ─────────────────────────────────────────────────────────

class RemediationRequestBody(BaseModel):
    drift_id: str
    action_type: str
    dry_run: bool = False


@router.post("/remediations", status_code=201)
def request_remediation(req: RemediationRequestBody):
    remediation = RemediationRequest.new(req.drift_id, req.action_type)
    remediation.dry_run = req.dry_run
    return {
        "id": remediation.id,
        "drift_id": remediation.drift_id,
        "action_type": remediation.action_type,
        "status": remediation.status.value,
        "approval_status": remediation.approval_status,
        "dry_run": remediation.dry_run,
        "created_at": remediation.created_at.isoformat(),
    }


@router.get("/remediations")
def list_remediations():
    return []


@router.get("/remediations/{remediation_id}")
def get_remediation(remediation_id: str):
    return {}


@router.post("/remediations/{remediation_id}/approve")
def approve_remediation(remediation_id: str):
    return {"message": "Remediation approved"}


@router.post("/remediations/{remediation_id}/reject")
def reject_remediation(remediation_id: str):
    return {"message": "Remediation rejected"}


@router.post("/remediations/{remediation_id}/execute")
def execute_remediation(remediation_id: str):
    return {"message": "Remediation executed"}


@router.post("/remediations/{remediation_id}/rollback")
def rollback_remediation(remediation_id: str):
    return {"message": "Remediation rolled back"}


# ── Health & metrics ─────────────────────────────────────────────────────

@router.get("/health")
def health():
    return {"status": "healthy"}


@router.get("/metrics")
def metrics():
    return "# Coherence Metrics\n"
