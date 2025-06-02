@grpc @user @positive @negative @concurrency
Feature: gRPC User Service
  As a gRPC client
  I want to manage user profiles via gRPC protocol
  So that I can register users, retrieve profiles, and update user information

  Background:
    Given the server is running
    And the gRPC user client is connected

  @positive
  Scenario: User registration with valid data should succeed
    Given I have valid user registration data
    When I register a user with valid data
    Then the user registration should succeed
    And the response should contain a valid user ID

  @positive
  Scenario: Get user profile with valid user ID should succeed
    Given an existing user is registered
    When I get the user profile
    Then the user profile should be retrieved

  @positive
  Scenario: Update user profile with valid data should succeed
    Given an existing user is registered
    When I update the user profile with display name "Updated Name"
    Then the user profile update should succeed

  @positive
  Scenario Outline: User registration with different auth providers should succeed
    Given I have valid user registration data
    When I register a user with valid data
    Then the user registration should succeed
    And the response should contain a valid user ID

    Examples:
      | auth_provider |
      | google        |
      | apple         |
      | facebook      |

  @positive
  Scenario: gRPC protocol should work correctly
    Given I have valid user registration data
    When I register a user with valid data
    Then the user registration should succeed

  @positive
  Scenario: gRPC-Web protocol should work correctly
    Given the gRPC-Web user client is connected
    And I have valid user registration data
    When I register a user with valid data
    Then the user registration should succeed



  @negative
  Scenario: User registration with duplicate Firebase UID should fail
    Given an existing user is registered
    When I register a user with the same Firebase UID
    Then the user registration should fail with AlreadyExists error

  @negative
  Scenario: User registration with empty Firebase UID should fail
    Given I have invalid user registration data with empty Firebase UID
    When I register a user with empty Firebase UID
    Then the user registration should fail with InvalidArgument error

  @negative
  Scenario: Get user profile with invalid user ID should fail
    When I get the user profile with invalid ID
    Then the user registration should fail with InvalidArgument error

  @negative
  Scenario: Get user profile with non-existent user ID should fail
    When I get the user profile with non-existent ID
    Then the user profile retrieval should fail with NotFound error

  @positive
  Scenario Outline: Concurrent user registrations
    When I make <num_requests> concurrent user registrations
    Then all concurrent user registrations should succeed

    Examples:
      | num_requests |
      | 3            |
      | 5            |
      | 10           |
