"""Notification backends for drift, scan, and remediation events."""
from __future__ import annotations

import logging
import threading
from dataclasses import dataclass
from datetime import datetime, timezone
from enum import Enum
from typing import Optional, Protocol

import requests 

from app.models import DriftItem, Scan, RemediationRequest

logger = logging.getLogger("coherence")


class AlertLevel(str, Enum):
    INFO = "info"
    WARNING = "warning"
    CRITICAL = "critical"


@dataclass
class Alert:
    level: AlertLevel
    title: str
    message: str
    timestamp: datetime
    drift_item: Optional[DriftItem] = None
    scan_id: str = ""
    action_url: str = ""


class Notifier(Protocol):
    def name(self) -> str: ...
    def send(self, alert: Alert) -> None: ...


class Manager:
    """Dispatches alerts to all registered notification backends concurrently."""

    def __init__(self):
        self._notifiers: list[Notifier] = []

    def add_notifier(self, notifier: Notifier) -> None:
        self._notifiers.append(notifier)

    def send_alert(self, alert: Alert) -> None:
        for notifier in self._notifiers:
            def _dispatch(n=notifier):
                try:
                    n.send(alert)
                except Exception:
                    logger.exception("Failed to send alert via %s", n.name())

            threading.Thread(target=_dispatch, daemon=True).start()


class SlackNotifier:
    def __init__(self, webhook_url: str, channel: str):
        self.webhook_url = webhook_url
        self.channel = channel

    def name(self) -> str:
        return "slack"

    def send(self, alert: Alert) -> None:
        color = {"info": "#36a64f", "warning": "#f59e0b", "critical": "#dc2626"}[alert.level.value]
        payload = {
            "channel": self.channel,
            "attachments": [
                {
                    "color": color,
                    "title": alert.title,
                    "text": alert.message,
                    "fields": [
                        {"title": "Severity", "value": alert.level.value, "short": "true"},
                        {"title": "Time", "value": alert.timestamp.isoformat(), "short": "true"},
                    ],
                    "actions": [{"type": "button", "text": "View in Coherence", "url": alert.action_url}],
                    "footer": "Coherence · Infrastructure Drift Detection",
                }
            ],
        }
        resp = requests.post(self.webhook_url, json=payload, timeout=10)
        if resp.status_code != 200:
            raise RuntimeError(f"slack: non-200 response: {resp.status_code}")


class PagerDutyNotifier:
    def __init__(self, integration_key: str):
        self.integration_key = integration_key
        self.api_url = "https://events.pagerduty.com/v2/enqueue"

    def name(self) -> str:
        return "pagerduty"

    def send(self, alert: Alert) -> None:
        severity = {"info": "info", "warning": "warning", "critical": "critical"}[alert.level.value]
        payload = {
            "routing_key": self.integration_key,
            "event_action": "trigger",
            "payload": {
                "summary": alert.title,
                "severity": severity,
                "source": "coherence",
                "timestamp": alert.timestamp.isoformat(),
                "custom_details": {
                    "message": alert.message,
                    "scan_id": alert.scan_id,
                    "action_url": alert.action_url,
                },
            },
        }
        resp = requests.post(self.api_url, json=payload, timeout=10)
        if resp.status_code not in (200, 202):
            raise RuntimeError(f"pagerduty: non-2xx response: {resp.status_code}")


def new_critical_drift_alert(drift: DriftItem, dashboard_url: str) -> Alert:
    return Alert(
        level=AlertLevel.CRITICAL,
        title=f"Critical Drift Detected: {drift.resource_id}",
        message=drift.description,
        drift_item=drift,
        timestamp=datetime.now(timezone.utc),
        action_url=f"{dashboard_url}/drifts/{drift.id}",
    )


def new_scan_completed_alert(scan: Scan, dashboard_url: str) -> Alert:
    level = AlertLevel.WARNING if scan.drift_count > 0 else AlertLevel.INFO
    return Alert(
        level=level,
        title=f"Scan completed: {scan.drift_count} drift(s) found",
        message=f"Scan {scan.id} finished scanning {scan.cloud_provider.value} resources.",
        scan_id=scan.id,
        timestamp=datetime.now(timezone.utc),
        action_url=f"{dashboard_url}/scans/{scan.id}",
    )


def new_remediation_failed_alert(req: RemediationRequest, drift: DriftItem, dashboard_url: str) -> Alert:
    return Alert(
        level=AlertLevel.CRITICAL,
        title=f"Remediation Failed: {drift.resource_id}",
        message=f"Remediation action '{req.action_type}' failed for resource {drift.resource_id}.",
        drift_item=drift,
        timestamp=datetime.now(timezone.utc),
        action_url=f"{dashboard_url}/remediations/{req.id}",
    )
