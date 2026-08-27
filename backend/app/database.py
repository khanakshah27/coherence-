"""Database connection and schema migrations."""
import logging

import psycopg2
import psycopg2.extras
import psycopg2.pool

logger = logging.getLogger("coherence")

_pool: psycopg2.pool.ThreadedConnectionPool | None = None

MIGRATIONS = [
    """CREATE TABLE IF NOT EXISTS scans (
        id VARCHAR(36) PRIMARY KEY,
        cloud_provider VARCHAR(50) NOT NULL,
        regions TEXT[] NOT NULL,
        resource_types TEXT[] NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'pending',
        drift_count INT DEFAULT 0,
        started_at TIMESTAMPTZ,
        completed_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ
    )""",
    """CREATE TABLE IF NOT EXISTS drift_items (
        id VARCHAR(36) PRIMARY KEY,
        scan_id VARCHAR(36) NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
        resource_id VARCHAR(255) NOT NULL,
        resource_type VARCHAR(100) NOT NULL,
        cloud_provider VARCHAR(50) NOT NULL,
        region VARCHAR(50),
        severity VARCHAR(50) NOT NULL,
        category VARCHAR(100) NOT NULL,
        title VARCHAR(255) NOT NULL,
        description TEXT,
        expected_state JSONB,
        actual_state JSONB,
        difference JSONB,
        remediation_status VARCHAR(50) DEFAULT 'pending',
        is_resolved BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ NOT NULL
    )""",
    "CREATE INDEX IF NOT EXISTS idx_drift_items_scan_id ON drift_items(scan_id)",
    "CREATE INDEX IF NOT EXISTS idx_drift_items_severity ON drift_items(severity)",
    "CREATE INDEX IF NOT EXISTS idx_drift_items_status ON drift_items(remediation_status)",
    "CREATE INDEX IF NOT EXISTS idx_drift_items_resource_id ON drift_items(resource_id)",
    """CREATE TABLE IF NOT EXISTS remediation_requests (
        id VARCHAR(36) PRIMARY KEY,
        drift_id VARCHAR(36) NOT NULL REFERENCES drift_items(id) ON DELETE CASCADE,
        action_type VARCHAR(100) NOT NULL,
        status VARCHAR(50) NOT NULL DEFAULT 'pending',
        dry_run BOOLEAN DEFAULT FALSE,
        approval_status VARCHAR(50) DEFAULT 'pending',
        approved_by VARCHAR(255),
        approved_at TIMESTAMPTZ,
        executed_at TIMESTAMPTZ,
        result JSONB,
        error TEXT,
        rolled_back_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ NOT NULL
    )""",
    "CREATE INDEX IF NOT EXISTS idx_remediation_drift_id ON remediation_requests(drift_id)",
    "CREATE INDEX IF NOT EXISTS idx_remediation_status ON remediation_requests(status)",
    """CREATE TABLE IF NOT EXISTS compliance_rules (
        id VARCHAR(36) PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        severity VARCHAR(50) NOT NULL,
        category VARCHAR(100) NOT NULL,
        resource_types TEXT[] NOT NULL,
        rules JSONB NOT NULL,
        enabled BOOLEAN DEFAULT TRUE,
        created_at TIMESTAMPTZ NOT NULL,
        updated_at TIMESTAMPTZ
    )""",
    """CREATE TABLE IF NOT EXISTS reports (
        id VARCHAR(36) PRIMARY KEY,
        scan_id VARCHAR(36) NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
        cloud_provider VARCHAR(50) NOT NULL,
        total_resources INT DEFAULT 0,
        drifted_resources INT DEFAULT 0,
        drift_percentage FLOAT DEFAULT 0,
        by_severity JSONB,
        by_category JSONB,
        cost_impact FLOAT DEFAULT 0,
        compliance_status VARCHAR(50),
        recommendations TEXT[],
        created_at TIMESTAMPTZ NOT NULL
    )""",
    "CREATE INDEX IF NOT EXISTS idx_reports_scan_id ON reports(scan_id)",
    """CREATE TABLE IF NOT EXISTS audit_logs (
        id VARCHAR(36) PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        action VARCHAR(255) NOT NULL,
        resource_type VARCHAR(100) NOT NULL,
        resource_id VARCHAR(255),
        details JSONB,
        timestamp TIMESTAMPTZ NOT NULL,
        ip_address VARCHAR(45)
    )""",
    "CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)",
    "CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp)",
    "CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id)",
]


def init_db(database_url: str) -> psycopg2.pool.ThreadedConnectionPool:
    """Initialize the connection pool, mirroring the Go backend's sql.DB pool settings."""
    global _pool
    _pool = psycopg2.pool.ThreadedConnectionPool(minconn=5, maxconn=25, dsn=database_url)
    conn = _pool.getconn()
    try:
        conn.close()
    finally:
        _pool.putconn(conn, close=True)
    return _pool


def migrate(pool: psycopg2.pool.ThreadedConnectionPool) -> None:
    conn = pool.getconn()
    try:
        with conn.cursor() as cur:
            for statement in MIGRATIONS:
                cur.execute(statement)
        conn.commit()
    finally:
        pool.putconn(conn)


def get_pool() -> psycopg2.pool.ThreadedConnectionPool:
    if _pool is None:
        raise RuntimeError("Database pool has not been initialized")
    return _pool


def get_conn():
    return get_pool().getconn()


def put_conn(conn) -> None:
    get_pool().putconn(conn)
