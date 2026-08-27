# Coherence Quick Start Guide

Get Coherence up and running in 5 minutes!

## Prerequisites

- **Docker** (version 20.10+)
- **Docker Compose** (version 1.29+)
- **Git**
- **AWS CLI** (optional, for AWS testing)

## Installation

### 1. Clone the Repository
```bash
git clone https://github.com/yourusername/coherence.git
cd coherence
```

### 2. Start the Development Environment
```bash
# Make the setup script executable
chmod +x scripts/setup.sh

# Run the setup script
./scripts/setup.sh
```

Or, if you prefer manual setup:
```bash
# Start all services with Docker Compose
docker-compose up -d

# Verify all services are running
docker-compose ps
```

### 3. Access the Applications

Once all containers are running, access:

- **Dashboard**: http://localhost:3000
- **API Documentation**: http://localhost:8080/swagger (coming soon)
- **Prometheus Metrics**: http://localhost:9090
- **Grafana**: http://localhost:3001 (default: admin/admin)

## First Scan

### Using the Web Dashboard

1. Navigate to http://localhost:3000
2. Click "+ New Scan" button
3. Select cloud provider (AWS, GCP, or Azure)
4. Select regions to scan
5. Click "Start Scan"
6. View results in real-time

### Using the CLI

```bash
# Configure Coherence
coherence config set cloud-provider aws
coherence config set aws-region us-east-1

# Run a scan
coherence scan --cloud aws --regions us-east-1,us-west-2

# View results
coherence report --format table
```

### Using the API

```bash
# Create a scan
curl -X POST http://localhost:8080/api/v1/scans \
  -H "Content-Type: application/json" \
  -d '{
    "cloud_provider": "aws",
    "regions": ["us-east-1"],
    "resource_types": ["ec2", "s3", "rds"]
  }'

# List scans
curl http://localhost:8080/api/v1/scans

# Get drifts
curl 'http://localhost:8080/api/v1/drifts?scan_id=YOUR_SCAN_ID'
```

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```bash
# Application
ENVIRONMENT=development
LOG_LEVEL=debug
PORT=8080

# Database
DATABASE_URL=postgresql://coherence:coherence@localhost:5432/coherence

# Redis
REDIS_URL=redis://localhost:6379

# AWS
AWS_REGION=us-east-1
AWS_PROFILE=default

# GCP (optional)
GCP_PROJECT_ID=your-project-id
GCP_CREDENTIALS=/path/to/credentials.json

# Azure (optional)
AZURE_SUBSCRIPTION_ID=your-subscription-id
AZURE_TENANT_ID=your-tenant-id
AZURE_CLIENT_ID=your-client-id
AZURE_CLIENT_SECRET=your-client-secret
```

### AWS Setup

To scan AWS resources, you need to configure AWS credentials:

```bash
# Option 1: Use AWS CLI configuration
aws configure

# Option 2: Set environment variables
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1

# Option 3: Use IAM role (for K8s deployment)
# Configure IRSA (IAM Role for Service Account) in Kubernetes
```

### Required AWS Permissions

Create an IAM policy with these permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "s3:List*",
        "s3:Get*",
        "rds:Describe*",
        "cloudtrail:LookupEvents"
      ],
      "Resource": "*"
    }
  ]
}
```

## Common Tasks

### View Logs

```bash
# Backend logs
docker-compose logs -f backend

# Frontend logs
docker-compose logs -f frontend

# All services
docker-compose logs -f
```

### Stop Services

```bash
docker-compose down
```

### Reset Everything

```bash
# Stop and remove all containers and volumes
docker-compose down -v

# Start fresh
docker-compose up -d
```

### Access Database

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U coherence -d coherence

# Common queries
SELECT COUNT(*) FROM scans;
SELECT COUNT(*) FROM drift_items;
SELECT severity, COUNT(*) FROM drift_items GROUP BY severity;
```

### View Redis Cache

```bash
# Connect to Redis
docker-compose exec redis redis-cli

# View keys
KEYS *
GET scan:123
```

## Development Workflow

### Backend Development

```bash
# Install Go dependencies
cd backend
go mod download

# Run tests
go test ./...

# Format code
go fmt ./...

# Build binary
go build -o coherence-server cmd/main.go
```

### Frontend Development

```bash
# Install dependencies
cd frontend
npm install

# Start development server
npm start

# Run tests
npm test

# Build for production
npm run build
```

### Database Migrations

```bash
# Run migrations (automatic on startup)
# OR manually via:
docker-compose exec backend ./coherence-server migrate

# Create a new migration
./scripts/create-migration.sh name_of_migration
```

## Troubleshooting

### Services won't start

```bash
# Check Docker daemon is running
docker ps

# Check logs for errors
docker-compose logs

# Rebuild images
docker-compose build --no-cache
docker-compose up -d
```

### Database connection errors

```bash
# Verify PostgreSQL is running
docker-compose ps postgres

# Check connection string in logs
docker-compose logs backend | grep DATABASE

# Test connection manually
docker-compose exec postgres psql -U coherence -d coherence -c "SELECT 1"
```

### Frontend can't connect to API

```bash
# Check if backend is running
curl http://localhost:8080/health

# Check CORS headers
curl -i -H "Origin: http://localhost:3000" http://localhost:8080/health

# Check network connectivity
docker-compose exec frontend curl http://backend:8080/health
```

### High memory usage

```bash
# Check resource usage
docker stats

# Reduce cache size or replicas
# Edit docker-compose.yml and restart
docker-compose restart
```

## Next Steps

1. **Read the Documentation**
   - [Architecture Guide](docs/ARCHITECTURE.md)
   - [API Documentation](docs/API.md)

2. **Deploy to Kubernetes**
   - [Kubernetes Setup](deployments/k8s/README.md)

3. **Integrate with Your Cloud**
   - [AWS Integration](examples/aws-config/)
   - [GCP Integration](examples/gcp-config/)

4. **Set Up Notifications**
   - Configure Slack/PagerDuty webhooks
   - Set up email alerts

5. **Enable Compliance Checking**
   - Import compliance rules
   - Configure compliance scans

## Getting Help

- **GitHub Issues**: https://github.com/yourusername/coherence/issues
- **Documentation**: https://coherence-docs.dev
- **Community Slack**: [Join our Slack](https://coherence.slack.com)
- **Email Support**: support@coherence.dev

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

---

**Happy drifting! 🚀**

If you run into issues, check the logs and feel free to open a GitHub issue.
