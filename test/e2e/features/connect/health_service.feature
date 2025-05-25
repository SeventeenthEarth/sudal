Feature: Health Service Connect-Go API
  As a client application
  I want to check the health status of services using Connect-Go protocol
  So that I can ensure the services are running properly and handle failures gracefully

  Background:
    Given the server is running on port 8080

  Scenario: Health check using Connect-Go client
    When I make a health check request using Connect-Go client
    Then the response should indicate SERVING status
    And the response should not be empty

  Scenario: Health check using HTTP/JSON over Connect-Go
    When I make a health check request using HTTP/JSON
    Then the HTTP response status should be 200
    And the JSON response should contain SERVING_STATUS_SERVING

  Scenario: Invalid content type rejection for Connect-Go endpoint
    When I make a health check request with invalid content type
    Then the HTTP response status should be 415
    And the server should reject the request

  Scenario: Non-existent Connect-Go method returns 404
    When I make a request to a non-existent endpoint
    Then the HTTP response status should be 404

  Scenario: Multiple concurrent Connect-Go health requests
    When I make 10 concurrent health check requests
    Then all requests should succeed
    And all responses should indicate SERVING status

  Scenario: Connect-Go health service error handling
    When I make a health check request using Connect-Go client
    Then the response should indicate SERVING status
    And the response should contain proper Connect-Go headers
