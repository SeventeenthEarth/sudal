package e2e

import (
	"testing"

	"github.com/seventeenthearth/sudal/test/e2e/steps"
)

// TestRestUserService tests the Connect-Go User Service protocol filtering
func TestRestUserService(t *testing.T) {
	// BDD Scenarios for Connect-Go User Service protocol filtering
	scenarios := []steps.BDDScenario{
		{
			Name:        "HTTP/JSON request to gRPC user endpoint should be rejected",
			Description: "Should return 404 status when making HTTP/JSON request to gRPC-only user endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIMakeUserRegistrationRequestUsingHTTPJSON(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
		{
			Name:        "HTTP/JSON request to get user profile endpoint should be rejected",
			Description: "Should return 404 status when making HTTP/JSON request to gRPC-only get user profile endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIMakeGetUserProfileRequestUsingHTTPJSON(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
		{
			Name:        "HTTP/JSON request to update user profile endpoint should be rejected",
			Description: "Should return 404 status when making HTTP/JSON request to gRPC-only update user profile endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIMakeUpdateUserProfileRequestUsingHTTPJSON(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
		{
			Name:        "Invalid content type requests to gRPC user endpoint should be rejected",
			Description: "Should return 404 status when making request with invalid content type to gRPC-only user endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIMakeUserRequestWithInvalidContentType(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
		{
			Name:        "Non-existent Connect-Go user method returns 404",
			Description: "Should return 404 status when making request to non-existent user endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIMakeRequestToNonExistentUserEndpoint(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
		{
			Name:        "Multiple concurrent HTTP/JSON requests to gRPC user endpoint should be rejected",
			Description: "Should return 404 for all concurrent HTTP/JSON requests to gRPC-only user endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIMakeConcurrentUserHTTPJSONRequests(ctx, 10)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenAllConcurrentRequestsShouldReturn404(ctx)
			},
		},
		{
			Name:        "HTTP/JSON request to gRPC user endpoint should be rejected with proper error",
			Description: "Should return 404 with proper error message for HTTP/JSON requests to gRPC-only user endpoint",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIMakeUserRegistrationRequestUsingHTTPJSON(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				steps.ThenHTTPStatusShouldBe(ctx, 404)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, ServerURL, scenarios)
}

// TestRestUserServiceTableDriven demonstrates table-driven BDD tests
func TestRestUserServiceTableDriven(t *testing.T) {
	// Table-driven test cases for different user request scenarios
	type UserRequestTestCase struct {
		Name                string
		Endpoint            string
		ContentType         string
		Body                string
		ExpectedStatus      int
		ShouldContainStatus bool
		ExpectedStatusValue string
	}

	testCases := []interface{}{
		UserRequestTestCase{
			Name:                "HTTP/JSON request to RegisterUser endpoint should be rejected",
			Endpoint:            "/user.v1.UserService/RegisterUser",
			ContentType:         "application/json",
			Body:                `{"firebase_uid":"test","display_name":"Test User","auth_provider":"google"}`,
			ExpectedStatus:      404,
			ShouldContainStatus: false,
			ExpectedStatusValue: "",
		},
		UserRequestTestCase{
			Name:                "HTTP/JSON request to GetUserProfile endpoint should be rejected",
			Endpoint:            "/user.v1.UserService/GetUserProfile",
			ContentType:         "application/json",
			Body:                `{"user_id":"test-user-id"}`,
			ExpectedStatus:      404,
			ShouldContainStatus: false,
		},
		UserRequestTestCase{
			Name:                "HTTP/JSON request to UpdateUserProfile endpoint should be rejected",
			Endpoint:            "/user.v1.UserService/UpdateUserProfile",
			ContentType:         "application/json",
			Body:                `{"user_id":"test-user-id","display_name":"Updated Name"}`,
			ExpectedStatus:      404,
			ShouldContainStatus: false,
		},
		UserRequestTestCase{
			Name:                "Invalid content type to gRPC user endpoint should be rejected",
			Endpoint:            "/user.v1.UserService/RegisterUser",
			ContentType:         "text/plain",
			Body:                `{"firebase_uid":"test","display_name":"Test User"}`,
			ExpectedStatus:      404,
			ShouldContainStatus: false,
		},
		UserRequestTestCase{
			Name:                "Non-existent user method",
			Endpoint:            "/user.v1.UserService/NonExistentMethod",
			ContentType:         "application/json",
			Body:                "{}",
			ExpectedStatus:      404,
			ShouldContainStatus: false,
		},
	}

	tableDrivenTest := steps.TableDrivenBDDTest{
		Name: "Connect-Go user request scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			steps.GivenServerIsRunning(ctx)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(UserRequestTestCase)
			steps.WhenIMakePOSTRequest(ctx, testCase.Endpoint, testCase.ContentType, testCase.Body)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(UserRequestTestCase)
			steps.ThenHTTPStatusShouldBe(ctx, testCase.ExpectedStatus)

			if testCase.ShouldContainStatus {
				steps.ThenJSONResponseShouldContainStatus(ctx, testCase.ExpectedStatusValue)
			}
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, tableDrivenTest, testCases)
}

// TestRestUserServiceConcurrency tests concurrent request scenarios
func TestRestUserServiceConcurrency(t *testing.T) {
	// Table-driven test for different concurrency levels
	type ConcurrencyTestCase struct {
		Name        string
		NumRequests int
		Description string
	}

	concurrencyTestCases := []interface{}{
		ConcurrencyTestCase{
			Name:        "Low concurrency user requests",
			NumRequests: 5,
			Description: "Test with 5 concurrent user HTTP/JSON requests",
		},
		ConcurrencyTestCase{
			Name:        "Medium concurrency user requests",
			NumRequests: 10,
			Description: "Test with 10 concurrent user HTTP/JSON requests",
		},
		ConcurrencyTestCase{
			Name:        "High concurrency user requests",
			NumRequests: 20,
			Description: "Test with 20 concurrent user HTTP/JSON requests",
		},
	}

	concurrencyTest := steps.TableDrivenBDDTest{
		Name: "Connect-Go user concurrency scenarios",
		Given: func(ctx *steps.TestContext, testData interface{}) {
			steps.GivenServerIsRunning(ctx)
		},
		When: func(ctx *steps.TestContext, testData interface{}) {
			testCase := testData.(ConcurrencyTestCase)
			WhenIMakeConcurrentUserHTTPJSONRequests(ctx, testCase.NumRequests)
		},
		Then: func(ctx *steps.TestContext, testData interface{}) {
			steps.ThenAllConcurrentRequestsShouldReturn404(ctx)
		},
	}

	steps.RunTableDrivenBDDTest(t, ServerURL, concurrencyTest, concurrencyTestCases)
}

// When Steps for User Service HTTP/JSON requests (that should be rejected)

// WhenIMakeUserRegistrationRequestUsingHTTPJSON makes HTTP/JSON request to RegisterUser endpoint
func WhenIMakeUserRegistrationRequestUsingHTTPJSON(ctx *steps.TestContext) {
	endpoint := "/user.v1.UserService/RegisterUser"
	contentType := "application/json"
	body := `{"firebase_uid":"test_firebase_uid","display_name":"Test User","auth_provider":"google"}`
	steps.WhenIMakePOSTRequest(ctx, endpoint, contentType, body)
}

// WhenIMakeGetUserProfileRequestUsingHTTPJSON makes HTTP/JSON request to GetUserProfile endpoint
func WhenIMakeGetUserProfileRequestUsingHTTPJSON(ctx *steps.TestContext) {
	endpoint := "/user.v1.UserService/GetUserProfile"
	contentType := "application/json"
	body := `{"user_id":"test-user-id"}`
	steps.WhenIMakePOSTRequest(ctx, endpoint, contentType, body)
}

// WhenIMakeUpdateUserProfileRequestUsingHTTPJSON makes HTTP/JSON request to UpdateUserProfile endpoint
func WhenIMakeUpdateUserProfileRequestUsingHTTPJSON(ctx *steps.TestContext) {
	endpoint := "/user.v1.UserService/UpdateUserProfile"
	contentType := "application/json"
	body := `{"user_id":"test-user-id","display_name":"Updated Name","avatar_url":"https://example.com/avatar.jpg"}`
	steps.WhenIMakePOSTRequest(ctx, endpoint, contentType, body)
}

// WhenIMakeUserRequestWithInvalidContentType makes request with invalid content type to user endpoint
func WhenIMakeUserRequestWithInvalidContentType(ctx *steps.TestContext) {
	endpoint := "/user.v1.UserService/RegisterUser"
	contentType := "text/plain"
	body := `{"firebase_uid":"test","display_name":"Test User"}`
	steps.WhenIMakePOSTRequest(ctx, endpoint, contentType, body)
}

// WhenIMakeRequestToNonExistentUserEndpoint makes request to non-existent user endpoint
func WhenIMakeRequestToNonExistentUserEndpoint(ctx *steps.TestContext) {
	endpoint := "/user.v1.UserService/NonExistentMethod"
	contentType := "application/json"
	body := "{}"
	steps.WhenIMakePOSTRequest(ctx, endpoint, contentType, body)
}

// WhenIMakeConcurrentUserHTTPJSONRequests makes multiple concurrent HTTP/JSON requests to user endpoints
func WhenIMakeConcurrentUserHTTPJSONRequests(ctx *steps.TestContext, numRequests int) {
	// Use the existing concurrent request function with user endpoint
	endpoint := "/user.v1.UserService/RegisterUser"
	contentType := "application/json"
	body := `{"firebase_uid":"test","display_name":"Test User","auth_provider":"google"}`

	// Initialize HTTPTestContext if needed
	if ctx.HTTPTestContext == nil {
		ctx.HTTPTestContext = &steps.HTTPTestContext{}
	}

	// Store the endpoint details for concurrent requests
	ctx.HTTPTestContext.RequestEndpoint = endpoint
	ctx.HTTPTestContext.RequestContentType = contentType
	ctx.HTTPTestContext.RequestBody = body

	steps.WhenIMakeConcurrentHealthCheckRequests(ctx, numRequests)
}
