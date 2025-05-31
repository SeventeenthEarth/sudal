package e2e

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"

	userv1 "github.com/seventeenthearth/sudal/gen/go/user/v1"
	"github.com/seventeenthearth/sudal/gen/go/user/v1/userv1connect"
	"github.com/seventeenthearth/sudal/test/e2e/steps"
)

// TestGRPCUserService tests the gRPC User Service functionality using native gRPC client
func TestGRPCUserService(t *testing.T) {
	// BDD Scenarios for gRPC User Service
	scenarios := []steps.BDDScenario{
		{
			Name:        "User registration with valid data should succeed",
			Description: "Should successfully register a new user with valid Firebase UID and display name",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				GivenUserServiceClientWithGRPC(ctx, ServerURL)
			},
			When: func(ctx *steps.TestContext) {
				WhenIRegisterUserWithValidData(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				ThenUserRegistrationShouldSucceed(ctx)
				ThenResponseShouldContainValidUserID(ctx)
			},
		},
		{
			Name:        "User registration with duplicate Firebase UID should fail",
			Description: "Should return AlreadyExists error when registering user with existing Firebase UID",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				GivenUserServiceClientWithGRPC(ctx, ServerURL)
				GivenExistingUserIsRegistered(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIRegisterUserWithSameFirebaseUID(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				ThenUserRegistrationShouldFailWithAlreadyExistsError(ctx)
			},
		},
		{
			Name:        "User registration with empty Firebase UID should fail",
			Description: "Should return InvalidArgument error when Firebase UID is empty",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				GivenUserServiceClientWithGRPC(ctx, ServerURL)
			},
			When: func(ctx *steps.TestContext) {
				WhenIRegisterUserWithEmptyFirebaseUID(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				ThenUserRegistrationShouldFailWithInvalidArgumentError(ctx)
			},
		},
		{
			Name:        "Get user profile with valid user ID should succeed",
			Description: "Should successfully retrieve user profile for existing user",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				GivenUserServiceClientWithGRPC(ctx, ServerURL)
				GivenExistingUserIsRegistered(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIGetUserProfileWithValidID(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				ThenUserProfileShouldBeRetrieved(ctx)
				ThenProfileShouldContainCorrectData(ctx)
			},
		},
		{
			Name:        "Get user profile with non-existent user ID should fail",
			Description: "Should return NotFound error when user ID does not exist",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				GivenUserServiceClientWithGRPC(ctx, ServerURL)
			},
			When: func(ctx *steps.TestContext) {
				WhenIGetUserProfileWithNonExistentID(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				ThenUserProfileRetrievalShouldFailWithNotFoundError(ctx)
			},
		},
		{
			Name:        "Update user profile with valid data should succeed",
			Description: "Should successfully update user profile with new display name and avatar URL",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				GivenUserServiceClientWithGRPC(ctx, ServerURL)
				GivenExistingUserIsRegistered(ctx)
			},
			When: func(ctx *steps.TestContext) {
				WhenIUpdateUserProfileWithValidData(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				ThenUserProfileUpdateShouldSucceed(ctx)
			},
		},
		{
			Name:        "gRPC protocol should work correctly",
			Description: "Should handle gRPC requests properly via HTTP/2",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				GivenUserServiceClientWithGRPC(ctx, ServerURL)
			},
			When: func(ctx *steps.TestContext) {
				WhenIRegisterUserWithValidData(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				ThenUserRegistrationShouldSucceed(ctx)
			},
		},
		{
			Name:        "gRPC-Web protocol should work correctly",
			Description: "Should handle gRPC-Web requests properly",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				GivenUserServiceClientWithGRPCWeb(ctx, ServerURL)
			},
			When: func(ctx *steps.TestContext) {
				WhenIRegisterUserWithValidData(ctx)
			},
			Then: func(ctx *steps.TestContext) {
				ThenUserRegistrationShouldSucceed(ctx)
			},
		},
	}

	// Run all BDD scenarios
	steps.RunBDDScenarios(t, ServerURL, scenarios)
}

// TestGRPCUserServiceConcurrency tests concurrent user operations
func TestGRPCUserServiceConcurrency(t *testing.T) {
	scenarios := []steps.BDDScenario{
		{
			Name:        "Concurrent user registrations should all succeed",
			Description: "Should handle multiple concurrent user registrations without conflicts",
			Given: func(ctx *steps.TestContext) {
				steps.GivenServerIsRunning(ctx)
				GivenUserServiceClientWithGRPC(ctx, ServerURL)
			},
			When: func(ctx *steps.TestContext) {
				WhenIMakeConcurrentUserRegistrations(ctx, 5)
			},
			Then: func(ctx *steps.TestContext) {
				ThenAllConcurrentUserRegistrationsShouldSucceed(ctx)
			},
		},
	}

	steps.RunBDDScenarios(t, ServerURL, scenarios)
}

// User Service Test Context
type UserServiceTestContext struct {
	Client            userv1connect.UserServiceClient
	RegisterRequest   *userv1.RegisterUserRequest
	RegisterResponse  *connect.Response[userv1.RegisterUserResponse]
	ProfileRequest    *userv1.GetUserProfileRequest
	ProfileResponse   *connect.Response[userv1.UserProfile]
	UpdateRequest     *userv1.UpdateUserProfileRequest
	UpdateResponse    *connect.Response[userv1.UpdateUserProfileResponse]
	LastError         error
	CreatedUserID     string
	TestFirebaseUID   string
	TestDisplayName   string
	ConcurrentResults []ConcurrentUserResult
}

type ConcurrentUserResult struct {
	Response *connect.Response[userv1.RegisterUserResponse]
	Error    error
}

// Helper function to get or create user service test context
func getUserServiceContext(ctx *steps.TestContext) *UserServiceTestContext {
	if ctx.UserTestContext == nil {
		ctx.UserTestContext = &steps.UserTestContext{}
	}

	// Create a new context if it doesn't exist
	userCtx := &UserServiceTestContext{}

	// Try to get existing data from the generic context
	if ctx.UserTestContext.TestFirebaseUID != "" {
		userCtx.TestFirebaseUID = ctx.UserTestContext.TestFirebaseUID
	}
	if ctx.UserTestContext.TestDisplayName != "" {
		userCtx.TestDisplayName = ctx.UserTestContext.TestDisplayName
	}
	if ctx.UserTestContext.CreatedUserID != "" {
		userCtx.CreatedUserID = ctx.UserTestContext.CreatedUserID
	}
	if ctx.UserTestContext.LastError != nil {
		userCtx.LastError = ctx.UserTestContext.LastError
	}

	// Try to cast the client
	if client, ok := ctx.UserTestContext.UserClient.(userv1connect.UserServiceClient); ok {
		userCtx.Client = client
	}

	// Try to cast stored responses
	if ctx.UserTestContext.RegisterUserResponse != nil {
		if resp, ok := ctx.UserTestContext.RegisterUserResponse.(*connect.Response[userv1.RegisterUserResponse]); ok {
			userCtx.RegisterResponse = resp
		}
	}
	if ctx.UserTestContext.GetUserProfileResponse != nil {
		if resp, ok := ctx.UserTestContext.GetUserProfileResponse.(*connect.Response[userv1.UserProfile]); ok {
			userCtx.ProfileResponse = resp
		}
	}
	if ctx.UserTestContext.UpdateUserProfileResponse != nil {
		if resp, ok := ctx.UserTestContext.UpdateUserProfileResponse.(*connect.Response[userv1.UpdateUserProfileResponse]); ok {
			userCtx.UpdateResponse = resp
		}
	}

	// Try to cast stored concurrent results
	if ctx.UserTestContext.ConcurrentResults != nil {
		concurrentResults := make([]ConcurrentUserResult, len(ctx.UserTestContext.ConcurrentResults))
		for i, result := range ctx.UserTestContext.ConcurrentResults {
			if concurrentResult, ok := result.(ConcurrentUserResult); ok {
				concurrentResults[i] = concurrentResult
			}
		}
		userCtx.ConcurrentResults = concurrentResults
	}

	return userCtx
}

// Helper function to save user service test context
func setUserServiceContext(ctx *steps.TestContext, userCtx *UserServiceTestContext) {
	if ctx.UserTestContext == nil {
		ctx.UserTestContext = &steps.UserTestContext{}
	}

	ctx.UserTestContext.UserClient = userCtx.Client
	ctx.UserTestContext.TestFirebaseUID = userCtx.TestFirebaseUID
	ctx.UserTestContext.TestDisplayName = userCtx.TestDisplayName
	ctx.UserTestContext.CreatedUserID = userCtx.CreatedUserID
	ctx.UserTestContext.LastError = userCtx.LastError

	// Store responses as interfaces
	ctx.UserTestContext.RegisterUserResponse = userCtx.RegisterResponse
	ctx.UserTestContext.GetUserProfileResponse = userCtx.ProfileResponse
	ctx.UserTestContext.UpdateUserProfileResponse = userCtx.UpdateResponse

	// Store concurrent results
	if userCtx.ConcurrentResults != nil {
		interfaceResults := make([]interface{}, len(userCtx.ConcurrentResults))
		for i, result := range userCtx.ConcurrentResults {
			interfaceResults[i] = result
		}
		ctx.UserTestContext.ConcurrentResults = interfaceResults
	}
}

// Given Steps

// GivenUserServiceClientWithGRPC creates a gRPC user service client
func GivenUserServiceClientWithGRPC(ctx *steps.TestContext, serverURL string) {
	userCtx := getUserServiceContext(ctx)

	// Use HTTP/2 client for pure gRPC protocol
	h2Client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}

	userCtx.Client = userv1connect.NewUserServiceClient(
		h2Client,
		serverURL,
		connect.WithGRPC(),
	)

	setUserServiceContext(ctx, userCtx)
}

// GivenUserServiceClientWithGRPCWeb creates a gRPC-Web user service client
func GivenUserServiceClientWithGRPCWeb(ctx *steps.TestContext, serverURL string) {
	userCtx := getUserServiceContext(ctx)

	userCtx.Client = userv1connect.NewUserServiceClient(
		http.DefaultClient,
		serverURL,
		connect.WithGRPCWeb(),
	)

	setUserServiceContext(ctx, userCtx)
}

// GivenExistingUserIsRegistered registers a user for testing purposes
func GivenExistingUserIsRegistered(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	// Generate unique test data
	userCtx.TestFirebaseUID = "firebase_" + uuid.New().String()
	userCtx.TestDisplayName = "Test User " + uuid.New().String()[:8]

	// Register the user
	registerReq := &userv1.RegisterUserRequest{
		FirebaseUid:  userCtx.TestFirebaseUID,
		DisplayName:  userCtx.TestDisplayName,
		AuthProvider: "google",
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(registerReq)
	resp, err := userCtx.Client.RegisterUser(connectCtx, req)

	require.NoError(ctx.T, err, "Failed to register user for test setup")
	require.NotNil(ctx.T, resp, "Registration response should not be nil")
	require.NotNil(ctx.T, resp.Msg, "Registration response message should not be nil")
	require.NotEmpty(ctx.T, resp.Msg.UserId, "User ID should not be empty")

	userCtx.CreatedUserID = resp.Msg.UserId
	userCtx.RegisterResponse = resp

	setUserServiceContext(ctx, userCtx)
}

// When Steps

// WhenIRegisterUserWithValidData registers a user with valid data
func WhenIRegisterUserWithValidData(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	// Generate unique test data
	userCtx.TestFirebaseUID = "firebase_" + uuid.New().String()
	userCtx.TestDisplayName = "Test User " + uuid.New().String()[:8]

	userCtx.RegisterRequest = &userv1.RegisterUserRequest{
		FirebaseUid:  userCtx.TestFirebaseUID,
		DisplayName:  userCtx.TestDisplayName,
		AuthProvider: "google",
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(userCtx.RegisterRequest)
	resp, err := userCtx.Client.RegisterUser(connectCtx, req)
	userCtx.RegisterResponse = resp
	userCtx.LastError = err

	setUserServiceContext(ctx, userCtx)
}

// WhenIRegisterUserWithSameFirebaseUID tries to register user with existing Firebase UID
func WhenIRegisterUserWithSameFirebaseUID(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	// Use the same Firebase UID as the existing user
	userCtx.RegisterRequest = &userv1.RegisterUserRequest{
		FirebaseUid:  userCtx.TestFirebaseUID, // Same as existing user
		DisplayName:  "Another User",
		AuthProvider: "google",
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(userCtx.RegisterRequest)
	resp, err := userCtx.Client.RegisterUser(connectCtx, req)
	userCtx.RegisterResponse = resp
	userCtx.LastError = err

	setUserServiceContext(ctx, userCtx)
}

// WhenIRegisterUserWithEmptyFirebaseUID tries to register user with empty Firebase UID
func WhenIRegisterUserWithEmptyFirebaseUID(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	userCtx.RegisterRequest = &userv1.RegisterUserRequest{
		FirebaseUid:  "", // Empty Firebase UID
		DisplayName:  "Test User",
		AuthProvider: "google",
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(userCtx.RegisterRequest)
	resp, err := userCtx.Client.RegisterUser(connectCtx, req)
	userCtx.RegisterResponse = resp
	userCtx.LastError = err

	setUserServiceContext(ctx, userCtx)
}

// WhenIGetUserProfileWithValidID gets user profile with valid user ID
func WhenIGetUserProfileWithValidID(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	userCtx.ProfileRequest = &userv1.GetUserProfileRequest{
		UserId: userCtx.CreatedUserID,
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(userCtx.ProfileRequest)
	resp, err := userCtx.Client.GetUserProfile(connectCtx, req)
	userCtx.ProfileResponse = resp
	userCtx.LastError = err

	setUserServiceContext(ctx, userCtx)
}

// WhenIGetUserProfileWithNonExistentID gets user profile with non-existent user ID
func WhenIGetUserProfileWithNonExistentID(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	nonExistentID := uuid.New().String()
	userCtx.ProfileRequest = &userv1.GetUserProfileRequest{
		UserId: nonExistentID,
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(userCtx.ProfileRequest)
	resp, err := userCtx.Client.GetUserProfile(connectCtx, req)
	userCtx.ProfileResponse = resp
	userCtx.LastError = err

	setUserServiceContext(ctx, userCtx)
}

// WhenIUpdateUserProfileWithValidData updates user profile with valid data
func WhenIUpdateUserProfileWithValidData(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	newDisplayName := "Updated User " + uuid.New().String()[:8]
	newAvatarURL := "https://example.com/avatar/" + uuid.New().String() + ".jpg"

	userCtx.UpdateRequest = &userv1.UpdateUserProfileRequest{
		UserId:      userCtx.CreatedUserID,
		DisplayName: &newDisplayName,
		AvatarUrl:   &newAvatarURL,
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(userCtx.UpdateRequest)
	resp, err := userCtx.Client.UpdateUserProfile(connectCtx, req)
	userCtx.UpdateResponse = resp
	userCtx.LastError = err

	setUserServiceContext(ctx, userCtx)
}

// WhenIMakeConcurrentUserRegistrations makes multiple concurrent user registrations
func WhenIMakeConcurrentUserRegistrations(ctx *steps.TestContext, numRequests int) {
	userCtx := getUserServiceContext(ctx)

	results := make([]ConcurrentUserResult, numRequests)
	var wg sync.WaitGroup

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Generate unique test data for each request
			uniqueFirebaseUID := "firebase_concurrent_" + uuid.New().String()
			uniqueDisplayName := "Concurrent User " + uuid.New().String()[:8]

			registerReq := &userv1.RegisterUserRequest{
				FirebaseUid:  uniqueFirebaseUID,
				DisplayName:  uniqueDisplayName,
				AuthProvider: "google",
			}

			connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req := connect.NewRequest(registerReq)
			resp, err := userCtx.Client.RegisterUser(connectCtx, req)

			results[index] = ConcurrentUserResult{
				Response: resp,
				Error:    err,
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	userCtx.ConcurrentResults = results
	setUserServiceContext(ctx, userCtx)
}

// Then Steps

// ThenUserRegistrationShouldSucceed checks that user registration succeeded
func ThenUserRegistrationShouldSucceed(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	assert.NoError(ctx.T, userCtx.LastError, "User registration should not return an error")
	assert.NotNil(ctx.T, userCtx.RegisterResponse, "Registration response should not be nil")
	assert.NotNil(ctx.T, userCtx.RegisterResponse.Msg, "Registration response message should not be nil")
	assert.NotEmpty(ctx.T, userCtx.RegisterResponse.Msg.UserId, "User ID should not be empty")

	// Store the created user ID for potential use in subsequent tests
	userCtx.CreatedUserID = userCtx.RegisterResponse.Msg.UserId
	setUserServiceContext(ctx, userCtx)
}

// ThenResponseShouldContainValidUserID checks that response contains a valid UUID
func ThenResponseShouldContainValidUserID(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	require.NotNil(ctx.T, userCtx.RegisterResponse, "Registration response should not be nil")
	require.NotNil(ctx.T, userCtx.RegisterResponse.Msg, "Registration response message should not be nil")

	userID := userCtx.RegisterResponse.Msg.UserId
	_, err := uuid.Parse(userID)
	assert.NoError(ctx.T, err, "User ID should be a valid UUID, got: %s", userID)
}

// ThenUserRegistrationShouldFailWithAlreadyExistsError checks for AlreadyExists error
func ThenUserRegistrationShouldFailWithAlreadyExistsError(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	assert.Error(ctx.T, userCtx.LastError, "User registration should return an error")

	if connectErr, ok := userCtx.LastError.(*connect.Error); ok {
		assert.Equal(ctx.T, connect.CodeAlreadyExists, connectErr.Code(),
			"Expected AlreadyExists error, got: %v", connectErr.Code())
	} else {
		ctx.T.Errorf("Expected Connect error, got: %T", userCtx.LastError)
	}
}

// ThenUserRegistrationShouldFailWithInvalidArgumentError checks for InvalidArgument error
func ThenUserRegistrationShouldFailWithInvalidArgumentError(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	assert.Error(ctx.T, userCtx.LastError, "User registration should return an error")

	if connectErr, ok := userCtx.LastError.(*connect.Error); ok {
		assert.Equal(ctx.T, connect.CodeInvalidArgument, connectErr.Code(),
			"Expected InvalidArgument error, got: %v", connectErr.Code())
	} else {
		ctx.T.Errorf("Expected Connect error, got: %T", userCtx.LastError)
	}
}

// ThenUserProfileShouldBeRetrieved checks that user profile was retrieved successfully
func ThenUserProfileShouldBeRetrieved(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	assert.NoError(ctx.T, userCtx.LastError, "User profile retrieval should not return an error")
	assert.NotNil(ctx.T, userCtx.ProfileResponse, "Profile response should not be nil")
	assert.NotNil(ctx.T, userCtx.ProfileResponse.Msg, "Profile response message should not be nil")
	assert.NotEmpty(ctx.T, userCtx.ProfileResponse.Msg.UserId, "User ID should not be empty")
	assert.NotNil(ctx.T, userCtx.ProfileResponse.Msg.CreatedAt, "Created timestamp should not be nil")
}

// ThenProfileShouldContainCorrectData checks that profile contains expected data
func ThenProfileShouldContainCorrectData(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	require.NotNil(ctx.T, userCtx.ProfileResponse, "Profile response should not be nil")
	require.NotNil(ctx.T, userCtx.ProfileResponse.Msg, "Profile response message should not be nil")

	profile := userCtx.ProfileResponse.Msg

	// Check that the profile contains the expected user ID
	assert.Equal(ctx.T, userCtx.CreatedUserID, profile.UserId,
		"Profile user ID should match the created user ID")

	// Check that the profile contains the expected display name
	if userCtx.TestDisplayName != "" {
		assert.Equal(ctx.T, userCtx.TestDisplayName, profile.DisplayName,
			"Profile display name should match the test display name")
	}

	// Check that candy balance is initialized to 0
	assert.Equal(ctx.T, int32(0), profile.CandyBalance,
		"Candy balance should be initialized to 0 for new user")
}

// ThenUserProfileRetrievalShouldFailWithNotFoundError checks for NotFound error
func ThenUserProfileRetrievalShouldFailWithNotFoundError(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	assert.Error(ctx.T, userCtx.LastError, "User profile retrieval should return an error")

	if connectErr, ok := userCtx.LastError.(*connect.Error); ok {
		assert.Equal(ctx.T, connect.CodeNotFound, connectErr.Code(),
			"Expected NotFound error, got: %v", connectErr.Code())
	} else {
		ctx.T.Errorf("Expected Connect error, got: %T", userCtx.LastError)
	}
}

// ThenUserProfileUpdateShouldSucceed checks that user profile update succeeded
func ThenUserProfileUpdateShouldSucceed(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	assert.NoError(ctx.T, userCtx.LastError, "User profile update should not return an error")
	assert.NotNil(ctx.T, userCtx.UpdateResponse, "Update response should not be nil")
	assert.NotNil(ctx.T, userCtx.UpdateResponse.Msg, "Update response message should not be nil")
}

// ThenAllConcurrentUserRegistrationsShouldSucceed checks that all concurrent registrations succeeded
func ThenAllConcurrentUserRegistrationsShouldSucceed(ctx *steps.TestContext) {
	userCtx := getUserServiceContext(ctx)

	assert.NotEmpty(ctx.T, userCtx.ConcurrentResults, "Concurrent results should not be empty")

	successCount := 0
	for i, result := range userCtx.ConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Concurrent registration %d failed with error: %v", i+1, result.Error)
			continue
		}

		if result.Response == nil {
			ctx.T.Errorf("Concurrent registration %d has nil response", i+1)
			continue
		}

		if result.Response.Msg == nil {
			ctx.T.Errorf("Concurrent registration %d has nil response message", i+1)
			continue
		}

		if result.Response.Msg.UserId == "" {
			ctx.T.Errorf("Concurrent registration %d has empty user ID", i+1)
			continue
		}

		// Validate that the returned user ID is a valid UUID
		if _, err := uuid.Parse(result.Response.Msg.UserId); err != nil {
			ctx.T.Errorf("Concurrent registration %d returned invalid UUID: %s", i+1, result.Response.Msg.UserId)
			continue
		}

		successCount++
	}

	expectedCount := len(userCtx.ConcurrentResults)
	assert.Equal(ctx.T, expectedCount, successCount,
		"Expected all %d concurrent registrations to succeed, but only %d succeeded",
		expectedCount, successCount)
}
