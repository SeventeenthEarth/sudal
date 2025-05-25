# E2E Tests

This directory contains End-to-End (E2E) tests for the Social Quiz Platform backend, implemented using Python with pytest and pytest-bdd.

## Overview

The E2E tests verify the complete functionality of the service by making actual HTTP requests to a running server instance. These tests use Behavior Driven Development (BDD) approach with Gherkin feature files.

## Test Structure

```
test/e2e/
├── features/                           # Gherkin feature files
│   ├── connect/                        # Connect-Go protocol tests
│   │   └── health_service.feature      # Health service via Connect-Go
│   └── rest/                           # REST API tests
│       ├── monitoring.feature          # Basic monitoring endpoints
│       └── database_health.feature     # Database health monitoring
├── test_connect_health_service.py      # Connect-Go health service tests
├── test_rest_monitoring.py             # REST monitoring endpoint tests
├── test_rest_database_health.py        # REST database health tests
├── conftest.py                         # Pytest configuration and all step definitions
├── pytest.ini                         # Pytest settings
├── requirements.txt                    # Python dependencies
├── run_tests.sh                       # Test runner script
└── README.md                          # This file
```

## Prerequisites

1. **Python Virtual Environment**: A virtual environment should be set up at the project root (`/venv`)
2. **Running Server**: The server must be running on the specified port (default: 8080)

## Setup

1. **Install Dependencies**:
   ```bash
   source venv/bin/activate
   pip install -r test/e2e/requirements.txt
   ```

2. **Start the Server**:
   ```bash
   # Option 1: Using Make
   make run

   # Option 2: Using Docker Compose
   docker-compose up
   ```

## Code Formatting

The project uses [Black](https://black.readthedocs.io/) for Python code formatting to ensure consistent code style.

### Format Python Code

```bash
# Format all Python code in E2E tests
make fmt-python

# Or run directly
./test/e2e/format_code.sh
```

### Check Code Formatting

```bash
# Check if Python code is properly formatted
make fmt-python-check

# Or run directly
./test/e2e/format_code.sh --check
```

### Black Configuration

Black is configured in `test/e2e/pyproject.toml` with the following settings:
- Line length: 88 characters
- Target Python versions: 3.8+
- Excludes common directories like `.venv`, `.pytest_cache`, etc.

## Running Tests

### Using the Test Runner Script (Recommended)

```bash
# Run all E2E tests
./test/e2e/run_tests.sh

# Run with custom server port
SERVER_PORT=9090 ./test/e2e/run_tests.sh
```

### Using pytest Directly

```bash
# Activate virtual environment
source venv/bin/activate

# Change to E2E test directory
cd test/e2e

# Run all tests
pytest -v

# Run specific test file
pytest test_health_service.py -v

# Run specific scenario
pytest -k "Health check using Connect-Go client" -v
```

## Test Scenarios

### Connect-Go Protocol Tests (`features/connect/health_service.feature`)

1. **Health check using Connect-Go client**: Tests the health endpoint using Connect-Go protocol
2. **Health check using HTTP/JSON over Connect-Go**: Tests the health endpoint using HTTP/JSON over Connect-Go
3. **Invalid content type rejection for Connect-Go endpoint**: Verifies server rejects requests with invalid content types
4. **Non-existent Connect-Go method returns 404**: Tests error handling for non-existent Connect-Go methods
5. **Multiple concurrent Connect-Go health requests**: Tests server performance under concurrent Connect-Go load
6. **Connect-Go health service error handling**: Tests proper Connect-Go headers and error handling

### REST Monitoring Tests (`features/rest/monitoring.feature`)

1. **Server ping endpoint responds correctly**: Tests the `/ping` endpoint for basic server availability
2. **Basic health endpoint responds correctly**: Tests the `/healthz` endpoint for simple health checks
3. **Health endpoint provides simple status**: Verifies `/healthz` provides lightweight monitoring data
4. **Multiple monitoring endpoints are accessible**: Tests accessibility of multiple monitoring endpoints

### REST Database Health Tests (`features/rest/database_health.feature`)

1. **Database health endpoint responds correctly**: Tests the `/health/database` endpoint
2. **Database health endpoint includes timestamp**: Verifies timestamp inclusion in database health responses
3. **Database connection pool status is healthy**: Tests database connection pool health validation
4. **Database health provides detailed connection metrics**: Tests detailed database connection statistics
5. **Database health endpoint performance**: Tests performance under concurrent database health requests

## Environment Variables

- `SERVER_PORT`: Port where the server is running (default: 8080)

## Test Features

- **BDD Style**: Tests are written in Gherkin syntax for better readability
- **Concurrent Testing**: Includes tests for concurrent request handling
- **Error Handling**: Tests various error scenarios
- **Protocol Support**: Tests both Connect-Go and HTTP/JSON protocols
- **Automatic Server Detection**: Automatically detects if server is running
- **Detailed Reporting**: Provides detailed test output and error messages

## Troubleshooting

### Server Not Running
```
Error: Server is not running on port 8080
```
**Solution**: Start the server using `make run` or `docker-compose up`

### Connection Refused
```
Failed to connect to server at http://localhost:8080
```
**Solution**:
1. Check if the server is running
2. Verify the correct port is being used
3. Check firewall settings

### Import Errors
```
ModuleNotFoundError: No module named 'pytest_bdd'
```
**Solution**: Install dependencies using `pip install -r test/e2e/requirements.txt`

## Adding New Tests

1. **Add Feature File**: Create or update `.feature` files in the `features/` directory
2. **Implement Steps**: Add step definitions in the `step_definitions/` directory
3. **Create Test Runner**: Create a new `test_*.py` file that loads the scenarios
4. **Update Documentation**: Update this README if needed

## Integration with CI/CD

The tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions step
- name: Run E2E Tests
  run: |
    # Start server in background
    make run &

    # Wait for server to be ready
    sleep 10

    # Run E2E tests
    ./test/e2e/run_tests.sh
```
