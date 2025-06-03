# /test/e2e/features/user/user_auth.feature
@user_auth
Feature: User Authentication and Profile Management
  As a new user, I want to register and manage my profile using Firebase authentication.

  Background:
    Given the gRPC server is running
    And I have a new random email and a secure password

  @positive @auth
  Scenario: Successful registration and profile retrieval for a new user
    When I sign up with Firebase and register with the service
    Then the registration should be successful
    And when I get my user profile
    Then the user profile should contain my registration details

  @positive @auth
  Scenario: Attempting to register with an already existing Firebase account
    Given I am already signed up and registered
    When I attempt to register again with the same credentials
    Then the registration should be successful and not create a duplicate user

  @negative @auth
  Scenario: Accessing a protected resource with an invalid token
    When I attempt to get my user profile with an "invalid-token"
    Then the request should fail with an unauthenticated error

  @negative @auth
  Scenario: Accessing a protected resource without a token
    When I attempt to get my user profile without a token
    Then the request should fail with an unauthenticated error
