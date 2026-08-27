from app.compliance import Checker, summary

def test_s3_bucket_without_encryption_fails_rule():
    checker = Checker()
    results = checker.run_checks({"arn:aws:s3:::bucket": {"encryption": ""}})
    s3_001 = [r for r in results if r.rule_id == "s3-001"][0]
    assert s3_001.passed is False

def test_s3_bucket_with_encryption_passes_rule():
    checker = Checker()
    results = checker.run_checks({"arn:aws:s3:::bucket": {"encryption": "AES256"}})
    s3_001 = [r for r in results if r.rule_id == "s3-001"][0]
    assert s3_001.passed is True

def test_rds_multi_az_not_required_outside_production():
    checker = Checker()
    results = checker.run_checks(
        {"db-1": {"tags": {"Environment": "staging"}, "multi_az": False}}
    )
    rds_002 = [r for r in results if r.rule_id == "rds-002"][0]
    assert rds_002.passed is True

def test_rds_multi_az_required_in_production():
    checker = Checker()
    results = checker.run_checks(
        {"db-1": {"tags": {"Environment": "production"}, "multi_az": False}}
    )
    rds_002 = [r for r in results if r.rule_id == "rds-002"][0]
    assert rds_002.passed is False


def test_summary_computes_score():
    checker = Checker()
    results = checker.run_checks({"arn:aws:s3:::bucket": {"encryption": "AES256"}})
    stats = summary(results)
    assert stats["total"] == len(results)
    assert stats["passed"] + stats["failed"] == stats["total"]
    assert 0 <= stats["score"] <= 100
