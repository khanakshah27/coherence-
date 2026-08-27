"""Coherence FastAPI application entrypoint."""
from __future__ import annotations

import logging

from dotenv import load_dotenv
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from app import api, database, web
from app.config import load_config
from app.drift import DriftService
from app.providers import AWSProvider, AzureProvider, GCPProvider


load_dotenv(".env")

logging.basicConfig(
    level=logging.INFO,
    format='{"time": "%(asctime)s", "level": "%(levelname)s", "message": "%(message)s"}',
)
logger = logging.getLogger("coherence")

app = FastAPI(title="Coherence", version="1.0.0")
app.mount("/static", StaticFiles(directory="app/static"), name="static")
app.include_router(api.router)
app.include_router(web.router)


@app.on_event("startup")
def on_startup() -> None:
    cfg = load_config()
    logger.info("Configuration loaded: env=%s, port=%s", cfg.environment, cfg.port)

    pool = database.init_db(cfg.database_url)
    database.migrate(pool)

    drift_service = DriftService(
        aws_provider=AWSProvider(cfg.aws),
        gcp_provider=GCPProvider(cfg.gcp),
        azure_provider=AzureProvider(cfg.azure),
    )
    api.set_drift_service(drift_service)
    web.set_drift_service(drift_service)

    logger.info("Starting Coherence server on port %s", cfg.port)


@app.get("/health")
def health():
    return {"status": "healthy", "version": "1.0.0"}
