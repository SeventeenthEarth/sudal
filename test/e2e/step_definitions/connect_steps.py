"""
Connect-Go protocol step definitions for E2E tests.
"""

import json
import requests
import concurrent.futures
from pytest_bdd import when, then, parsers


@when("I make a health check request using Connect-Go client")
def make_connect_health_request(server_url, http_client, context):
    """Make a health check request using Connect-Go protocol."""
    # Since we're using Python, we'll simulate the Connect-Go request
    # by making an HTTP POST request to the Connect endpoint
    url = f"{server_url}/health.v1.HealthService/Check"
    headers = {"Content-Type": "application/json"}
    data = "{}"

    try:
        response = http_client.post(url, data=data, headers=headers)
        context["response"] = response
        context["error"] = None
    except requests.exceptions.RequestException as e:
        context["response"] = None
        context["error"] = e


@when("I make a health check request using HTTP/JSON")
def make_http_health_request(server_url, http_client, context):
    """Make a health check request using HTTP/JSON."""
    url = f"{server_url}/health.v1.HealthService/Check"
    headers = {"Content-Type": "application/json"}
    data = "{}"

    try:
        response = http_client.post(url, data=data, headers=headers)
        context["response"] = response
        context["error"] = None
    except requests.exceptions.RequestException as e:
        context["response"] = None
        context["error"] = e


@when("I make a health check request with invalid content type")
def make_invalid_content_type_request(server_url, http_client, context):
    """Make a health check request with invalid content type."""
    url = f"{server_url}/health.v1.HealthService/Check"
    headers = {"Content-Type": "text/plain"}
    data = "{}"

    try:
        response = http_client.post(url, data=data, headers=headers)
        context["response"] = response
        context["error"] = None
    except requests.exceptions.RequestException as e:
        context["response"] = None
        context["error"] = e


@when("I make a request to a non-existent endpoint")
def make_nonexistent_endpoint_request(server_url, http_client, context):
    """Make a request to a non-existent endpoint."""
    url = f"{server_url}/health.v1.HealthService/NonExistentMethod"
    headers = {"Content-Type": "application/json"}
    data = "{}"

    try:
        response = http_client.post(url, data=data, headers=headers)
        context["response"] = response
        context["error"] = None
    except requests.exceptions.RequestException as e:
        context["response"] = None
        context["error"] = e


@when(parsers.parse("I make {num_requests:d} concurrent health check requests"))
def make_concurrent_requests(num_requests, server_url, context):
    """Make multiple concurrent health check requests."""
    url = f"{server_url}/health.v1.HealthService/Check"
    headers = {"Content-Type": "application/json"}
    data = "{}"

    def make_request():
        try:
            response = requests.post(url, data=data, headers=headers, timeout=5)
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


@then("the response should indicate SERVING status")
def check_serving_status(context):
    """Check that the response indicates SERVING status."""
    response = context.get("response")
    assert response is not None, "No response received"
    assert (
        response.status_code == 200
    ), f"Expected status 200, got {response.status_code}"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    # Check for Connect-Go style response
    assert "status" in json_data, "Response does not contain 'status' field"
    assert (
        json_data["status"] == "SERVING_STATUS_SERVING"
    ), f"Expected SERVING_STATUS_SERVING, got {json_data['status']}"


@then("the response should not be empty")
def check_response_not_empty(context):
    """Check that the response is not empty."""
    response = context.get("response")
    assert response is not None, "No response received"
    assert response.content, "Response body is empty"


@then("the JSON response should contain SERVING_STATUS_SERVING")
def check_json_serving_status(context):
    """Check that the JSON response contains SERVING_STATUS_SERVING."""
    response = context.get("response")
    assert response is not None, "No response received"

    try:
        json_data = response.json()
    except json.JSONDecodeError:
        assert False, "Response is not valid JSON"

    assert "status" in json_data, "Response does not contain 'status' field"
    assert (
        json_data["status"] == "SERVING_STATUS_SERVING"
    ), f"Expected SERVING_STATUS_SERVING, got {json_data['status']}"


@then("the server should reject the request")
def check_request_rejected(context):
    """Check that the server rejected the request."""
    response = context.get("response")
    assert response is not None, "No response received"
    # For invalid content type, we expect 415 status code
    assert (
        response.status_code == 415
    ), f"Expected status 415 for rejected request, got {response.status_code}"


@then("all requests should succeed")
def check_all_requests_succeed(context):
    """Check that all concurrent requests succeeded."""
    results = context.get("concurrent_results", [])
    assert results, "No concurrent results found"

    for response, error in results:
        assert error is None, f"Request failed with error: {error}"
        assert response is not None, "Response is None"
        assert (
            response.status_code == 200
        ), f"Expected status 200, got {response.status_code}"


@then("all responses should indicate SERVING status")
def check_all_responses_serving(context):
    """Check that all concurrent responses indicate SERVING status."""
    results = context.get("concurrent_results", [])
    assert results, "No concurrent results found"

    for response, error in results:
        assert error is None, f"Request failed with error: {error}"
        assert response is not None, "Response is None"

        try:
            json_data = response.json()
        except json.JSONDecodeError:
            assert False, "Response is not valid JSON"

        assert "status" in json_data, "Response does not contain 'status' field"
        assert (
            json_data["status"] == "SERVING_STATUS_SERVING"
        ), f"Expected SERVING_STATUS_SERVING, got {json_data['status']}"


@then("the response should contain proper Connect-Go headers")
def check_connect_go_headers(context):
    """Check that the response contains proper Connect-Go headers."""
    response = context.get("response")
    assert response is not None, "No response received"

    # Check for Connect-Go specific headers
    headers = response.headers
    content_type = headers.get("Content-Type", "")

    # Connect-Go typically uses application/json for HTTP/JSON
    assert (
        "application/json" in content_type
    ), f"Expected JSON content type for Connect-Go, got: {content_type}"
