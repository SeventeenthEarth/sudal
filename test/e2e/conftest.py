"""
Pytest configuration and fixtures for E2E tests.
"""

import os
import time
import requests
import pytest
from typing import Dict, Any


@pytest.fixture(scope="session")
def server_url():
    """Get server URL from environment or use default."""
    server_port = os.getenv("SERVER_PORT", "8080")
    # Clean up port value in case it contains comments
    if isinstance(server_port, str) and "#" in server_port:
        server_port = server_port.split("#")[0].strip()
    return f"http://localhost:{server_port}"


@pytest.fixture(scope="session")
def wait_for_server(server_url):
    """Wait for server to be ready before running tests."""
    max_retries = 5
    retry_delay = 1
    server_port = os.getenv("SERVER_PORT", "8080")

    for i in range(max_retries):
        try:
            response = requests.get(f"{server_url}/ping", timeout=2)
            if response.status_code == 200:
                return True
        except requests.exceptions.RequestException:
            pass

        if i < max_retries - 1:
            time.sleep(retry_delay)

    pytest.fail(
        f"Failed to connect to server at {server_url}. "
        f"Make sure the server is running in Docker on port {server_port}"
    )


@pytest.fixture
def http_client():
    """HTTP client for making requests."""
    return requests.Session()


@pytest.fixture
def context():
    """Test context to store data between steps."""
    return {}
