package steps

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"golang.org/x/net/http2"

	userv1 "github.com/seventeenthearth/sudal/gen/go/user/v1"
	"github.com/seventeenthearth/sudal/gen/go/user/v1/userv1connect"
	"github.com/seventeenthearth/sudal/test/e2e/helpers"
)

// UserCtx holds the context for user-related test scenarios
// Now uses real Firebase authentication instead of fake tokens
type UserCtx struct {
	baseURL      string
	grpcEndpoint string

	// Firebase Auth client for real authentication
	firebaseClient *helpers.FirebaseAuthClient

	// HTTP-related fields (for negative REST tests)
	httpClient   *http.Client
	lastResponse *http.Response
	lastError    error

	// gRPC-related fields
	grpcClient       userv1connect.UserServiceClient
	registerRequest  *userv1.RegisterUserRequest
	registerResponse *connect.Response[userv1.RegisterUserResponse]
	profileRequest   *userv1.GetUserProfileRequest
	profileResponse  *connect.Response[userv1.UserProfile]
	updateRequest    *userv1.UpdateUserProfileRequest
	updateResponse   *connect.Response[userv1.UpdateUserProfileResponse]
	grpcError        error

	// Firebase authentication data
	firebaseIDToken string
	firebaseUID     string
	email           string
	password        string

	// Test data
	testFirebaseUID string
	testDisplayName string
	createdUserID   string

	// Concurrent test results
	concurrentResults     []ConcurrentUserResult
	concurrentHTTPResults []ConcurrentHTTPResult

	// Shared response for cross-context communication
	sharedResponse *SharedHTTPResponse
}

// ConcurrentUserResult holds the result of a concurrent user operation
type ConcurrentUserResult struct {
	RegisterResponse *connect.Response[userv1.RegisterUserResponse]
	ProfileResponse  *connect.Response[userv1.UserProfile]
	UpdateResponse   *connect.Response[userv1.UpdateUserProfileResponse]
	Error            error
	OperationType    string // "register", "get_profile", "update_profile"
}

// ConcurrentHTTPResult holds the result of a concurrent HTTP request
type ConcurrentHTTPResult struct {
	Response *http.Response
	Body     []byte
	Error    error
}

// NewUserCtx creates a new UserCtx instance
func NewUserCtx() *UserCtx {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	grpcEndpoint := os.Getenv("GRPC_ADDR")
	if grpcEndpoint == "" {
		grpcEndpoint = "localhost:8080"
	}

	// Initialize Firebase client
	firebaseClient, err := helpers.NewFirebaseAuthClient()
	if err != nil {
		// Log error but don't fail initialization
		// This allows tests that don't need Firebase to still work
		fmt.Printf("Warning: Failed to initialize Firebase client: %v\n", err)
	}

	return &UserCtx{
		baseURL:        baseURL,
		grpcEndpoint:   grpcEndpoint,
		firebaseClient: firebaseClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Cleanup cleans up resources used by UserCtx
func (u *UserCtx) Cleanup() {
	// Close HTTP response if exists
	if u.lastResponse != nil {
		_ = u.lastResponse.Body.Close()
	}

	// Clean up Firebase user if exists
	if u.firebaseClient != nil && u.firebaseIDToken != "" {
		if err := u.firebaseClient.DeleteUser(u.firebaseIDToken); err != nil {
			// Log warning but don't fail the test
			fmt.Printf("Warning: Failed to delete Firebase user: %v\n", err)
		}
	}
}

// createFirebaseUser creates a new Firebase user and returns authentication data
func (u *UserCtx) createFirebaseUser() error {
	if u.firebaseClient == nil {
		return fmt.Errorf("firebase client not initialized")
	}

	// Generate random credentials
	u.email = helpers.GenerateRandomEmail()
	u.password = helpers.GenerateSecurePassword()

	// Add delay to avoid rate limiting
	time.Sleep(1 * time.Second)

	// Sign up with Firebase
	authResp, err := u.firebaseClient.SignUpWithEmailPassword(u.email, u.password)
	if err != nil {
		return fmt.Errorf("failed to sign up with Firebase: %w", err)
	}

	u.firebaseIDToken = authResp.IDToken
	u.firebaseUID = authResp.LocalID

	return nil
}

// ensureFirebaseAuth ensures Firebase authentication is set up
func (u *UserCtx) ensureFirebaseAuth() error {
	if u.firebaseIDToken == "" || u.firebaseUID == "" {
		return u.createFirebaseUser()
	}
	return nil
}

// Given Steps

func (u *UserCtx) theServerIsRunning() error {
	// Check if server is accessible via ping endpoint
	resp, err := u.httpClient.Get(fmt.Sprintf("%s/api/ping", u.baseURL))
	if err != nil {
		return fmt.Errorf("server is not running: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server is not healthy, status: %d", resp.StatusCode)
	}

	return nil
}

func (u *UserCtx) theGRPCUserClientIsConnected() error {
	// Use HTTP/2 client for pure gRPC protocol
	h2Client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
		Timeout: 10 * time.Second,
	}

	u.grpcClient = userv1connect.NewUserServiceClient(
		h2Client,
		u.baseURL,
		connect.WithGRPC(),
	)

	return nil
}

func (u *UserCtx) theGRPCWebUserClientIsConnected() error {
	// Use standard HTTP client for gRPC-Web
	u.grpcClient = userv1connect.NewUserServiceClient(
		u.httpClient,
		u.baseURL,
		connect.WithGRPCWeb(),
	)

	return nil
}

func (u *UserCtx) iHaveValidUserRegistrationData() error {
	// Ensure Firebase authentication is set up
	if err := u.ensureFirebaseAuth(); err != nil {
		return fmt.Errorf("failed to set up Firebase authentication: %w", err)
	}

	// Use real Firebase UID and generate display name
	u.testFirebaseUID = u.firebaseUID
	u.testDisplayName = "Test User " + uuid.New().String()[:8]

	u.registerRequest = &userv1.RegisterUserRequest{
		FirebaseUid:  u.testFirebaseUID,
		DisplayName:  u.testDisplayName,
		AuthProvider: "email",
	}

	return nil
}

func (u *UserCtx) iHaveInvalidUserRegistrationDataWithEmptyFirebaseUID() error {
	u.testDisplayName = "Test User"

	u.registerRequest = &userv1.RegisterUserRequest{
		FirebaseUid:  "", // Empty Firebase UID
		DisplayName:  u.testDisplayName,
		AuthProvider: "google",
	}

	return nil
}

func (u *UserCtx) anExistingUserIsRegistered() error {
	// First ensure we have a client
	if u.grpcClient == nil {
		if err := u.theGRPCUserClientIsConnected(); err != nil {
			return err
		}
	}

	// Ensure Firebase authentication is set up
	if err := u.ensureFirebaseAuth(); err != nil {
		return fmt.Errorf("failed to set up Firebase authentication: %w", err)
	}

	// Use Firebase UID and generate display name
	u.testFirebaseUID = u.firebaseUID
	u.testDisplayName = "Existing User " + uuid.New().String()[:8]

	// Register the user
	registerReq := &userv1.RegisterUserRequest{
		FirebaseUid:  u.testFirebaseUID,
		DisplayName:  u.testDisplayName,
		AuthProvider: "email",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(registerReq)
	// Use real Firebase token
	req.Header().Set("Authorization", "Bearer "+u.firebaseIDToken)

	resp, err := u.grpcClient.RegisterUser(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to register existing user: %v", err)
	}

	if resp.Msg == nil || resp.Msg.UserId == "" {
		return fmt.Errorf("failed to get user ID from registration response")
	}

	u.createdUserID = resp.Msg.UserId
	return nil
}

// When Steps

func (u *UserCtx) iRegisterAUserWithValidData() error {
	if u.grpcClient == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	if u.registerRequest == nil {
		return fmt.Errorf("registration request not prepared")
	}

	// Ensure Firebase authentication is set up
	if err := u.ensureFirebaseAuth(); err != nil {
		return fmt.Errorf("failed to set up Firebase authentication: %w", err)
	}

	// Update the request with real Firebase UID
	u.registerRequest.FirebaseUid = u.firebaseUID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(u.registerRequest)
	// Use real Firebase token
	req.Header().Set("Authorization", "Bearer "+u.firebaseIDToken)

	resp, err := u.grpcClient.RegisterUser(ctx, req)
	u.registerResponse = resp
	u.grpcError = err

	return nil
}

func (u *UserCtx) iRegisterAUserWithTheSameFirebaseUID() error {
	if u.grpcClient == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	// Use the same Firebase UID as the existing user
	u.registerRequest = &userv1.RegisterUserRequest{
		FirebaseUid:  u.testFirebaseUID, // Same as existing user
		DisplayName:  "Another User",
		AuthProvider: "email",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(u.registerRequest)
	// Use real Firebase token (same user trying to register again)
	req.Header().Set("Authorization", "Bearer "+u.firebaseIDToken)

	resp, err := u.grpcClient.RegisterUser(ctx, req)
	u.registerResponse = resp
	u.grpcError = err

	return nil
}

func (u *UserCtx) iRegisterAUserWithEmptyFirebaseUID() error {
	if u.grpcClient == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	if u.registerRequest == nil {
		return fmt.Errorf("registration request not prepared")
	}

	// Ensure Firebase authentication is set up
	if err := u.ensureFirebaseAuth(); err != nil {
		return fmt.Errorf("failed to set up Firebase authentication: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(u.registerRequest)
	// Use real Firebase token but with empty UID in request
	req.Header().Set("Authorization", "Bearer "+u.firebaseIDToken)

	resp, err := u.grpcClient.RegisterUser(ctx, req)
	u.registerResponse = resp
	u.grpcError = err

	return nil
}

func (u *UserCtx) iGetTheUserProfile() error {
	if u.grpcClient == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	if u.createdUserID == "" {
		return fmt.Errorf("no user ID available for profile request")
	}

	u.profileRequest = &userv1.GetUserProfileRequest{
		UserId: u.createdUserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(u.profileRequest)
	// Use real Firebase token
	req.Header().Set("Authorization", "Bearer "+u.firebaseIDToken)

	resp, err := u.grpcClient.GetUserProfile(ctx, req)
	u.profileResponse = resp
	u.grpcError = err

	return nil
}

func (u *UserCtx) iGetTheUserProfileWithNonExistentID() error {
	if u.grpcClient == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	// Ensure Firebase authentication is set up
	if err := u.ensureFirebaseAuth(); err != nil {
		return fmt.Errorf("failed to set up Firebase authentication: %w", err)
	}

	// Use a valid UUID that doesn't exist in the database
	// Generate a random UUID that is very unlikely to exist
	nonExistentID := "12345678-1234-5678-9abc-123456789abc"

	u.profileRequest = &userv1.GetUserProfileRequest{
		UserId: nonExistentID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(u.profileRequest)
	// Use real Firebase token
	req.Header().Set("Authorization", "Bearer "+u.firebaseIDToken)

	resp, err := u.grpcClient.GetUserProfile(ctx, req)
	u.profileResponse = resp
	u.grpcError = err

	return nil
}

func (u *UserCtx) iGetTheUserProfileWithInvalidID() error {
	if u.grpcClient == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	// Ensure Firebase authentication is set up
	if err := u.ensureFirebaseAuth(); err != nil {
		return fmt.Errorf("failed to set up Firebase authentication: %w", err)
	}

	// Use nil UUID (all zeros) which is considered invalid
	invalidID := "00000000-0000-0000-0000-000000000000"

	u.profileRequest = &userv1.GetUserProfileRequest{
		UserId: invalidID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(u.profileRequest)
	// Use real Firebase token
	req.Header().Set("Authorization", "Bearer "+u.firebaseIDToken)

	resp, err := u.grpcClient.GetUserProfile(ctx, req)
	u.profileResponse = resp
	u.grpcError = err

	return nil
}

func (u *UserCtx) iUpdateTheUserProfileWithDisplayName(displayName string) error {
	if u.grpcClient == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	if u.createdUserID == "" {
		return fmt.Errorf("no user ID available for update request")
	}

	// Make display name unique to avoid constraint violations
	uniqueDisplayName := displayName + " " + uuid.New().String()[:8]

	u.updateRequest = &userv1.UpdateUserProfileRequest{
		UserId:      u.createdUserID,
		DisplayName: &uniqueDisplayName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := connect.NewRequest(u.updateRequest)
	// Use real Firebase token
	req.Header().Set("Authorization", "Bearer "+u.firebaseIDToken)

	resp, err := u.grpcClient.UpdateUserProfile(ctx, req)
	u.updateResponse = resp
	u.grpcError = err

	return nil
}

func (u *UserCtx) iMakeAGETRequestTo(endpoint string) error {
	url := fmt.Sprintf("%s%s", u.baseURL, endpoint)
	resp, err := u.httpClient.Get(url)

	u.lastResponse = resp
	u.lastError = err

	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	return nil
}

func (u *UserCtx) iMakeAPOSTRequestToWithContentTypeAndBody(endpoint, contentType, body string) error {
	url := fmt.Sprintf("%s%s", u.baseURL, endpoint)

	// Handle escaped JSON in body
	unescapedBody := strings.ReplaceAll(body, "\\\"", "\"")

	req, err := http.NewRequest("POST", url, strings.NewReader(unescapedBody))
	if err != nil {
		u.lastError = err
		// Also set shared response if available
		if u.sharedResponse != nil {
			u.sharedResponse.Error = err
		}
		return nil
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := u.httpClient.Do(req)
	u.lastResponse = resp
	u.lastError = err

	// Also set shared response if available
	if u.sharedResponse != nil {
		u.sharedResponse.Response = resp
		u.sharedResponse.Error = err
	}

	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	return nil
}

func (u *UserCtx) iMakeConcurrentUserRegistrations(numRequests int) error {
	if u.grpcClient == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	if u.firebaseClient == nil {
		return fmt.Errorf("firebase client not initialized")
	}

	var wg sync.WaitGroup
	results := make([]ConcurrentUserResult, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		// Add delay between starting goroutines to avoid rate limiting
		time.Sleep(3 * time.Second)

		go func(index int) {
			defer wg.Done()

			// Add additional delay within goroutine to stagger Firebase requests
			time.Sleep(time.Duration(index) * 2 * time.Second)

			// Create a unique Firebase user for each concurrent request
			email := helpers.GenerateRandomEmail()
			password := helpers.GenerateSecurePassword()
			displayName := fmt.Sprintf("Concurrent User %d %s", index, uuid.New().String()[:8])

			// Sign up with Firebase
			authResp, err := u.firebaseClient.SignUpWithEmailPassword(email, password)
			if err != nil {
				results[index] = ConcurrentUserResult{
					Error:         fmt.Errorf("failed to create Firebase user: %w", err),
					OperationType: "register",
				}
				return
			}

			registerReq := &userv1.RegisterUserRequest{
				FirebaseUid:  authResp.LocalID,
				DisplayName:  displayName,
				AuthProvider: "email",
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			req := connect.NewRequest(registerReq)
			// Use real Firebase token
			req.Header().Set("Authorization", "Bearer "+authResp.IDToken)

			resp, err := u.grpcClient.RegisterUser(ctx, req)

			results[index] = ConcurrentUserResult{
				RegisterResponse: resp,
				Error:            err,
				OperationType:    "register",
			}

			// Clean up Firebase user after test
			if cleanupErr := u.firebaseClient.DeleteUser(authResp.LocalID); cleanupErr != nil {
				fmt.Printf("Warning: Failed to delete Firebase user %s: %v\n", authResp.LocalID, cleanupErr)
			}
		}(i)
	}

	wg.Wait()
	u.concurrentResults = results

	return nil
}

// Then Steps

func (u *UserCtx) theUserRegistrationShouldSucceed() error {
	if u.grpcError != nil {
		return fmt.Errorf("user registration should not return an error, got: %v", u.grpcError)
	}

	if u.registerResponse == nil {
		return fmt.Errorf("registration response should not be nil")
	}

	if u.registerResponse.Msg == nil {
		return fmt.Errorf("registration response message should not be nil")
	}

	if u.registerResponse.Msg.UserId == "" {
		return fmt.Errorf("user ID should not be empty")
	}

	// Store the created user ID for potential use in subsequent tests
	u.createdUserID = u.registerResponse.Msg.UserId
	return nil
}

func (u *UserCtx) theResponseShouldContainAValidUserID() error {
	if u.registerResponse == nil {
		return fmt.Errorf("registration response should not be nil")
	}

	if u.registerResponse.Msg == nil {
		return fmt.Errorf("registration response message should not be nil")
	}

	userID := u.registerResponse.Msg.UserId
	_, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("user ID should be a valid UUID, got: %s, error: %v", userID, err)
	}

	return nil
}

func (u *UserCtx) theUserRegistrationShouldFailWithAlreadyExistsError() error {
	if u.grpcError == nil {
		return fmt.Errorf("expected AlreadyExists error but got no error")
	}

	// Check if error contains "already exists" or "AlreadyExists"
	errorStr := u.grpcError.Error()
	if !contains(errorStr, "already exists") && !contains(errorStr, "AlreadyExists") {
		return fmt.Errorf("expected AlreadyExists error, got: %v", u.grpcError)
	}

	return nil
}

func (u *UserCtx) theUserRegistrationShouldFailWithInvalidArgumentError() error {
	if u.grpcError == nil {
		return fmt.Errorf("expected InvalidArgument error but got no error")
	}

	// Check if error contains "invalid" or "InvalidArgument"
	errorStr := u.grpcError.Error()
	if !contains(errorStr, "invalid") && !contains(errorStr, "InvalidArgument") {
		return fmt.Errorf("expected InvalidArgument error, got: %v", u.grpcError)
	}

	return nil
}

func (u *UserCtx) theUserProfileRetrievalShouldFailWithNotFoundError() error {
	if u.grpcError == nil {
		return fmt.Errorf("expected NotFound error but got no error")
	}

	// Check if error contains "not found" or "NotFound"
	errorStr := u.grpcError.Error()
	if !contains(errorStr, "not found") && !contains(errorStr, "NotFound") {
		return fmt.Errorf("expected NotFound error, got: %v", u.grpcError)
	}

	return nil
}

func (u *UserCtx) theUserProfileRetrievalShouldFailWithPermissionDeniedError() error {
	if u.grpcError == nil {
		return fmt.Errorf("expected PermissionDenied error but got no error")
	}

	// Check if error contains "permission denied" or "PermissionDenied"
	errorStr := u.grpcError.Error()
	if !strings.Contains(strings.ToLower(errorStr), "permission") &&
		!strings.Contains(errorStr, "PermissionDenied") {
		return fmt.Errorf("expected PermissionDenied error, got: %v", u.grpcError)
	}

	return nil
}

func (u *UserCtx) theUserProfileShouldBeRetrieved() error {
	if u.grpcError != nil {
		return fmt.Errorf("user profile retrieval should not return an error, got: %v", u.grpcError)
	}

	if u.profileResponse == nil {
		return fmt.Errorf("profile response should not be nil")
	}

	if u.profileResponse.Msg == nil {
		return fmt.Errorf("profile response message should not be nil")
	}

	if u.profileResponse.Msg.UserId == "" {
		return fmt.Errorf("profile user ID should not be empty")
	}

	return nil
}

func (u *UserCtx) theUserProfileShouldContainDisplayName(expectedDisplayName string) error {
	if u.profileResponse == nil || u.profileResponse.Msg == nil {
		return fmt.Errorf("profile response should not be nil")
	}

	if u.profileResponse.Msg.DisplayName != expectedDisplayName {
		return fmt.Errorf("expected display name %s, got %s", expectedDisplayName, u.profileResponse.Msg.DisplayName)
	}

	return nil
}

func (u *UserCtx) theUserProfileUpdateShouldSucceed() error {
	if u.grpcError != nil {
		return fmt.Errorf("user profile update should not return an error, got: %v", u.grpcError)
	}

	if u.updateResponse == nil {
		return fmt.Errorf("update response should not be nil")
	}

	if u.updateResponse.Msg == nil {
		return fmt.Errorf("update response message should not be nil")
	}

	return nil
}

// Removed unused step: theHTTPStatusShouldBe

func (u *UserCtx) allConcurrentUserRegistrationsShouldSucceed() error {
	if len(u.concurrentResults) == 0 {
		return fmt.Errorf("no concurrent results available")
	}

	for i, result := range u.concurrentResults {
		if result.Error != nil {
			return fmt.Errorf("concurrent request %d failed: %v", i, result.Error)
		}

		if result.RegisterResponse == nil {
			return fmt.Errorf("concurrent request %d: response should not be nil", i)
		}

		if result.RegisterResponse.Msg == nil {
			return fmt.Errorf("concurrent request %d: response message should not be nil", i)
		}

		if result.RegisterResponse.Msg.UserId == "" {
			return fmt.Errorf("concurrent request %d: user ID should not be empty", i)
		}
	}

	return nil
}

func (u *UserCtx) iMakeConcurrentPOSTRequestsToWithContentTypeAndBody(numRequests int, endpoint, contentType, body string) error {
	var wg sync.WaitGroup
	results := make([]ConcurrentHTTPResult, numRequests)

	// Handle escaped JSON in body
	unescapedBody := strings.ReplaceAll(body, "\\\"", "\"")

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			url := fmt.Sprintf("%s%s", u.baseURL, endpoint)

			req, err := http.NewRequest("POST", url, strings.NewReader(unescapedBody))
			if err != nil {
				results[index] = ConcurrentHTTPResult{Error: err}
				return
			}

			req.Header.Set("Content-Type", contentType)

			resp, err := u.httpClient.Do(req)
			results[index] = ConcurrentHTTPResult{
				Response: resp,
				Error:    err,
			}

			if resp != nil {
				defer func() { _ = resp.Body.Close() }()
			}
		}(i)
	}

	wg.Wait()
	u.concurrentHTTPResults = results

	return nil
}

func (u *UserCtx) allConcurrentRequestsShouldReturnHTTPStatus(expectedStatus int) error {
	if len(u.concurrentHTTPResults) == 0 {
		return fmt.Errorf("no concurrent HTTP results available")
	}

	for i, result := range u.concurrentHTTPResults {
		if result.Error != nil {
			return fmt.Errorf("concurrent request %d failed: %v", i, result.Error)
		}

		if result.Response == nil {
			return fmt.Errorf("concurrent request %d: response should not be nil", i)
		}

		if result.Response.StatusCode != expectedStatus {
			return fmt.Errorf("concurrent request %d: expected HTTP status %d, got %d", i, expectedStatus, result.Response.StatusCode)
		}
	}

	return nil
}

// Additional step definitions for JSON body patterns that godog auto-generates

func (u *UserCtx) iMakeAPOSTRequestToWithJSONBody2Fields(endpoint, contentType, key1, value1 string) error {
	// Reconstruct JSON body from individual fields
	body := fmt.Sprintf(`{"%s":"%s"}`, key1, value1)
	return u.iMakeAPOSTRequestToWithContentTypeAndBody(endpoint, contentType, body)
}

func (u *UserCtx) iMakeAPOSTRequestToWithJSONBody4Fields(endpoint, contentType, key1, value1, key2, value2 string) error {
	// Reconstruct JSON body from individual fields
	body := fmt.Sprintf(`{"%s":"%s","%s":"%s"}`, key1, value1, key2, value2)
	return u.iMakeAPOSTRequestToWithContentTypeAndBody(endpoint, contentType, body)
}

func (u *UserCtx) iMakeAPOSTRequestToWithJSONBody6Fields(endpoint, contentType, key1, value1, key2, value2, key3, value3 string) error {
	// Reconstruct JSON body from individual fields
	body := fmt.Sprintf(`{"%s":"%s","%s":"%s","%s":"%s"}`, key1, value1, key2, value2, key3, value3)
	return u.iMakeAPOSTRequestToWithContentTypeAndBody(endpoint, contentType, body)
}

func (u *UserCtx) iMakeConcurrentPOSTRequestsToWithJSONBody6Fields(numRequests int, endpoint, contentType, key1, value1, key2, value2, key3, value3 string) error {
	// Reconstruct JSON body from individual fields
	body := fmt.Sprintf(`{"%s":"%s","%s":"%s","%s":"%s"}`, key1, value1, key2, value2, key3, value3)
	return u.iMakeConcurrentPOSTRequestsToWithContentTypeAndBody(numRequests, endpoint, contentType, body)
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Register registers all user-related step definitions
func (u *UserCtx) Register(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the server is running$`, u.theServerIsRunning)
	sc.Step(`^the gRPC user client is connected$`, u.theGRPCUserClientIsConnected)
	sc.Step(`^the gRPC-Web user client is connected$`, u.theGRPCWebUserClientIsConnected)
	sc.Step(`^I have valid user registration data$`, u.iHaveValidUserRegistrationData)
	sc.Step(`^I have invalid user registration data with empty Firebase UID$`, u.iHaveInvalidUserRegistrationDataWithEmptyFirebaseUID)
	sc.Step(`^an existing user is registered$`, u.anExistingUserIsRegistered)

	// When steps
	sc.Step(`^I register a user with valid data$`, u.iRegisterAUserWithValidData)
	sc.Step(`^I register a user with the same Firebase UID$`, u.iRegisterAUserWithTheSameFirebaseUID)
	sc.Step(`^I register a user with empty Firebase UID$`, u.iRegisterAUserWithEmptyFirebaseUID)
	sc.Step(`^I get the user profile$`, u.iGetTheUserProfile)
	sc.Step(`^I get the user profile with invalid ID$`, u.iGetTheUserProfileWithInvalidID)
	sc.Step(`^I get the user profile with non-existent ID$`, u.iGetTheUserProfileWithNonExistentID)
	sc.Step(`^I update the user profile with display name "([^"]*)"$`, u.iUpdateTheUserProfileWithDisplayName)
	sc.Step(`^I make a GET request to "([^"]*)"$`, u.iMakeAGETRequestTo)
	sc.Step(`^I make a POST request to "([^"]*)" with content type "([^"]*)" and body "([^"]*)"$`, u.iMakeAPOSTRequestToWithContentTypeAndBody)
	sc.Step(`^I make (\d+) concurrent user registrations$`, u.iMakeConcurrentUserRegistrations)
	sc.Step(`^I make (\d+) concurrent POST requests to "([^"]*)" with content type "([^"]*)" and body "([^"]*)"$`, u.iMakeConcurrentPOSTRequestsToWithContentTypeAndBody)

	// Additional step definitions for complex JSON patterns that godog auto-generates
	// These handle the specific JSON structures in the feature files
	sc.Step(`^I make a POST request to "([^"]*)" with content type "([^"]*)" and body "{\\"([^"]*)":\\"([^"]*)"}"$`, u.iMakeAPOSTRequestToWithJSONBody2Fields)
	sc.Step(`^I make a POST request to "([^"]*)" with content type "([^"]*)" and body "{\\"([^"]*)":\\"([^"]*)",\\"([^"]*)":\\"([^"]*)"}"$`, u.iMakeAPOSTRequestToWithJSONBody4Fields)
	sc.Step(`^I make a POST request to "([^"]*)" with content type "([^"]*)" and body "{\\"([^"]*)":\\"([^"]*)",\\"([^"]*)":\\"([^"]*)",\\"([^"]*)":\\"([^"]*)"}"$`, u.iMakeAPOSTRequestToWithJSONBody6Fields)
	sc.Step(`^I make (\d+) concurrent POST requests to "([^"]*)" with content type "([^"]*)" and body "{\\"([^"]*)":\\"([^"]*)",\\"([^"]*)":\\"([^"]*)",\\"([^"]*)":\\"([^"]*)"}"$`, u.iMakeConcurrentPOSTRequestsToWithJSONBody6Fields)

	// Then steps
	sc.Step(`^the user registration should succeed$`, u.theUserRegistrationShouldSucceed)
	sc.Step(`^the response should contain a valid user ID$`, u.theResponseShouldContainAValidUserID)
	sc.Step(`^the user registration should fail with AlreadyExists error$`, u.theUserRegistrationShouldFailWithAlreadyExistsError)
	sc.Step(`^the user registration should fail with InvalidArgument error$`, u.theUserRegistrationShouldFailWithInvalidArgumentError)
	sc.Step(`^the user profile retrieval should fail with NotFound error$`, u.theUserProfileRetrievalShouldFailWithNotFoundError)
	sc.Step(`^the user profile retrieval should fail with PermissionDenied error$`, u.theUserProfileRetrievalShouldFailWithPermissionDeniedError)
	sc.Step(`^the user profile should be retrieved$`, u.theUserProfileShouldBeRetrieved)
	sc.Step(`^the user profile should contain display name "([^"]*)"$`, u.theUserProfileShouldContainDisplayName)
	sc.Step(`^the user profile update should succeed$`, u.theUserProfileUpdateShouldSucceed)
	// Note: HTTP status checks are handled by HealthCtx to avoid conflicts
	sc.Step(`^all concurrent user registrations should succeed$`, u.allConcurrentUserRegistrationsShouldSucceed)
	sc.Step(`^all concurrent requests should return HTTP status (\d+)$`, u.allConcurrentRequestsShouldReturnHTTPStatus)
}
