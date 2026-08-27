# Coherence - Infrastructure State Drift Detection & Auto-Remediation

![Coherence Banner](https://via.placeholder.com/1200x200?text=Coherence+DevOps+Platform)

**Coherence** is an enterprise-grade infrastructure state management platform that continuously detects and remediates drift between your Infrastructure-as-Code (IaC) and actual cloud resources.

## 🎯 Problem Statement

Every organization with Infrastructure-as-Code faces **state drift**:
- Manual console changes that bypass version control
- Infrastructure modifications by ops teams
- Compliance violations and security group changes
- Cost optimizations that aren't documented
- Accidental deletions and modifications

This creates **drift**: the gap between what your code says should exist and what actually exists in your cloud.

## ✨ Solution: Coherence

Coherence automatically:
- **Detects** infrastructure changes across AWS, GCP, Azure in real-time
- **Categorizes** drift by severity (breaking, compliance, performance, cost)
- **Traces** changes to source (user, API, timestamp) via cloud audit logs
- **Suggests** fixes intelligently (reapply IaC, update code, safe auto-remediation)
- **Remediates** low-risk drifts with approval workflows
- **Reports** on compliance and state health with Slack/PagerDuty integration

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- AWS CLI (for AWS testing)

### Local Development (60 seconds)

```bash
# Clone and setup
git clone https://github.com/yourusername/coherence.git
cd coherence
docker-compose up -d

# Run migrations
./scripts/migrate.sh

# Start backend
cd backend && go run cmd/main.go

# Start frontend (in another terminal)
cd frontend && npm install && npm start

# Access dashboard
open http://localhost:3000
```

### Using the CLI

```bash
# Install
go install ./cli/coherence

# Configure
coherence config set aws-profile default
coherence config set terraform-path ./terraform

# Run a drift check
coherence scan --cloud aws --region us-east-1

# View results
coherence report --format json > drift-report.json

# Auto-remediate (with approval)
coherence remediate --severity low --auto-approve
```

## 📊 Features

### Core Detection
- ✅ Multi-cloud support (AWS, GCP, Azure)
- ✅ Real-time drift detection via cloud APIs
- ✅ Terraform, CloudFormation, Helm support
- ✅ Custom resource type plugins

### Intelligence
- ✅ Drift severity classification
- ✅ Audit trail correlation (who changed what, when)
- ✅ Change impact analysis
- ✅ Cost impact estimation
- ✅ Compliance rule checking

### Remediation
- ✅ Intelligent fix suggestions
- ✅ Safe auto-remediation for low-risk changes
- ✅ Approval workflows
- ✅ Rollback capabilities
- ✅ Dry-run preview

### Integration & Observability
- ✅ Slack/PagerDuty alerting
- ✅ GitHub webhook integration
- ✅ REST API (OpenAPI spec)
- ✅ Prometheus metrics
- ✅ Audit logging

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Web Dashboard (React)                 │
└───────────────────┬─────────────────────────────────────┘
                    │
        ┌───────────┴──────────────┐
        │                          │
┌───────▼────────┐      ┌──────────▼─────────┐
│  REST API      │      │  WebSocket (Events)│
│  (Go + Gin)    │      │                    │
└───────┬────────┘      └────────────────────┘
        │
┌───────▼──────────────────────────────────────────┐
│         Coherence Engine (Go)                    │
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

## 📁 Project Structure

```
coherence/
├── backend/
│   ├── cmd/
│   │   └── main.go                 # Server entry point
│   ├── internal/
│   │   ├── api/                    # REST API handlers
│   │   ├── drift/                  # Drift detection logic
│   │   ├── providers/              # Cloud provider adapters
│   │   ├── iac/                    # IaC parsing
│   │   ├── remediation/            # Auto-fix logic
│   │   ├── models/                 # Data models
│   │   └── config/                 # Configuration
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   └── App.tsx
│   ├── package.json
│   └── Dockerfile
├── cli/
│   ├── cmd/
│   │   ├── scan.go
│   │   ├── remediate.go
│   │   ├── report.go
│   │   └── main.go
│   └── go.mod
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

## 🔌 API Examples

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

## 🧪 Testing

```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test

# Integration tests
./scripts/integration-test.sh

# Load testing
./scripts/load-test.sh
```

## 📦 Deployment

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

## 🔐 Security

- End-to-end encryption for cloud credentials
- RBAC with teams and permissions
- Audit logging for all operations
- Secrets stored in encrypted vault
- Regular security scanning (Trivy, Snyk)

## 📊 Use Cases

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

## 🤝 Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md)

## 📄 License

MIT

## 📧 Support

- Docs: https://coherence-docs.dev
- Issues: GitHub Issues
- Community: Slack channel

---

**Built for DevOps engineers who want their infrastructure to tell the truth.**
