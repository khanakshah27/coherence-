# Coherence API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
Include JWT token in the Authorization header:
```
Authorization: Bearer YOUR_JWT_TOKEN
```

---

## Endpoints

### Scans

#### Create a New Scan
```http
POST /scans
Content-Type: application/json

{
  "cloud_provider": "aws",
  "regions": ["us-east-1", "eu-west-1"],
  "resource_types": ["ec2", "s3", "rds"]
}
```

**Response:**
```json
{
  "id": "scan_123",
  "cloud_provider": "aws",
  "regions": ["us-east-1", "eu-west-1"],
  "resource_types": ["ec2", "s3", "rds"],
  "status": "pending",
  "drift_count": 0,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### List All Scans
```http
GET /scans
```

**Query Parameters:**
- `limit`: Max results (default: 100)
- `offset`: Pagination offset (default: 0)
- `status`: Filter by status (pending, running, completed, failed)

**Response:**
```json
{
  "scans": [
    {
      "id": "scan_123",
      "cloud_provider": "aws",
      "status": "completed",
      "drift_count": 42,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 156
}
```

#### Get Scan Details
```http
GET /scans/{scan_id}
```

#### Retry a Scan
```http
POST /scans/{scan_id}/retry
```

---

### Drift Items

#### List Drifts
```http
GET /drifts
```

**Query Parameters:**
- `scan_id`: Filter by scan
- `severity`: Filter by severity (critical, high, medium, low, info)
- `category`: Filter by category (breaking, compliance, performance, cost)
- `is_resolved`: Filter by resolution status

**Response:**
```json
{
  "drifts": [
    {
      "id": "drift_456",
      "scan_id": "scan_123",
      "resource_id": "i-0123456789abcdef0",
      "resource_type": "ec2",
      "cloud_provider": "aws",
      "severity": "high",
      "category": "configuration",
      "title": "EC2 instance has security group mismatch",
      "description": "Security group configuration differs from IaC",
      "expected_state": { /* ... */ },
      "actual_state": { /* ... */ },
      "is_resolved": false,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 167
}
```

#### Get Drift Details
```http
GET /drifts/{drift_id}
```

#### Resolve Drift
```http
POST /drifts/{drift_id}/resolve
```

#### Bulk Resolve Drifts
```http
POST /drifts/bulk-resolve
Content-Type: application/json

{
  "drift_ids": ["drift_1", "drift_2", "drift_3"]
}
```

---

### Remediation

#### Request Remediation
```http
POST /remediations
Content-Type: application/json

{
  "drift_id": "drift_456",
  "action_type": "apply_iac",
  "dry_run": true
}
```

**Response:**
```json
{
  "id": "remediation_789",
  "drift_id": "drift_456",
  "action_type": "apply_iac",
  "status": "pending",
  "approval_status": "pending",
  "dry_run": true,
  "created_at": "2024-01-15T10:35:00Z"
}
```

#### List Remediations
```http
GET /remediations
```

**Query Parameters:**
- `status`: Filter by status
- `approval_status`: pending, approved, rejected

#### Approve Remediation
```http
POST /remediations/{remediation_id}/approve
Content-Type: application/json

{
  "comment": "Approved by DevOps team"
}
```

#### Execute Remediation
```http
POST /remediations/{remediation_id}/execute
```

**Response:**
```json
{
  "id": "remediation_789",
  "status": "executing",
  "executed_at": "2024-01-15T10:40:00Z"
}
```

#### Rollback Remediation
```http
POST /remediations/{remediation_id}/rollback
```

---

### Reports

#### Generate Report
```http
POST /reports/generate
Content-Type: application/json

{
  "scan_id": "scan_123",
  "format": "json"
}
```

#### Get Report
```http
GET /reports/{report_id}
```

#### Export Report
```http
GET /reports/{report_id}/export?format=pdf
```

**Supported formats:** json, pdf, csv, html

---

### Compliance

#### List Compliance Rules
```http
GET /compliance/rules
```

#### Get Compliance Status
```http
GET /compliance/status?scan_id={scan_id}
```

---

### Health & Metrics

#### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 3600,
  "database": "connected"
}
```

#### Prometheus Metrics
```http
GET /metrics
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request parameters",
  "details": {
    "cloud_provider": "unsupported provider"
  }
}
```

### 401 Unauthorized
```json
{
  "error": "Authentication required",
  "message": "Invalid or missing JWT token"
}
```

### 403 Forbidden
```json
{
  "error": "Insufficient permissions",
  "message": "User does not have access to this resource"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found",
  "resource_id": "scan_123"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "request_id": "req_abc123"
}
```

---

## Rate Limiting

- **Requests per minute**: 1000
- **Headers returned**:
  - `X-RateLimit-Limit`: Request limit
  - `X-RateLimit-Remaining`: Requests remaining
  - `X-RateLimit-Reset`: Unix timestamp of rate limit reset

---

## Pagination

List endpoints support pagination with:
- `limit`: Number of items (1-100, default: 20)
- `offset`: Starting position (default: 0)

Example:
```http
GET /scans?limit=50&offset=100
```

---

## Sorting

Use `sort` parameter with format: `field:asc|desc`

Example:
```http
GET /drifts?sort=created_at:desc
```

---

## Webhooks

Register webhooks for events:

```http
POST /webhooks
Content-Type: application/json

{
  "url": "https://your-domain.com/webhook",
  "events": ["scan.completed", "drift.detected", "remediation.failed"],
  "secret": "webhook_secret_key"
}
```

**Events:**
- `scan.started`
- `scan.completed`
- `scan.failed`
- `drift.detected`
- `drift.resolved`
- `remediation.requested`
- `remediation.approved`
- `remediation.executed`
- `remediation.failed`
- `remediation.rolled_back`

---

## Examples

### Complete Workflow Example

1. Start a scan
```bash
curl -X POST http://localhost:8080/api/v1/scans \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cloud_provider": "aws",
    "regions": ["us-east-1"],
    "resource_types": ["ec2", "s3"]
  }'
```

2. Get drifts from scan
```bash
curl http://localhost:8080/api/v1/drifts?scan_id=scan_123 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

3. Request remediation
```bash
curl -X POST http://localhost:8080/api/v1/remediations \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "drift_id": "drift_456",
    "action_type": "apply_iac",
    "dry_run": true
  }'
```

4. Approve and execute
```bash
curl -X POST http://localhost:8080/api/v1/remediations/remediation_789/approve \
  -H "Authorization: Bearer YOUR_TOKEN"

curl -X POST http://localhost:8080/api/v1/remediations/remediation_789/execute \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## SDK Usage

### JavaScript/TypeScript
```typescript
import { CoherenceClient } from 'coherence-sdk';

const client = new CoherenceClient({
  baseURL: 'http://localhost:8080/api/v1',
  token: 'YOUR_JWT_TOKEN'
});

const scans = await client.scans.list();
const drift = await client.drifts.get('drift_123');
```

### Python
```python
from coherence import Client

client = Client(
    base_url='http://localhost:8080/api/v1',
    token='YOUR_JWT_TOKEN'
)

scans = client.scans.list()
drift = client.drifts.get('drift_123')
```
