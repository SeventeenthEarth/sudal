"""
Common step definitions for E2E tests.
"""

import json
import requests
import concurrent.futures
from pytest_bdd import given, when, then, parsers


@given("the server is running on port 8080")
def server_is_running(wait_for_server):
    """Ensure server is running and accessible."""
    # The wait_for_server fixture already handles this
    pass


@when(parsers.parse('I make a GET request to "{endpoint}"'))
def make_get_request(endpoint, server_url, http_client, context):
    """Make a GET request to the specified endpoint."""
    url = f"{server_url}{endpoint}"
    try:
        response = http_client.get(url)
        context["response"] = response
        context["error"] = None
    except requests.exceptions.RequestException as e:
        context["response"] = None
        context["error"] = e


@then(parsers.parse("the HTTP response status should be {status_code:d}"))
def check_status_code(status_code, context):
    """Check that the HTTP response has the expected status code."""
    response = context.get("response")
    assert response is not None, "No response received"
    assert (
        response.status_code == status_code
    ), f"Expected status code {status_code}, got {response.status_code}"


@then(parsers.parse('the JSON response should contain status "{expected_status}"'))
def check_json_status(expected_status, context):
    """Check that the JSON response contains the expected status."""
    response = context.get("response")
    assert response is not None, "No response received"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    assert "status" in json_data, "Response does not contain 'status' field"
    assert (
        json_data["status"] == expected_status
    ), f"Expected status '{expected_status}', got '{json_data['status']}'"


@when(parsers.parse('I make {num_requests:d} concurrent requests to "{endpoint}"'))
def make_concurrent_endpoint_requests(num_requests, endpoint, server_url, context):
    """Make multiple concurrent requests to a specific endpoint."""
    url = f"{server_url}{endpoint}"

    def make_request():
        try:
            response = requests.get(url, timeout=5)
            return response, None
        except requests.exceptions.RequestException as e:
            return None, e

    # Make concurrent requests
    with concurrent.futures.ThreadPoolExecutor(max_workers=num_requests) as executor:
        futures = [executor.submit(make_request) for _ in range(num_requests)]
        results = [
            future.result() for future in concurrent.futures.as_completed(futures)
        ]

    context["concurrent_results"] = results
