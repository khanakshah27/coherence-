"""Server-rendered HTML pages that replace the former React dashboard."""
from __future__ import annotations

import threading
import uuid
from datetime import datetime, timezone
from typing import Any

import psycopg2.extras
from fastapi import APIRouter, Form, Request
from fastapi.responses import RedirectResponse
from fastapi.templating import Jinja2Templates

from app import database
from app.drift import DriftService

router = APIRouter()
templates = Jinja2Templates(directory="app/templates")

_drift_service: DriftService | None = None


def set_drift_service(service: DriftService) -> None:
    global _drift_service
    _drift_service = service


# The KPI tiles and charts below mirror the original dashboard's demo data,
# which was illustrative and not derived from live scan results.
_SEVERITY_BARS = [
    ("Critical", 12, "#dc2626"),
    ("High", 24, "#ea580c"),
    ("Medium", 42, "#f59e0b"),
    ("Low", 89, "#10b981"),
]
_RESOURCE_TYPES = [
    {"type": "EC2", "count": 234, "drifted": 12},
    {"type": "S3", "count": 89, "drifted": 5},
    {"type": "RDS", "count": 45, "drifted": 8},
    {"type": "Lambda", "count": 156, "drifted": 3},
]
_COMPLIANCE_DOMAINS = [
    {"subject": "Encryption", "score": 90},
    {"subject": "Access Ctrl", "score": 68},
    {"subject": "Tagging", "score": 55},
    {"subject": "Backup", "score": 82},
    {"subject": "Networking", "score": 72},
    {"subject": "Logging", "score": 95},
]
_COMPLIANCE_RULES = [
    {"rule": "S3 Bucket Encryption Required", "passed": True},
    {"rule": "RDS Multi-AZ in Production", "passed": False},
    {"rule": "EC2 IMDSv2 Enforced", "passed": True},
    {"rule": "CloudTrail Logging Enabled", "passed": True},
    {"rule": "VPC Flow Logs Enabled", "passed": False},
    {"rule": "IAM Password Policy Enforced", "passed": True},
]


def _now() -> datetime:
    return datetime.now(timezone.utc)


_MOCK_DRIFTS = [
    {"id": "d1", "scan_id": "s1", "resource_id": "i-0abc123", "resource_type": "ec2", "cloud_provider": "aws",
     "severity": "critical", "category": "security", "title": "Security group allows 0.0.0.0/0 on port 22",
     "description": "EC2 instance has SSH open to the world, violating security policy", "is_resolved": False},
    {"id": "d2", "scan_id": "s1", "resource_id": "my-app-bucket", "resource_type": "s3", "cloud_provider": "aws",
     "severity": "high", "category": "compliance", "title": "S3 bucket versioning disabled",
     "description": "Versioning was manually disabled on the bucket, IaC expects it enabled", "is_resolved": False},
    {"id": "d3", "scan_id": "s1", "resource_id": "i-0def456", "resource_type": "ec2", "cloud_provider": "aws",
     "severity": "medium", "category": "configuration", "title": "Instance type changed from t3.medium to t3.large",
     "description": "Instance was resized manually in the console, differs from IaC definition", "is_resolved": False},
    {"id": "d4", "scan_id": "s1", "resource_id": "prod-postgres", "resource_type": "rds", "cloud_provider": "aws",
     "severity": "high", "category": "configuration", "title": "RDS multi-AZ disabled",
     "description": "Multi-AZ was disabled on production DB, IaC expects it enabled", "is_resolved": False},
    {"id": "d5", "scan_id": "s1", "resource_id": "i-0ghi789", "resource_type": "ec2", "cloud_provider": "aws",
     "severity": "low", "category": "cost", "title": "Missing cost allocation tags",
     "description": "Instance is missing required \"CostCenter\" and \"Team\" tags", "is_resolved": True},
]
_MOCK_REMEDIATIONS = [
    {"id": "r1", "drift_id": "d2", "action_type": "apply_iac", "status": "pending",
     "approval_status": "pending", "dry_run": True},
    {"id": "r2", "drift_id": "d3", "action_type": "safe_remediate", "status": "success",
     "approval_status": "approved", "dry_run": False},
    {"id": "r3", "drift_id": "d4", "action_type": "apply_iac", "status": "failed",
     "approval_status": "approved", "dry_run": False},
]


def _fetch_scans() -> list[dict[str, Any]]:
    conn = database.get_conn()
    try:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            cur.execute(
                """SELECT id, cloud_provider, status, drift_count, created_at
                   FROM scans ORDER BY created_at DESC LIMIT 100"""
            )
            return [dict(row) for row in cur.fetchall()]
    finally:
        database.put_conn(conn)


def _fetch_drifts(severity: str | None, show_resolved: bool) -> list[dict[str, Any]]:
    conn = database.get_conn()
    try:
        with conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor) as cur:
            query = """SELECT id, scan_id, resource_id, resource_type, cloud_provider, severity,
                              category, title, description, is_resolved, created_at
                       FROM drift_items WHERE 1=1"""
            args: list[Any] = []
            if severity:
                query += " AND severity = %s"
                args.append(severity)
            if not show_resolved:
                query += " AND is_resolved = false"
            query += " ORDER BY created_at DESC LIMIT 1000"
            cur.execute(query, args)
            rows = [dict(row) for row in cur.fetchall()]
    finally:
        database.put_conn(conn)

    if rows:
        return rows

    mock = _MOCK_DRIFTS
    if severity:
        mock = [d for d in mock if d["severity"] == severity]
    if not show_resolved:
        mock = [d for d in mock if not d["is_resolved"]]
    return mock


@router.get("/")
def dashboard(request: Request):
    scans = _fetch_scans()
    return templates.TemplateResponse(
        "dashboard.html",
        {
            "request": request,
            "active_page": "dashboard",
            "scans": scans,
            "total_resources": 524,
            "total_drift": 167,
            "compliance_score": 72,
            "severity_counts": {"critical": 12},
            "severity_bars": _SEVERITY_BARS,
            "resource_types": _RESOURCE_TYPES,
        },
    )


@router.post("/scans")
def create_scan_form(cloud_provider: str = Form(...), regions: str = Form(...)):
    from app.models import CloudProvider, Scan

    region_list = [r.strip() for r in regions.split(",") if r.strip()]
    scan_id = str(uuid.uuid4())
    created_at = _now()

    conn = database.get_conn()
    try:
        with conn.cursor() as cur:
            cur.execute(
                """INSERT INTO scans (id, cloud_provider, regions, resource_types, status, created_at)
                   VALUES (%s, %s, %s, %s, %s, %s)""",
                (scan_id, cloud_provider, region_list, ["ec2", "s3", "rds"], "pending", created_at),
            )
        conn.commit()
    finally:
        database.put_conn(conn)

    scan = Scan(
        id=scan_id,
        cloud_provider=CloudProvider(cloud_provider),
        regions=region_list,
        resource_types=["ec2", "s3", "rds"],
        created_at=created_at,
    )

    def _run_scan():
        try:
            assert _drift_service is not None
            _drift_service.scan_drift(scan)
        except Exception:
            pass

    threading.Thread(target=_run_scan, daemon=True).start()
    return RedirectResponse("/", status_code=303)


@router.get("/drifts")
def drifts_page(request: Request, severity: str = "", show_resolved: bool = False):
    drifts = _fetch_drifts(severity or None, show_resolved)
    return templates.TemplateResponse(
        "drifts.html",
        {
            "request": request,
            "active_page": "drifts",
            "drifts": drifts,
            "filter_severity": severity,
            "show_resolved": show_resolved,
        },
    )


@router.post("/drifts/{drift_id}/resolve")
def resolve_drift_form(drift_id: str, severity: str = Form(""), show_resolved: str = Form("")):
    conn = database.get_conn()
    try:
        with conn.cursor() as cur:
            cur.execute(
                "UPDATE drift_items SET is_resolved = true, remediation_status = %s, updated_at = %s WHERE id = %s",
                ("resolved", _now(), drift_id),
            )
        conn.commit()
    finally:
        database.put_conn(conn)

    qs = f"?severity={severity}" if severity else ""
    qs += ("&" if qs else "?") + f"show_resolved={'1' if show_resolved else ''}"
    return RedirectResponse(f"/drifts{qs}", status_code=303)


@router.get("/remediations")
def remediations_page(request: Request):
    counts = {"pending": 0, "executing": 0, "success": 0, "failed": 0}
    for r in _MOCK_REMEDIATIONS:
        if r["approval_status"] == "pending":
            counts["pending"] += 1
        if r["status"] in counts:
            counts[r["status"]] += 1
    return templates.TemplateResponse(
        "remediations.html",
        {
            "request": request,
            "active_page": "remediations",
            "remediations": _MOCK_REMEDIATIONS,
            "summary": counts,
        },
    )


@router.post("/remediations/{remediation_id}/approve")
@router.post("/remediations/{remediation_id}/execute")
@router.post("/remediations/{remediation_id}/rollback")
def remediation_action(remediation_id: str):
    return RedirectResponse("/remediations", status_code=303)


@router.get("/reports")
def reports_page(request: Request):
    passed = sum(1 for r in _COMPLIANCE_RULES if r["passed"])
    total = len(_COMPLIANCE_RULES)
    return templates.TemplateResponse(
        "reports.html",
        {
            "request": request,
            "active_page": "reports",
            "reports": [],
            "compliance_domains": _COMPLIANCE_DOMAINS,
            "compliance_rules": _COMPLIANCE_RULES,
            "passed_rules": passed,
            "total_rules": total,
            "compliance_score": round(passed / total * 100) if total else 0,
        },
    )


@router.post("/reports/generate")
def generate_report_form():
    return RedirectResponse("/reports", status_code=303)
