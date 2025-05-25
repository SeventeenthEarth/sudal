"""
Monitoring endpoint step definitions for E2E tests.
"""

import json
from pytest_bdd import then


@then("the response should be lightweight for monitoring")
def check_lightweight_response(context):
    """Check that the response is lightweight and suitable for monitoring."""
    response = context.get("response")
    assert response is not None, "No response received"

    # Check response size is small (good for monitoring)
    content_length = len(response.content)
    assert (
        content_length < 1000
    ), f"Response should be lightweight for monitoring, got {content_length} bytes"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    # Should have minimal fields for quick monitoring
    assert (
        len(json_data) <= 3
    ), f"Monitoring response should have minimal fields, got {len(json_data)} fields"

    # Should contain status field
    assert "status" in json_data, "Monitoring response should contain 'status' field"
