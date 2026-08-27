"""Application configuration loaded from environment variables."""
import os
from dataclasses import dataclass, field


@dataclass
class AWSConfig:
    region: str = "us-east-1"
    profile: str = "default"


@dataclass
class GCPConfig:
    project_id: str = ""
    credentials: str = ""


@dataclass
class AzureConfig:
    subscription_id: str = ""
    tenant_id: str = ""
    client_id: str = ""
    client_secret: str = ""


@dataclass
class Config:
    database_url: str = "postgresql://coherence:coherence@localhost:5432/coherence"
    redis_url: str = "redis://localhost:6379"
    environment: str = "development"
    port: str = "8080"
    log_level: str = "info"

    aws: AWSConfig = field(default_factory=AWSConfig)
    gcp: GCPConfig = field(default_factory=GCPConfig)
    azure: AzureConfig = field(default_factory=AzureConfig)


def load_config() -> Config:
    return Config(
        database_url=os.getenv("DATABASE_URL", Config.database_url),
        redis_url=os.getenv("REDIS_URL", Config.redis_url),
        environment=os.getenv("ENVIRONMENT", Config.environment),
        port=os.getenv("PORT", Config.port),
        log_level=os.getenv("LOG_LEVEL", Config.log_level),
        aws=AWSConfig(
            region=os.getenv("AWS_REGION", "us-east-1"),
            profile=os.getenv("AWS_PROFILE", "default"),
        ),
        gcp=GCPConfig(
            project_id=os.getenv("GCP_PROJECT_ID", ""),
            credentials=os.getenv("GCP_CREDENTIALS", ""),
        ),
        azure=AzureConfig(
            subscription_id=os.getenv("AZURE_SUBSCRIPTION_ID", ""),
            tenant_id=os.getenv("AZURE_TENANT_ID", ""),
            client_id=os.getenv("AZURE_CLIENT_ID", ""),
            client_secret=os.getenv("AZURE_CLIENT_SECRET", ""),
        ),
    )
