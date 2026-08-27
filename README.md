# Coherence - Infrastructure State Drift Detection & Auto-Remediation

**Coherence** is an enterprise-grade infrastructure state management platform that continuously detects and remediates drift between your Infrastructure-as-Code (IaC) and actual cloud resources.

## Problem Statement

Every organization with Infrastructure-as-Code faces **state drift**:
- Manual console changes that bypass version control
- Infrastructure modifications by ops teams
- Compliance violations and security group changes
- Cost optimizations that aren't documented
- Accidental deletions and modifications

This creates **drift**: the gap between what your code says should exist and what actually exists in your cloud.

## Solution: Coherence

Coherence automatically:
- **Detects** infrastructure changes across AWS, GCP, Azure in real-time
- **Categorizes** drift by severity (breaking, compliance, performance, cost)
- **Traces** changes to source (user, API, timestamp) via cloud audit logs
- **Suggests** fixes intelligently (reapply IaC, update code, safe auto-remediation)
- **Remediates** low-risk drifts with approval workflows
- **Reports** on compliance and state health with Slack/PagerDuty integration

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Python 3.11+
- PostgreSQL 14+
- AWS CLI (for AWS testing)

### Local Development (60 seconds)

```bash
# Clone and setup
git clone https://github.com/yourusername/coherence.git
cd coherence
docker-compose up -d

# Or run the backend directly (migrations run automatically on startup)
cd backend
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8080

# Access the dashboard
open http://localhost:8080
```

### Using the CLI

```bash
# Install
cd cli && pip install -r ../backend/requirements.txt

# Configure
python coherence.py config set aws-profile default
python coherence.py config set terraform-path ./terraform

# Run a drift check
python coherence.py scan --cloud aws --region us-east-1

# View results
python coherence.py report --format json > drift-report.json

# Auto-remediate (with approval)
python coherence.py remediate --severity low --auto-approve
```

## Features

### Core Detection
- Multi-cloud support (AWS, GCP, Azure)
- Real-time drift detection via cloud APIs
- Terraform, CloudFormation, Helm support
- Custom resource type plugins

### Intelligence
- Drift severity classification
- Audit trail correlation (who changed what, when)
- Change impact analysis
- Cost impact estimation
- Compliance rule checking

### Remediation
- Intelligent fix suggestions
- Safe auto-remediation for low-risk changes
- Approval workflows
- Rollback capabilities
- Dry-run preview

### Integration & Observability
- Slack/PagerDuty alerting
- GitHub webhook integration
- REST API (OpenAPI spec)
- Prometheus metrics
- Audit logging

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│         Web Dashboard (server-rendered, Jinja2)          │
└───────────────────┬─────────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────────┐
│              REST API (FastAPI + uvicorn)                 │
└───────────────────┬─────────────────────────────────────┘
        │
┌───────▼──────────────────────────────────────────┐
│         Coherence Engine (Python)                │
├─────────────────────────────────────────────────┤
│ • Drift Detection Service                       │
│ • Cloud Provider Adapters (AWS/GCP/Azure)       │
│ • IaC Parser (Terraform, CF, Helm)              │
│ • Drift Analyzer & Classifier                   │
│ • Remediation Engine                            │
│ • Audit & Compliance Checker                    │
└───────┬──────────────────────────────────────────┘
        │
        ├─────────────┬────────────┬──────────────┐
        │             │            │              │
┌───────▼──┐  ┌──────▼───┐  ┌────▼──────┐  ┌───▼────────┐
│ PostgreSQL│  │ Redis    │  │AWS APIs   │  │ GCP/Azure  │
│  (State)  │  │ (Cache)  │  │           │  │   APIs     │
└──────────┘  └──────────┘  └───────────┘  └────────────┘
```

## Project Structure

```
coherence/
├── backend/
│   ├── app/
│   │   ├── main.py                 # FastAPI app entry point
│   │   ├── api.py                  # REST API routes
│   │   ├── web.py                  # Server-rendered dashboard routes
│   │   ├── drift.py                # Drift detection logic
│   │   ├── providers.py            # Cloud provider adapters
│   │   ├── remediation.py          # Auto-fix logic
│   │   ├── compliance.py           # Compliance rule checker
│   │   ├── alerts.py               # Slack/PagerDuty notifiers
│   │   ├── models.py               # Data models
│   │   ├── config.py               # Configuration
│   │   ├── database.py             # DB connection & migrations
│   │   ├── templates/              # Jinja2 dashboard templates
│   │   └── static/                 # CSS
│   ├── tests/
│   ├── requirements.txt
│   └── Dockerfile
├── cli/
│   ├── coherence.py                # CLI entry point
│   └── tests/
├── database/
│   ├── migrations/
│   └── schema.sql
├── docker-compose.yml
├── deployments/
│   ├── k8s/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── ingress.yaml
│   └── terraform/
│       └── main.tf
├── examples/
│   ├── terraform-drift/
│   ├── aws-config/
│   └── compliance-rules/
└── docs/
    ├── ARCHITECTURE.md
    ├── API.md
    └── SETUP.md
```

## API Examples

### Trigger Drift Scan
```bash
curl -X POST http://localhost:8080/api/v1/scans \
  -H "Content-Type: application/json" \
  -d '{
    "cloud_provider": "aws",
    "regions": ["us-east-1", "eu-west-1"],
    "resource_types": ["ec2", "s3", "rds"]
  }'
```

### Get Drift Report
```bash
curl http://localhost:8080/api/v1/reports/latest \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Remediate Drift
```bash
curl -X POST http://localhost:8080/api/v1/remediate \
  -H "Content-Type: application/json" \
  -d '{
    "drift_id": "drift_123",
    "action": "apply_iac",
    "dry_run": true
  }'
```

## Testing

```bash
# Backend tests
cd backend && pytest -v

# CLI tests
cd cli && pytest -v

# Integration tests
./scripts/integration-test.sh

# Load testing
./scripts/load-test.sh
```

## Deployment

### Docker Compose (Development)
```bash
docker-compose up -d
```

### Kubernetes (Production)
```bash
kubectl apply -f deployments/k8s/
```

### Terraform (AWS)
```bash
cd deployments/terraform
terraform init
terraform apply
```

## Security

- End-to-end encryption for cloud credentials
- RBAC with teams and permissions
- Audit logging for all operations
- Secrets stored in encrypted vault
- Regular security scanning (Trivy, Snyk)

## Use Cases

### Platform Teams
Maintain infrastructure compliance without manual audits

### DevOps Engineers
Catch breaking changes before they hit production

### Finance/FinOps
Detect unexpected resource creation and optimize costs

### Security Teams
Ensure security groups and IAM policies match compliance rules

### Dev Teams
Understand infrastructure changes that affect their services

## Contributing

Contributions welcome!

## License

MIT

## Support

- Docs: https://coherence-docs.dev
- Issues: GitHub Issues
- Community: Slack channel

---

**Built for DevOps engineers who want their infrastructure to tell the truth.**
