Feature: Database Health REST API
  As a DevOps engineer or database administrator
  I want to monitor database connection status and performance
  So that I can ensure database availability and troubleshoot connection issues

  Background:
    Given the server is running on port 8080

  Scenario: Database health endpoint responds correctly
    When I make a GET request to "/health/database"
    Then the HTTP response status should be 200
    And the JSON response should contain status "healthy"
    And the JSON response should contain database information
    And the JSON response should contain connection statistics

  Scenario: Database health endpoint includes timestamp
    When I make a GET request to "/health/database"
    Then the HTTP response status should be 200
    And the JSON response should contain a timestamp field

  Scenario: Database connection pool status is healthy
    When I make a GET request to "/health/database"
    Then the HTTP response status should be 200
    And the database connection pool should be healthy
    And the connection statistics should be valid

  Scenario: Database health provides detailed connection metrics
    When I make a GET request to "/health/database"
    Then the HTTP response status should be 200
    And the JSON response should contain connection statistics
    And the connection statistics should include max_open_connections
    And the connection statistics should include current usage metrics

  Scenario: Database health endpoint performance
    When I make 5 concurrent requests to "/health/database"
    Then all database health requests should succeed
    And all responses should contain valid connection statistics
