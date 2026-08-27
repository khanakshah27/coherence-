#!/bin/bash

# Coherence Setup Script
# This script sets up the Coherence environment for local development

set -e

echo "Setting up Coherence..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cat > .env << EOF
# Coherence Configuration
ENVIRONMENT=development
LOG_LEVEL=debug
PORT=8080

# Database
DATABASE_URL=postgresql://coherence:coherence@localhost:5432/coherence

# Redis
REDIS_URL=redis://localhost:6379

# AWS Configuration (optional)
AWS_REGION=us-east-1
AWS_PROFILE=default

# GCP Configuration (optional)
GCP_PROJECT_ID=

# Azure Configuration (optional)
AZURE_SUBSCRIPTION_ID=
AZURE_TENANT_ID=
AZURE_CLIENT_ID=
AZURE_CLIENT_SECRET=
EOF
    echo "✅ .env file created"
fi

# Start Docker Compose services
echo "🐳 Starting Docker Compose services..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 5

# The backend runs its database migrations automatically on startup.

echo ""
echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Dashboard & API: http://localhost:8080"
echo "  2. Prometheus: http://localhost:9090"
echo "  3. Grafana: http://localhost:3001 (admin/admin)"
echo ""
echo "Documentation: https://coherence-docs.dev"
echo ""
