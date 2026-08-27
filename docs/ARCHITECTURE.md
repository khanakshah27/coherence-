# Coherence Architecture

## Overview

Coherence is a distributed system designed to detect and remediate infrastructure state drift across multi-cloud environments. The architecture follows a modular, scalable design with clear separation of concerns.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend Layer (React)                   │
│  - Web Dashboard (localhost:3000)                               │
│  - Real-time notifications via WebSocket                        │
│  - Charts and visualizations (Recharts)                         │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                      API Layer (Gin/Go)                          │
│  - REST API (localhost:8080/api/v1)                             │
│  - WebSocket Server for real-time events                        │
│  - Request validation and authentication                        │
│  - CORS handling and rate limiting                              │
└──────────────────────────────┬──────────────────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                    Business Logic Layer (Go)                     │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ Core Services:                                          │    │
│  │  - Drift Detection Service                             │    │
│  │  - IaC Parser (Terraform, CloudFormation, Helm)        │    │
│  │  - Remediation Engine                                  │    │
│  │  - Compliance Checker                                  │    │
│  │  - Audit & Logging                                     │    │
│  └─────────────────────────────────────────────────────────┘    │
└──────────────────────────────┬──────────────────────────────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
┌───────▼────────┐   ┌────────▼────────┐   ┌────────▼─────────┐
│ Cloud Provider │   │  IaC State Store │   │  Audit & Events  │
│   Adapters     │   │  (Git/Terraform) │   │  (Event Stream)  │
│                │   │                  │   │                  │
│ - AWS Adapter  │   │ - Terraform State│   │ - CloudTrail     │
│ - GCP Adapter  │   │ - CF Templates   │   │ - Audit Logs     │
│ - Azure Adapt. │   │ - Helm Charts    │   │ - Event Logs     │
└────────┬────────┘   └────────┬────────┘   └────────┬─────────┘
         │                     │                     │
         └─────────────────────┼─────────────────────┘
                               │
┌──────────────────────────────▼──────────────────────────────────┐
│                      Data Layer (PostgreSQL)                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │    Scans     │  │  Drift Items │  │ Remediation  │          │
│  │    Table     │  │    Table     │  │  Requests    │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Compliance  │  │   Reports    │  │ Audit Logs   │          │
│  │   Rules      │  │   Table      │  │   Table      │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└──────────────────────────────────────────────────────────────────┘
                               │
                               ▼
                         ┌──────────┐
                         │  Cache   │
                         │  (Redis) │
                         └──────────┘
```

## Core Components

### 1. Frontend (React)
- **Technology**: React 18, TypeScript, TailwindCSS, Recharts
- **Responsibilities**:
  - Real-time dashboard with drift visualization
  - Scan management and triggering
  - Drift details and remediation workflows
  - Compliance reporting
- **Features**:
  - WebSocket connection for live updates
  - Chart-based data visualization
  - Responsive design
  - Accessibility compliance

### 2. API Server (Gin + Go)
- **Technology**: Go 1.21, Gin Web Framework
- **Responsibilities**:
  - REST API endpoints for all operations
  - Request validation and authentication
  - Rate limiting and caching
  - WebSocket server for real-time events
  - Metrics exposition (Prometheus)
- **Key Features**:
  - OpenAPI 3.0 specification
  - JWT-based authentication
  - RBAC authorization
  - Comprehensive logging

### 3. Drift Detection Service
- **Responsibilities**:
  - Fetch actual cloud state via cloud provider APIs
  - Parse and understand expected state (IaC)
  - Detect differences and categorize drift
  - Enrich with audit information
  - Calculate impact analysis
- **Algorithm**:
  ```
  1. Fetch cloud resources (AWS EC2, S3, RDS, etc.)
  2. Parse IaC (Terraform, CloudFormation)
  3. Compare state (resource by resource)
  4. Detect changes (create, update, delete)
  5. Correlate with audit logs
  6. Classify severity and category
  7. Suggest fixes
  8. Store results in database
  ```

### 4. Cloud Provider Adapters
- **AWS Adapter**:
  - EC2 (instances, security groups, networks)
  - S3 (buckets, lifecycle policies, encryption)
  - RDS (databases, parameters, backups)
  - IAM (roles, policies, users)
  - CloudTrail (audit events)
  - Cost Explorer (cost analysis)

- **GCP Adapter**:
  - Compute Engine (instances, templates)
  - Cloud Storage (buckets, policies)
  - Cloud SQL (databases)
  - Cloud Audit Logs

- **Azure Adapter**:
  - Virtual Machines
  - Azure Storage
  - Azure SQL
  - Activity Logs

### 5. IaC Parser
- **Supported Formats**:
  - Terraform (HCL)
  - AWS CloudFormation (JSON/YAML)
  - Helm Charts (YAML)
- **Capabilities**:
  - Parse and validate configuration files
  - Extract resource definitions
  - Resolve variables and interpolations
  - Handle module dependencies

### 6. Remediation Engine
- **Functions**:
  - Apply IaC changes (Terraform apply, CF update)
  - Safe auto-remediation (low-risk changes)
  - Approval workflow management
  - Dry-run preview
  - Rollback capability
  - Change history tracking

### 7. Database Layer (PostgreSQL)
- **Schema**:
  - `scans`: Scan execution records
  - `drift_items`: Detected drift items
  - `remediation_requests`: Remediation history
  - `compliance_rules`: Compliance definitions
  - `reports`: Generated reports
  - `audit_logs`: Audit trail

- **Indexes**:
  - On `scan_id` for fast drift lookup
  - On `severity` and `status` for filtering
  - On `resource_id` for resource tracking
  - On `timestamp` for time-range queries

### 8. Cache Layer (Redis)
- **Purpose**: Reduce database load
- **Cached Items**:
  - Recent scans
  - Drift summaries
  - Compliance status
  - Session data
- **TTL**: 5 minutes default, 1 hour for static data

## Data Flow

### Scan Workflow
```
User triggers scan
    ↓
API creates Scan record (status: pending)
    ↓
Background worker picks up scan
    ↓
Fetch actual cloud state (AWS/GCP/Azure)
    ↓
Fetch expected state (Terraform/CF/Helm)
    ↓
Compare resources → Detect drift
    ↓
Enrich with audit information
    ↓
Save DriftItems to database
    ↓
Update Scan status (completed) + drift_count
    ↓
Send WebSocket event to frontend
    ↓
Generate Report
```

### Remediation Workflow
```
User requests remediation
    ↓
Create RemediationRequest (status: pending, approval_status: pending)
    ↓
Send notification to approvers
    ↓
Approver reviews and approves/rejects
    ↓
If approved:
  - Perform dry-run to preview changes
  - Show expected changes to user
  - User confirms execution
  - Execute remediation (Terraform apply, etc.)
  - Update DriftItem (remediation_status: success)
  - Generate audit log
    ↓
WebSocket event sent to frontend
```

## Deployment Architecture

### Local Development (Docker Compose)
```
Your Machine
├── PostgreSQL (port 5432)
├── Redis (port 6379)
├── Backend API (port 8080)
├── Frontend (port 3000)
├── Prometheus (port 9090)
└── Grafana (port 3001)
```

### Kubernetes Production
```
K8s Cluster (coherence namespace)
├── coherence-backend Deployment (3 replicas)
│  └── Pod with backend container
├── coherence-frontend Deployment (2 replicas)
│  └── Pod with frontend container
├── coherence-backend Service (ClusterIP)
├── coherence-frontend Service (ClusterIP)
├── PostgreSQL StatefulSet
├── Redis StatefulSet
└── Ingress (external access)
```

## High Availability & Scaling

### Backend Scaling
- **Horizontal Scaling**: Multiple replicas behind LoadBalancer
- **Database Connection Pooling**: 25 max open, 5 idle
- **In-Memory Caching**: Redis for hot data
- **Async Processing**: Background workers for scans

### Database Scaling
- **Read Replicas**: For read-heavy operations
- **Connection Pooling**: PgBouncer for connection management
- **Partitioning**: By scan_id for large drift_items tables
- **Archival**: Old scans moved to archive tables

### Frontend Optimization
- **CDN**: Static assets cached at edge
- **Code Splitting**: Lazy-loaded React components
- **Service Worker**: Offline capability
- **Compression**: GZIP/Brotli enabled

## Security Architecture

### Authentication & Authorization
- **JWT Tokens**: Stateless authentication
- **RBAC**: Role-based access control
  - Admin
  - CloudEngineer
  - DevOpsEngineer
  - ReadOnly
- **Scopes**: Fine-grained permissions per operation

### Data Protection
- **Encryption at Rest**: PostgreSQL pgcrypto
- **Encryption in Transit**: TLS 1.3
- **Secrets Management**: Kubernetes Secrets (production), .env (dev)
- **Audit Logging**: All operations logged

### Network Security
- **Network Policies**: K8s NetworkPolicy
- **Service Mesh**: (Optional) Istio for advanced routing
- **API Gateway**: For rate limiting and WAF

## Monitoring & Observability

### Metrics (Prometheus)
- API request latency (histogram)
- API error rates (counter)
- Database connection pool stats
- Drift detection duration
- Cache hit/miss ratio
- Queue size for remediation jobs

### Logging
- Structured logging (JSON)
- Centralized log aggregation (ELK/Loki)
- Log levels: DEBUG, INFO, WARN, ERROR, FATAL
- Request ID tracking for debugging

### Alerting
- Alert Manager integration
- Critical drift detected → PagerDuty
- Compliance violation → Slack
- Remediation failed → Email
- High API error rate → OpsGenie

### Tracing
- OpenTelemetry integration
- Distributed request tracing
- Jaeger/Datadog backend
- Service dependency map

## Disaster Recovery

### Backup Strategy
- **Database**: Daily full backups + continuous WAL archiving
- **Configuration**: IaC stored in Git with versioning
- **Audit Logs**: Immutable archive in S3

### Recovery Procedures
- **RTO** (Recovery Time Objective): 1 hour
- **RPO** (Recovery Point Objective): 15 minutes
- **Failover**: Automated to standby in different AZ

### Disaster Recovery Testing
- Monthly failover drills
- Backup restoration testing
- Documentation and runbooks maintained

## Technology Stack Summary

| Component | Technology | Version |
|-----------|-----------|---------|
| API Server | Go | 1.21 |
| Web Framework | Gin | 1.9+ |
| Frontend | React | 18.2+ |
| Database | PostgreSQL | 14+ |
| Cache | Redis | 7.0+ |
| Container Runtime | Docker | 20.10+ |
| Orchestration | Kubernetes | 1.24+ |
| Monitoring | Prometheus | 2.40+ |
| Visualization | Grafana | 9.0+ |

## Performance Characteristics

| Operation | Latency | Throughput |
|-----------|---------|-----------|
| List scans | < 100ms | 1000 req/s |
| Get drift details | < 50ms | 5000 req/s |
| Trigger scan | < 500ms | 100 scans/min |
| Detect drift | 5-30s | Depends on resources |
| Apply remediation | 30s-5min | Depends on change |

## Future Enhancements

1. **ML-Powered Anomaly Detection**
   - Detect unusual patterns in drift
   - Predict future drift based on history

2. **Cost Optimization**
   - Identify underutilized resources
   - Suggest cost-saving configurations

3. **Advanced Compliance**
   - CIS benchmarks integration
   - Custom compliance rule builder
   - Compliance dashboard

4. **GitOps Integration**
   - Auto-sync with Git repositories
   - Pull request automation for fixes

5. **Multi-Tenancy**
   - Support for multiple organizations
   - Isolated data per tenant
