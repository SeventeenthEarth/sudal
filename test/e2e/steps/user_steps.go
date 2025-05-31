package steps

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/http2"

	userv1 "github.com/seventeenthearth/sudal/gen/go/user/v1"
	"github.com/seventeenthearth/sudal/gen/go/user/v1/userv1connect"
)

// UserResult represents the result of a user service call
type UserResult struct {
	RegisterResponse *connect.Response[userv1.RegisterUserResponse]
	ProfileResponse  *connect.Response[userv1.UserProfile]
	UpdateResponse   *connect.Response[userv1.UpdateUserProfileResponse]
	Error            error
	Protocol         string
	OperationType    string // "register", "get_profile", "update_profile"
}

// UserStepContext holds user-specific test context with concrete types
type UserStepContext struct {
	UserClient                userv1connect.UserServiceClient
	RegisterUserRequest       *userv1.RegisterUserRequest
	RegisterUserResponse      *connect.Response[userv1.RegisterUserResponse]
	GetUserProfileRequest     *userv1.GetUserProfileRequest
	GetUserProfileResponse    *connect.Response[userv1.UserProfile]
	UpdateUserProfileRequest  *userv1.UpdateUserProfileRequest
	UpdateUserProfileResponse *connect.Response[userv1.UpdateUserProfileResponse]
	LastError                 error
	CreatedUserID             string
	TestFirebaseUID           string
	TestDisplayName           string
	TestAvatarURL             string
	Protocol                  string
	Timeout                   time.Duration
	ConcurrentResults         []UserResult
}

// getUserStepContext gets or creates a UserStepContext from TestContext
func getUserStepContext(ctx *TestContext) *UserStepContext {
	if ctx.UserTestContext == nil {
		ctx.UserTestContext = &UserTestContext{}
	}

	// If we don't have a concrete context stored, create one
	if ctx.UserTestContext.UserClient == nil {
		return &UserStepContext{}
	}

	// Try to cast the stored client back to concrete type
	client, ok := ctx.UserTestContext.UserClient.(userv1connect.UserServiceClient)
	if !ok {
		return &UserStepContext{}
	}

	userCtx := &UserStepContext{
		UserClient:      client,
		Protocol:        ctx.UserTestContext.Protocol,
		Timeout:         ctx.UserTestContext.Timeout,
		TestFirebaseUID: ctx.UserTestContext.TestFirebaseUID,
		TestDisplayName: ctx.UserTestContext.TestDisplayName,
		TestAvatarURL:   ctx.UserTestContext.TestAvatarURL,
		CreatedUserID:   ctx.UserTestContext.CreatedUserID,
		LastError:       ctx.UserTestContext.LastError,
	}

	// Cast stored requests and responses
	if ctx.UserTestContext.RegisterUserRequest != nil {
		if req, ok := ctx.UserTestContext.RegisterUserRequest.(*userv1.RegisterUserRequest); ok {
			userCtx.RegisterUserRequest = req
		}
	}
	if ctx.UserTestContext.RegisterUserResponse != nil {
		if resp, ok := ctx.UserTestContext.RegisterUserResponse.(*connect.Response[userv1.RegisterUserResponse]); ok {
			userCtx.RegisterUserResponse = resp
		}
	}
	if ctx.UserTestContext.GetUserProfileRequest != nil {
		if req, ok := ctx.UserTestContext.GetUserProfileRequest.(*userv1.GetUserProfileRequest); ok {
			userCtx.GetUserProfileRequest = req
		}
	}
	if ctx.UserTestContext.GetUserProfileResponse != nil {
		if resp, ok := ctx.UserTestContext.GetUserProfileResponse.(*connect.Response[userv1.UserProfile]); ok {
			userCtx.GetUserProfileResponse = resp
		}
	}
	if ctx.UserTestContext.UpdateUserProfileRequest != nil {
		if req, ok := ctx.UserTestContext.UpdateUserProfileRequest.(*userv1.UpdateUserProfileRequest); ok {
			userCtx.UpdateUserProfileRequest = req
		}
	}
	if ctx.UserTestContext.UpdateUserProfileResponse != nil {
		if resp, ok := ctx.UserTestContext.UpdateUserProfileResponse.(*connect.Response[userv1.UpdateUserProfileResponse]); ok {
			userCtx.UpdateUserProfileResponse = resp
		}
	}
	if ctx.UserTestContext.ConcurrentResults != nil {
		// Convert []protocol{} to []UserResult
		userResults := make([]UserResult, len(ctx.UserTestContext.ConcurrentResults))
		for i, result := range ctx.UserTestContext.ConcurrentResults {
			if userResult, ok := result.(UserResult); ok {
				userResults[i] = userResult
			}
		}
		userCtx.ConcurrentResults = userResults
	}

	return userCtx
}

// setUserStepContext stores a UserStepContext back to TestContext
func setUserStepContext(ctx *TestContext, userCtx *UserStepContext) {
	if ctx.UserTestContext == nil {
		ctx.UserTestContext = &UserTestContext{}
	}

	ctx.UserTestContext.UserClient = userCtx.UserClient
	ctx.UserTestContext.Protocol = userCtx.Protocol
	ctx.UserTestContext.Timeout = userCtx.Timeout
	ctx.UserTestContext.TestFirebaseUID = userCtx.TestFirebaseUID
	ctx.UserTestContext.TestDisplayName = userCtx.TestDisplayName
	ctx.UserTestContext.TestAvatarURL = userCtx.TestAvatarURL
	ctx.UserTestContext.CreatedUserID = userCtx.CreatedUserID
	ctx.UserTestContext.LastError = userCtx.LastError

	// Store concrete types as interfaces
	ctx.UserTestContext.RegisterUserRequest = userCtx.RegisterUserRequest
	ctx.UserTestContext.RegisterUserResponse = userCtx.RegisterUserResponse
	ctx.UserTestContext.GetUserProfileRequest = userCtx.GetUserProfileRequest
	ctx.UserTestContext.GetUserProfileResponse = userCtx.GetUserProfileResponse
	ctx.UserTestContext.UpdateUserProfileRequest = userCtx.UpdateUserProfileRequest
	ctx.UserTestContext.UpdateUserProfileResponse = userCtx.UpdateUserProfileResponse

	// Convert []UserResult to []protocol{}
	if userCtx.ConcurrentResults != nil {
		interfaceResults := make([]interface{}, len(userCtx.ConcurrentResults))
		for i, result := range userCtx.ConcurrentResults {
			interfaceResults[i] = result
		}
		ctx.UserTestContext.ConcurrentResults = interfaceResults
	}
}

// User-specific Given Steps

// GivenUserServiceClientWithProtocol establishes a user service client with specific protocol
func GivenUserServiceClientWithProtocol(ctx *TestContext, serverURL, protocol string) {
	userCtx := getUserStepContext(ctx)

	var client userv1connect.UserServiceClient

	switch protocol {
	case "grpc":
		// Use HTTP/2 client for pure gRPC protocol
		h2Client := &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		}
		client = userv1connect.NewUserServiceClient(
			h2Client,
			serverURL,
			connect.WithGRPC(),
		)
	case "grpc-web":
		client = userv1connect.NewUserServiceClient(
			http.DefaultClient,
			serverURL,
			connect.WithGRPCWeb(),
		)
	case "http":
		fallthrough
	default:
		client = userv1connect.NewUserServiceClient(
			http.DefaultClient,
			serverURL,
		)
	}

	userCtx.UserClient = client
	userCtx.Protocol = protocol
	setUserStepContext(ctx, userCtx)
}

// GivenUserServiceClientWithProtocolAndTimeout establishes a user service client with protocol and timeout
func GivenUserServiceClientWithProtocolAndTimeout(ctx *TestContext, serverURL, protocol string, timeout time.Duration) {
	GivenUserServiceClientWithProtocol(ctx, serverURL, protocol)
	if ctx.UserTestContext != nil {
		ctx.UserTestContext.Timeout = timeout
	}
}

// GivenValidUserRegistrationData sets up valid user registration data
func GivenValidUserRegistrationData(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	// Generate unique test data
	userCtx.TestFirebaseUID = "firebase_" + uuid.New().String()
	userCtx.TestDisplayName = "Test User " + uuid.New().String()[:8]

	userCtx.RegisterUserRequest = &userv1.RegisterUserRequest{
		FirebaseUid:  userCtx.TestFirebaseUID,
		DisplayName:  userCtx.TestDisplayName,
		AuthProvider: "google",
	}

	setUserStepContext(ctx, userCtx)
}

// GivenInvalidUserRegistrationData sets up invalid user registration data
func GivenInvalidUserRegistrationData(ctx *TestContext, invalidType string) {
	if ctx.UserTestContext == nil {
		ctx.UserTestContext = &UserTestContext{}
	}

	switch invalidType {
	case "empty_firebase_uid":
		ctx.UserTestContext.RegisterUserRequest = &userv1.RegisterUserRequest{
			FirebaseUid:  "",
			DisplayName:  "Test User",
			AuthProvider: "google",
		}
	case "empty_auth_provider":
		ctx.UserTestContext.RegisterUserRequest = &userv1.RegisterUserRequest{
			FirebaseUid:  "firebase_" + uuid.New().String(),
			DisplayName:  "Test User",
			AuthProvider: "",
		}
	case "invalid_display_name":
		ctx.UserTestContext.RegisterUserRequest = &userv1.RegisterUserRequest{
			FirebaseUid:  "firebase_" + uuid.New().String(),
			DisplayName:  "", // Empty display name
			AuthProvider: "google",
		}
	case "long_display_name":
		longName := ""
		for i := 0; i < 101; i++ { // Exceed 100 character limit
			longName += "a"
		}
		ctx.UserTestContext.RegisterUserRequest = &userv1.RegisterUserRequest{
			FirebaseUid:  "firebase_" + uuid.New().String(),
			DisplayName:  longName,
			AuthProvider: "google",
		}
	default:
		ctx.UserTestContext.RegisterUserRequest = &userv1.RegisterUserRequest{
			FirebaseUid:  "",
			DisplayName:  "",
			AuthProvider: "",
		}
	}
}

// GivenExistingUser sets up an existing user for testing
func GivenExistingUser(ctx *TestContext) {
	// First create a valid user
	GivenValidUserRegistrationData(ctx)
	WhenIRegisterUser(ctx)

	userCtx := getUserStepContext(ctx)

	// Verify registration was successful
	if userCtx.LastError != nil {
		ctx.T.Fatalf("Failed to create existing user for test: %v", userCtx.LastError)
	}
	if userCtx.RegisterUserResponse == nil || userCtx.RegisterUserResponse.Msg == nil {
		ctx.T.Fatalf("Failed to get user ID from registration response")
	}

	userCtx.CreatedUserID = userCtx.RegisterUserResponse.Msg.UserId
	setUserStepContext(ctx, userCtx)
}

// GivenValidUserProfileUpdateData sets up valid user profile update data
func GivenValidUserProfileUpdateData(ctx *TestContext) {
	if ctx.UserTestContext == nil {
		ctx.UserTestContext = &UserTestContext{}
	}

	newDisplayName := "Updated User " + uuid.New().String()[:8]
	newAvatarURL := "https://example.com/avatar/" + uuid.New().String() + ".jpg"

	ctx.UserTestContext.TestDisplayName = newDisplayName
	ctx.UserTestContext.TestAvatarURL = newAvatarURL

	ctx.UserTestContext.UpdateUserProfileRequest = &userv1.UpdateUserProfileRequest{
		UserId:      ctx.UserTestContext.CreatedUserID,
		DisplayName: &newDisplayName,
		AvatarUrl:   &newAvatarURL,
	}
}

// User-specific When Steps

// WhenIRegisterUser makes a user registration request
func WhenIRegisterUser(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	if userCtx.UserClient == nil {
		userCtx.LastError = assert.AnError
		setUserStepContext(ctx, userCtx)
		return
	}

	timeout := userCtx.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := connect.NewRequest(userCtx.RegisterUserRequest)
	resp, err := userCtx.UserClient.RegisterUser(connectCtx, req)
	userCtx.RegisterUserResponse = resp
	userCtx.LastError = err

	setUserStepContext(ctx, userCtx)
}

// WhenIGetUserProfile makes a get user profile request
func WhenIGetUserProfile(ctx *TestContext, userID string) {
	userCtx := getUserStepContext(ctx)

	if userCtx.UserClient == nil {
		userCtx.LastError = assert.AnError
		setUserStepContext(ctx, userCtx)
		return
	}

	timeout := userCtx.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	userCtx.GetUserProfileRequest = &userv1.GetUserProfileRequest{
		UserId: userID,
	}

	req := connect.NewRequest(userCtx.GetUserProfileRequest)
	resp, err := userCtx.UserClient.GetUserProfile(connectCtx, req)
	userCtx.GetUserProfileResponse = resp
	userCtx.LastError = err

	setUserStepContext(ctx, userCtx)
}

// WhenIUpdateUserProfile makes an update user profile request
func WhenIUpdateUserProfile(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	if userCtx.UserClient == nil {
		userCtx.LastError = assert.AnError
		setUserStepContext(ctx, userCtx)
		return
	}

	timeout := userCtx.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := connect.NewRequest(userCtx.UpdateUserProfileRequest)
	resp, err := userCtx.UserClient.UpdateUserProfile(connectCtx, req)
	userCtx.UpdateUserProfileResponse = resp
	userCtx.LastError = err

	setUserStepContext(ctx, userCtx)
}

// WhenIMakeConcurrentUserRegistrations makes multiple concurrent user registration requests
func WhenIMakeConcurrentUserRegistrations(ctx *TestContext, numRequests int) {
	userCtx := getUserStepContext(ctx)

	if userCtx.UserClient == nil {
		userCtx.LastError = assert.AnError
		setUserStepContext(ctx, userCtx)
		return
	}

	var wg sync.WaitGroup
	results := make([]UserResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			timeout := userCtx.Timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}

			connectCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Create unique registration data for each request
			uniqueFirebaseUID := "firebase_concurrent_" + uuid.New().String()
			uniqueDisplayName := "Concurrent User " + uuid.New().String()[:8]

			registerReq := &userv1.RegisterUserRequest{
				FirebaseUid:  uniqueFirebaseUID,
				DisplayName:  uniqueDisplayName,
				AuthProvider: "google",
			}

			req := connect.NewRequest(registerReq)
			resp, err := userCtx.UserClient.RegisterUser(connectCtx, req)

			results[index] = UserResult{
				RegisterResponse: resp,
				Error:            err,
				Protocol:         userCtx.Protocol,
				OperationType:    "register",
			}
		}(i)
	}

	wg.Wait()
	userCtx.ConcurrentResults = results
	setUserStepContext(ctx, userCtx)
}

// User-specific Then Steps

// ThenUserRegistrationShouldSucceed checks that user registration succeeded in BDD style
func ThenUserRegistrationShouldSucceed(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	if userCtx.LastError != nil {
		ctx.T.Errorf("Expected user registration to succeed, but got error: %v", userCtx.LastError)
		return
	}

	if userCtx.RegisterUserResponse == nil {
		ctx.T.Errorf("Expected user registration response to exist, but it was nil")
		return
	}

	if userCtx.RegisterUserResponse.Msg == nil {
		ctx.T.Errorf("Expected user registration response message to exist, but it was nil")
		return
	}

	if userCtx.RegisterUserResponse.Msg.UserId == "" {
		ctx.T.Errorf("Expected user registration response to contain user ID, but it was empty")
		return
	}

	// Validate that the returned user ID is a valid UUID
	if _, err := uuid.Parse(userCtx.RegisterUserResponse.Msg.UserId); err != nil {
		ctx.T.Errorf("Expected user ID to be a valid UUID, but got: %s", userCtx.RegisterUserResponse.Msg.UserId)
	}

	ctx.T.Logf("User registration succeeded with ID: %s", userCtx.RegisterUserResponse.Msg.UserId)
}

// ThenUserRegistrationShouldFail checks that user registration failed as expected in BDD style
func ThenUserRegistrationShouldFail(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	if userCtx.LastError == nil {
		ctx.T.Errorf("Expected user registration to fail, but it succeeded")
		return
	}

	ctx.T.Logf("User registration failed as expected: %v", userCtx.LastError)
}

// ThenUserRegistrationShouldFailWithCode checks that user registration failed with specific error code
func ThenUserRegistrationShouldFailWithCode(ctx *TestContext, expectedCode connect.Code) {
	userCtx := getUserStepContext(ctx)

	if userCtx.LastError == nil {
		ctx.T.Errorf("Expected user registration to fail, but it succeeded")
		return
	}

	connectErr, ok := userCtx.LastError.(*connect.Error)
	if !ok {
		ctx.T.Errorf("Expected error to be a Connect error, but got: %T", userCtx.LastError)
		return
	}

	if connectErr.Code() != expectedCode {
		ctx.T.Errorf("Expected error code to be %v, but got %v", expectedCode, connectErr.Code())
		return
	}

	ctx.T.Logf("User registration failed with expected code %v: %v", expectedCode, userCtx.LastError)
}

// ThenUserProfileShouldBeRetrieved checks that user profile was retrieved successfully in BDD style
func ThenUserProfileShouldBeRetrieved(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	if userCtx.LastError != nil {
		ctx.T.Errorf("Expected user profile retrieval to succeed, but got error: %v", userCtx.LastError)
		return
	}

	if userCtx.GetUserProfileResponse == nil {
		ctx.T.Errorf("Expected user profile response to exist, but it was nil")
		return
	}

	if userCtx.GetUserProfileResponse.Msg == nil {
		ctx.T.Errorf("Expected user profile response message to exist, but it was nil")
		return
	}

	profile := userCtx.GetUserProfileResponse.Msg
	if profile.UserId == "" {
		ctx.T.Errorf("Expected user profile to contain user ID, but it was empty")
		return
	}

	if profile.CreatedAt == nil {
		ctx.T.Errorf("Expected user profile to contain created_at timestamp, but it was nil")
		return
	}

	ctx.T.Logf("User profile retrieved successfully for ID: %s", profile.UserId)
}

// ThenUserProfileShouldContainCorrectData checks that user profile contains expected data
func ThenUserProfileShouldContainCorrectData(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	if userCtx.GetUserProfileResponse == nil || userCtx.GetUserProfileResponse.Msg == nil {
		ctx.T.Errorf("Expected user profile response to exist")
		return
	}

	profile := userCtx.GetUserProfileResponse.Msg

	// Check that the profile contains the expected display name
	if userCtx.TestDisplayName != "" {
		if profile.DisplayName != userCtx.TestDisplayName {
			ctx.T.Errorf("Expected display name to be '%s', but got '%s'", userCtx.TestDisplayName, profile.DisplayName)
		}
	}

	// Check that candy balance is initialized to 0
	if profile.CandyBalance != 0 {
		ctx.T.Errorf("Expected candy balance to be 0 for new user, but got %d", profile.CandyBalance)
	}

	ctx.T.Logf("User profile contains correct data: display_name=%s, candy_balance=%d", profile.DisplayName, profile.CandyBalance)
}

// ThenUserProfileRetrievalShouldFail checks that user profile retrieval failed as expected
func ThenUserProfileRetrievalShouldFail(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	if userCtx.LastError == nil {
		ctx.T.Errorf("Expected user profile retrieval to fail, but it succeeded")
		return
	}

	ctx.T.Logf("User profile retrieval failed as expected: %v", userCtx.LastError)
}

// ThenUserProfileRetrievalShouldFailWithCode checks that user profile retrieval failed with specific error code
func ThenUserProfileRetrievalShouldFailWithCode(ctx *TestContext, expectedCode connect.Code) {
	userCtx := getUserStepContext(ctx)

	if userCtx.LastError == nil {
		ctx.T.Errorf("Expected user profile retrieval to fail, but it succeeded")
		return
	}

	connectErr, ok := userCtx.LastError.(*connect.Error)
	if !ok {
		ctx.T.Errorf("Expected error to be a Connect error, but got: %T", userCtx.LastError)
		return
	}

	if connectErr.Code() != expectedCode {
		ctx.T.Errorf("Expected error code to be %v, but got %v", expectedCode, connectErr.Code())
		return
	}

	ctx.T.Logf("User profile retrieval failed with expected code %v: %v", expectedCode, userCtx.LastError)
}

// ThenUserProfileUpdateShouldSucceed checks that user profile update succeeded in BDD style
func ThenUserProfileUpdateShouldSucceed(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	if userCtx.LastError != nil {
		ctx.T.Errorf("Expected user profile update to succeed, but got error: %v", userCtx.LastError)
		return
	}

	if userCtx.UpdateUserProfileResponse == nil {
		ctx.T.Errorf("Expected user profile update response to exist, but it was nil")
		return
	}

	if userCtx.UpdateUserProfileResponse.Msg == nil {
		ctx.T.Errorf("Expected user profile update response message to exist, but it was nil")
		return
	}

	ctx.T.Logf("User profile update succeeded")
}

// ThenAllConcurrentUserRegistrationsShouldSucceed checks that all concurrent user registrations succeeded
func ThenAllConcurrentUserRegistrationsShouldSucceed(ctx *TestContext) {
	userCtx := getUserStepContext(ctx)

	if len(userCtx.ConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent user registration results to exist, but none were found")
		return
	}

	failedCount := 0
	for i, result := range userCtx.ConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected concurrent user registration %d to succeed, but got error: %v", i+1, result.Error)
			failedCount++
			continue
		}
		if result.RegisterResponse == nil {
			ctx.T.Errorf("Expected concurrent user registration %d to have a response, but none was received", i+1)
			failedCount++
			continue
		}
		if result.RegisterResponse.Msg == nil {
			ctx.T.Errorf("Expected concurrent user registration %d response message to exist, but it was nil", i+1)
			failedCount++
			continue
		}
		if result.RegisterResponse.Msg.UserId == "" {
			ctx.T.Errorf("Expected concurrent user registration %d to return user ID, but it was empty", i+1)
			failedCount++
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent user registrations to succeed, but %d failed", len(userCtx.ConcurrentResults), failedCount)
	} else {
		ctx.T.Logf("All %d concurrent user registrations succeeded", len(userCtx.ConcurrentResults))
	}
}
