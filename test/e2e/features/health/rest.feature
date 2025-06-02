@rest @connect @health @negative @concurrency
Feature: REST Health Endpoints
  As a monitoring system
  I want to check the health of the service via REST endpoints
  So that I can ensure the service is running properly

  Background:
    Given the server is running

  @positive
  Scenario Outline: Health endpoints respond correctly
    When I make a GET request to "<endpoint>"
    Then the HTTP status should be <status>
    And the response should contain status "<expected_status>"

    Examples:
      | endpoint             | status | expected_status |
      | /api/ping           | 200    | ok              |
      | /api/healthz        | 200    | healthy         |
      | /api/health/database| 200    | healthy         |

  @positive
  Scenario: Database health endpoint provides detailed information
    When I make a GET request to "/api/health/database"
    Then the HTTP status should be 200
    And the response should contain status "healthy"
    And the response should contain field "database" with value "connected"

  @positive
  Scenario: Health endpoint provides lightweight response for monitoring
    When I make a GET request to "/api/healthz"
    Then the HTTP status should be 200
    And the response should contain status "healthy"
    And the response should be lightweight for monitoring
    And the content type should be "application/json"

  @positive
  Scenario: Server ping endpoint provides lightweight response for monitoring
    When I make a GET request to "/api/ping"
    Then the HTTP status should be 200
    And the response should contain status "ok"
    And the response should be lightweight for monitoring
    And the content type should be "application/json"

  @negative
  Scenario: Connect-Go HTTP/JSON request to gRPC endpoint should be rejected
    When I make a Connect-Go health request
    Then the HTTP status should be 404

  @negative
  Scenario Outline: Invalid requests to Connect-Go endpoints
    When I make a POST request to "<endpoint>" with content type "<content_type>" and body "<body>"
    Then the HTTP status should be <expected_status>

    Examples:
      | endpoint                             | content_type     | body | expected_status |
      | /health.v1.HealthService/Check       | application/json | {}   | 404             |
      | /health.v1.HealthService/Check       | text/plain       | {}   | 404             |
      | /health.v1.HealthService/NonExistent | application/json | {}   | 404             |

  @positive @connect
  Scenario: Basic Connect-Go gRPC-Web health check
    Given the Connect-Go client is configured with "grpc-web" protocol
    When I make a Connect-Go health check request
    Then the Connect-Go response should indicate serving status
    And the Connect-Go response should not be empty

  @positive @connect
  Scenario Outline: Connect-Go protocol comparison with gRPC-only restriction
    Given the Connect-Go client is configured with "<protocol>" protocol
    When I make a Connect-Go health check request
    Then the Connect-Go request should <result>

    Examples:
      | protocol  | result |
      | grpc-web  | succeed |
      | http      | fail    |

  @positive @connect
  Scenario Outline: Connect-Go protocol timeout handling
    Given the Connect-Go client is configured with "<protocol>" protocol and <timeout>ms timeout
    When I make a Connect-Go health check request
    Then the Connect-Go request should <result>

    Examples:
      | protocol  | timeout | result  |
      | grpc-web  | 5000    | succeed |
      | grpc-web  | 100     | succeed |
      | http      | 5000    | fail    |

  @positive @connect
  Scenario Outline: Connect-Go concurrent requests
    Given the Connect-Go client is configured with "<protocol>" protocol
    When I make <num_requests> concurrent Connect-Go health check requests
    Then all Connect-Go requests should <result>

    Examples:
      | protocol  | num_requests | result  |
      | grpc-web  | 5           | succeed |
      | grpc-web  | 10          | succeed |
      | grpc-web  | 30          | succeed |

  @positive
  Scenario Outline: REST concurrent requests
    When I make <num_requests> concurrent "REST" health requests
    Then all concurrent "REST" requests should succeed

    Examples:
      | num_requests |
      | 5            |
      | 10           |
      | 20           |

  @positive
  Scenario Outline: Monitoring endpoint concurrent requests
    When I make <num_requests> concurrent GET requests to "<endpoint>"
    Then all requests should succeed
    And all responses should contain status "<expected_status>"

    Examples:
      | endpoint     | num_requests | expected_status |
      | /api/ping    | 5           | ok             |
      | /api/healthz | 5           | healthy        |
      | /api/ping    | 10          | ok             |
      | /api/healthz | 10          | healthy        |
