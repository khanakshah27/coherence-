from app.drift import DriftService


def test_extract_resource_type_ec2():
    assert DriftService._extract_resource_type("i-1234567890abcdef0") == "ec2"


def test_extract_resource_type_unknown():
    assert DriftService._extract_resource_type("arn:aws:s3:::my-bucket") == "unknown"


def test_is_critical_resource_matches_production_prefix():
    assert DriftService._is_critical_resource("production-db-1") is True


def test_is_critical_resource_false_for_other_names():
    assert DriftService._is_critical_resource("i-1234567890abcdef0") is False


def test_states_equal_for_identical_dicts():
    a = {"instance_type": "t3.medium"}
    b = {"instance_type": "t3.medium"}
    assert DriftService._states_equal(a, b) is True


def test_states_equal_false_for_different_dicts():
    a = {"instance_type": "t3.medium"}
    b = {"instance_type": "t3.large"}
    assert DriftService._states_equal(a, b) is False
