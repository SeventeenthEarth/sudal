"""
Database health step definitions for E2E tests.
"""

import json
import re
from pytest_bdd import then


@then("the JSON response should contain database information")
def check_database_information(context):
    """Check that the JSON response contains database information."""
    response = context.get("response")
    assert response is not None, "No response received"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    assert "database" in json_data, "Response does not contain 'database' field"
    database_info = json_data["database"]

    assert isinstance(database_info, dict), "Database field should be an object"
    assert "status" in database_info, "Database info does not contain 'status' field"
    assert "message" in database_info, "Database info does not contain 'message' field"


@then("the JSON response should contain connection statistics")
def check_connection_statistics(context):
    """Check that the JSON response contains connection statistics."""
    response = context.get("response")
    assert response is not None, "No response received"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    assert "database" in json_data, "Response does not contain 'database' field"
    database_info = json_data["database"]

    assert "stats" in database_info, "Database info does not contain 'stats' field"
    stats = database_info["stats"]

    assert isinstance(stats, dict), "Stats field should be an object"

    # Check for expected statistics fields
    expected_stats = [
        "max_open_connections",
        "open_connections",
        "in_use",
        "idle",
        "wait_count",
        "wait_duration",
        "max_idle_closed",
        "max_lifetime_closed",
    ]

    for stat_field in expected_stats:
        assert stat_field in stats, f"Stats does not contain '{stat_field}' field"
        assert isinstance(
            stats[stat_field], (int, float)
        ), f"Stat '{stat_field}' should be a number, got {type(stats[stat_field])}"


@then("the JSON response should contain a timestamp field")
def check_timestamp_field(context):
    """Check that the JSON response contains a timestamp field."""
    response = context.get("response")
    assert response is not None, "No response received"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    assert "timestamp" in json_data, "Response does not contain 'timestamp' field"
    timestamp = json_data["timestamp"]

    assert isinstance(timestamp, str), "Timestamp should be a string"
    assert len(timestamp) > 0, "Timestamp should not be empty"

    # Basic ISO 8601 format check (YYYY-MM-DDTHH:MM:SSZ)
    iso_pattern = r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$"
    assert re.match(
        iso_pattern, timestamp
    ), f"Timestamp should be in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ), got: {timestamp}"


@then("the database connection pool should be healthy")
def check_connection_pool_healthy(context):
    """Check that the database connection pool is healthy."""
    response = context.get("response")
    assert response is not None, "No response received"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    assert "database" in json_data, "Response does not contain 'database' field"
    database_info = json_data["database"]

    assert "status" in database_info, "Database info does not contain 'status' field"
    assert (
        database_info["status"] == "healthy"
    ), f"Expected database status to be 'healthy', got '{database_info['status']}'"

    assert "stats" in database_info, "Database info does not contain 'stats' field"
    stats = database_info["stats"]

    # Check that we have at least some open connections configured
    assert (
        stats["max_open_connections"] > 0
    ), "Max open connections should be greater than 0"

    # Check that open connections doesn't exceed max
    assert (
        stats["open_connections"] <= stats["max_open_connections"]
    ), "Open connections should not exceed max open connections"


@then("the connection statistics should be valid")
def check_connection_statistics_valid(context):
    """Check that the connection statistics are valid and consistent."""
    response = context.get("response")
    assert response is not None, "No response received"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    assert "database" in json_data, "Response does not contain 'database' field"
    database_info = json_data["database"]

    assert "stats" in database_info, "Database info does not contain 'stats' field"
    stats = database_info["stats"]

    # Validate that statistics are consistent
    open_conns = stats["open_connections"]
    in_use = stats["in_use"]
    idle = stats["idle"]

    # Open connections should equal in_use + idle
    assert (
        open_conns == in_use + idle
    ), f"Open connections ({open_conns}) should equal in_use ({in_use}) + idle ({idle})"

    # All counts should be non-negative
    for field in [
        "open_connections",
        "in_use",
        "idle",
        "wait_count",
        "max_idle_closed",
        "max_lifetime_closed",
    ]:
        assert stats[field] >= 0, f"{field} should be non-negative, got {stats[field]}"

    # Wait duration should be non-negative
    assert (
        stats["wait_duration"] >= 0
    ), f"Wait duration should be non-negative, got {stats['wait_duration']}"


@then("the connection statistics should include max_open_connections")
def check_max_open_connections_stat(context):
    """Check that connection statistics include max_open_connections."""
    response = context.get("response")
    assert response is not None, "No response received"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    assert "database" in json_data, "Response does not contain 'database' field"
    database_info = json_data["database"]

    assert "stats" in database_info, "Database info does not contain 'stats' field"
    stats = database_info["stats"]

    assert (
        "max_open_connections" in stats
    ), "Stats does not contain 'max_open_connections'"
    assert isinstance(
        stats["max_open_connections"], (int, float)
    ), "max_open_connections should be a number"
    assert (
        stats["max_open_connections"] > 0
    ), "max_open_connections should be greater than 0"


@then("the connection statistics should include current usage metrics")
def check_current_usage_metrics(context):
    """Check that connection statistics include current usage metrics."""
    response = context.get("response")
    assert response is not None, "No response received"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    assert "database" in json_data, "Response does not contain 'database' field"
    database_info = json_data["database"]

    assert "stats" in database_info, "Database info does not contain 'stats' field"
    stats = database_info["stats"]

    # Check for current usage metrics
    usage_metrics = ["open_connections", "in_use", "idle"]
    for metric in usage_metrics:
        assert metric in stats, f"Stats does not contain '{metric}'"
        assert isinstance(stats[metric], (int, float)), f"{metric} should be a number"
        assert stats[metric] >= 0, f"{metric} should be non-negative"


@then("all database health requests should succeed")
def check_all_database_health_requests_succeed(context):
    """Check that all concurrent database health requests succeeded."""
    results = context.get("concurrent_results", [])
    assert results, "No concurrent results found"

    for response, error in results:
        assert error is None, f"Request failed with error: {error}"
        assert response is not None, "Response is None"
        assert (
            response.status_code == 200
        ), f"Expected status 200, got {response.status_code}"


@then("all responses should contain valid connection statistics")
def check_all_responses_contain_valid_stats(context):
    """Check that all concurrent responses contain valid connection statistics."""
    results = context.get("concurrent_results", [])
    assert results, "No concurrent results found"

    for response, error in results:
        assert error is None, f"Request failed with error: {error}"
        assert response is not None, "Response is None"

        try:
            json_data = response.json()
        except json.JSONDecodeError:
            assert False, "Response is not valid JSON"

        assert "database" in json_data, "Response does not contain 'database' field"
        database_info = json_data["database"]

        assert "stats" in database_info, "Database info does not contain 'stats' field"
        stats = database_info["stats"]

        # Basic validation of stats structure
        required_fields = ["max_open_connections", "open_connections", "in_use", "idle"]
        for field in required_fields:
            assert field in stats, f"Stats does not contain '{field}' field"
            assert isinstance(stats[field], (int, float)), f"{field} should be a number"
