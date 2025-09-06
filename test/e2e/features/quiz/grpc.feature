@grpc @quiz @scaffolding
Feature: gRPC QuizService Scaffolding Testing
  As a gRPC client
  I want to test QuizService basic connectivity
  So that I can ensure proper scaffolding setup

  Background:
    Given the server is running
    And the gRPC or gRPC-Web quiz client is connected

  @positive @public @grpc
  Scenario: ListQuizSets should return unimplemented
    When I call ListQuizSets with empty request
    Then I should receive CodeUnimplemented error
    And the error message should contain "not implemented"

  @positive @public @grpc
  Scenario: GetQuizSet should return unimplemented
    When I call GetQuizSet with quiz_set_id "1"
    Then I should receive CodeUnimplemented error
    And the error message should contain "not implemented"

  @positive @protected @auth_required @grpc
  Scenario: SubmitQuizResult requires authentication
    Given I have no authentication token
    When I call SubmitQuizResult with quiz_set_id "1"
    Then I should receive Unauthenticated error

  @positive @protected @auth_required @grpc
  Scenario: GetUserQuizHistory requires authentication
    Given I have no authentication token
    When I call GetUserQuizHistory with empty request
    Then I should receive Unauthenticated error

  @negative @protocol_filter @connect
  Scenario: Connect(JSON) protocol should be blocked by gRPC-only filter
    When I call ListQuizSets via Connect JSON protocol
    Then I should receive HTTP status 404
    And this scenario may fail until Quiz path is added to GetGRPCOnlyPaths

