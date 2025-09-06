@grpc @user @positive @negative @concurrency
Feature: gRPC User Service Protocol and Operations Testing
  As a gRPC client
  I want to test gRPC protocol functionality and advanced operations
  So that I can ensure proper gRPC/gRPC-Web support and service operations

  Background:
    Given the server is running
    And the gRPC user client is connected

  @positive @protocol
  Scenario: gRPC protocol should work correctly
    Given I have valid user registration data
    When I register a user with valid data
    Then the user registration should succeed
    And the response should contain a valid user ID

  @positive @protocol
  Scenario: gRPC-Web protocol should work correctly
    Given the gRPC-Web user client is connected
    And I have valid user registration data
    When I register a user with valid data
    Then the user registration should succeed
    And the response should contain a valid user ID

  @positive @crud
  Scenario: Get user profile with valid user ID should succeed
    Given an existing user is registered
    When I get the user profile
    Then the user profile should be retrieved

  @positive @crud
  Scenario: Update user profile with valid data should succeed
    Given an existing user is registered
    When I update the user profile with display name "Updated Name"
    Then the user profile update should succeed

  @negative @validation
  Scenario: User registration with empty Firebase UID should fail
    Given I have invalid user registration data with empty Firebase UID
    When I register a user with empty Firebase UID
    Then the user registration should fail with InvalidArgument error

  @negative @validation
  Scenario: Get user profile with invalid user ID should fail
    When I get the user profile with invalid ID
    Then the user profile retrieval should fail with PermissionDenied error

  @negative @validation
  Scenario: Get user profile with non-existent user ID should fail
    When I get the user profile with non-existent ID
    Then the user profile retrieval should fail with PermissionDenied error

  @concurrency @firebase_rate_limit
  Scenario Outline: Concurrent user registrations
    When I make <num_requests> concurrent user registrations
    Then all concurrent user registrations should succeed

    Examples:
      | num_requests |
      | 2            |
