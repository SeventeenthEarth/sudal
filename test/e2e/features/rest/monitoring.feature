Feature: Server Monitoring REST API
  As a DevOps engineer or monitoring system
  I want to access simple monitoring endpoints
  So that I can check server status and availability

  Background:
    Given the server is running on port 8080

  Scenario: Server ping endpoint responds correctly
    When I make a GET request to "/ping"
    Then the HTTP response status should be 200
    And the JSON response should contain status "ok"

  Scenario: Basic health endpoint responds correctly
    When I make a GET request to "/healthz"
    Then the HTTP response status should be 200
    And the JSON response should contain status "healthy"

  Scenario: Health endpoint provides simple status
    When I make a GET request to "/healthz"
    Then the HTTP response status should be 200
    And the JSON response should contain status "healthy"
    And the response should be lightweight for monitoring

  Scenario: Multiple monitoring endpoints are accessible
    When I make a GET request to "/ping"
    Then the HTTP response status should be 200
    When I make a GET request to "/healthz"
    Then the HTTP response status should be 200
