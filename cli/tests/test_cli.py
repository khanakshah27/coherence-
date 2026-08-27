import subprocess
import sys
from pathlib import Path

CLI = Path(__file__).resolve().parent.parent / "coherence.py"


def run(*args: str) -> subprocess.CompletedProcess:
    return subprocess.run(
        [sys.executable, str(CLI), *args],
        capture_output=True,
        text=True,
        timeout=10,
    )


def test_version_flag():
    result = run("--version")
    assert result.returncode == 0
    assert "coherence" in result.stdout


def test_no_command_prints_help_and_exits_nonzero():
    result = run()
    assert result.returncode == 1
    assert "usage" in result.stdout.lower()


def test_scan_command_reports_provider_and_region():
    result = run("scan", "--cloud", "aws", "--region", "us-east-1")
    assert result.returncode == 0
    assert "aws" in result.stdout
    assert "us-east-1" in result.stdout


def test_report_command_default_format():
    result = run("report")
    assert result.returncode == 0
    assert "json" in result.stdout
