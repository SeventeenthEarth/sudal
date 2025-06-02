@grpc @health @positive @concurrency
Feature: gRPC Health Service
  As a gRPC client
  I want to check the health of the service via gRPC protocol
  So that I can ensure the gRPC service is available and responding

  Background:
    Given the server is running
    And the gRPC client is connected

  @positive
  Scenario: Basic gRPC health check
    When I make a gRPC health check request
    Then the gRPC response should indicate serving status
    And the gRPC response should not be empty

  @positive
  Scenario: gRPC health check with metadata
    When I make a gRPC health check request with metadata
    Then the gRPC response should indicate serving status
    And the gRPC response should not be empty
    And the gRPC response should contain metadata

  @positive
  Scenario: gRPC health check using protocol call
    When I call the "gRPC" health endpoint
    Then the status should be healthy

  @positive
  Scenario Outline: gRPC health check with different timeouts
    Given the gRPC client is connected with timeout "<timeout>"
    When I make a gRPC health check request
    Then the gRPC response should indicate serving status
    And the gRPC response should not be empty

    Examples:
      | timeout |
      | 1s      |
      | 5s      |
      | 10s     |

  @positive
  Scenario Outline: gRPC concurrent requests
    When I make <num_requests> concurrent "gRPC" health requests
    Then all concurrent "gRPC" requests should succeed

    Examples:
      | num_requests |
      | 5            |
      | 10           |
      | 20           |


