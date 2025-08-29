package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

// FirebaseAuthClient provides methods to interact with Firebase Auth REST API
// This client is used for E2E testing to create real Firebase users and obtain ID tokens
type FirebaseAuthClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// SignUpRequest represents the request payload for Firebase Auth sign up
type SignUpRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

// SignInRequest represents the request payload for Firebase Auth sign in
type SignInRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

// AuthResponse represents the response from Firebase Auth API
type AuthResponse struct {
	IDToken      string `json:"idToken"`
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"` // This is the Firebase UID
}

// ErrorResponse represents an error response from Firebase Auth API
type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Message string `json:"message"`
			Domain  string `json:"domain"`
			Reason  string `json:"reason"`
		} `json:"errors"`
	} `json:"error"`
}

// NewFirebaseAuthClient creates a new Firebase Auth REST API client
func NewFirebaseAuthClient() (*FirebaseAuthClient, error) {
	apiKey := os.Getenv("FIREBASE_WEB_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("FIREBASE_WEB_API_KEY environment variable is required")
	}

	return &FirebaseAuthClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://identitytoolkit.googleapis.com/v1",
	}, nil
}

// SignUpWithEmailPassword creates a new Firebase user with email and password
func (c *FirebaseAuthClient) SignUpWithEmailPassword(email, password string) (*AuthResponse, error) {
	url := fmt.Sprintf("%s/accounts:signUp?key=%s", c.baseURL, c.apiKey)

	payload := SignUpRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}

	return c.makeAuthRequestWithRetry(url, payload)
}

// SignInWithEmailPassword signs in an existing Firebase user with email and password
func (c *FirebaseAuthClient) SignInWithEmailPassword(email, password string) (*AuthResponse, error) {
	url := fmt.Sprintf("%s/accounts:signInWithPassword?key=%s", c.baseURL, c.apiKey)

	payload := SignInRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}

	return c.makeAuthRequestWithRetry(url, payload)
}

// DeleteUser deletes a Firebase user by their ID token
func (c *FirebaseAuthClient) DeleteUser(idToken string) error {
	url := fmt.Sprintf("%s/accounts:delete?key=%s", c.baseURL, c.apiKey)

	payload := map[string]interface{}{
		"idToken": idToken,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete user, status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// isRateLimitError checks if the error is a rate limiting error
func isRateLimitError(errorMsg string) bool {
	rateLimitMessages := []string{
		"TOO_MANY_ATTEMPTS_TRY_LATER",
		"QUOTA_EXCEEDED",
		"RATE_LIMIT_EXCEEDED",
		"too many requests",
	}

	lowerMsg := strings.ToLower(errorMsg)
	for _, msg := range rateLimitMessages {
		if strings.Contains(lowerMsg, strings.ToLower(msg)) {
			return true
		}
	}
	return false
}

// makeAuthRequestWithRetry makes an HTTP request to Firebase Auth API with exponential backoff retry
func (c *FirebaseAuthClient) makeAuthRequestWithRetry(url string, payload interface{}) (*AuthResponse, error) {
	const maxRetries = 10
	const baseDelay = 3 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter - more aggressive delays
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
			// Add more jitter to avoid thundering herd
			jitter := time.Duration(float64(delay) * 0.3 * (0.5 + 0.5*float64(time.Now().UnixNano()%1000)/1000))
			totalDelay := delay + jitter

			// Cap maximum delay at 60 seconds
			if totalDelay > 60*time.Second {
				totalDelay = 60*time.Second + jitter
			}

			fmt.Printf("Firebase rate limit hit, retrying in %v (attempt %d/%d)\n", totalDelay, attempt+1, maxRetries+1)
			time.Sleep(totalDelay)
		}

		resp, err := c.makeAuthRequest(url, payload)
		if err != nil {
			// Check if this is a rate limiting error
			if isRateLimitError(err.Error()) && attempt < maxRetries {
				fmt.Printf("Rate limit detected: %v, will retry...\n", err)
				continue // Retry
			}
			return nil, err // Non-retryable error or max retries exceeded
		}

		return resp, nil // Success
	}

	return nil, fmt.Errorf("max retries exceeded for Firebase Auth request after %d attempts", maxRetries+1)
}

// makeAuthRequest makes an HTTP request to Firebase Auth API and returns the parsed response
func (c *FirebaseAuthClient) makeAuthRequest(url string, payload interface{}) (*AuthResponse, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("firebase auth error: %s", errorResp.Error.Message)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &authResp, nil
}

// GenerateRandomEmail generates a random email address for testing
func GenerateRandomEmail() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("test-user-%d@example.com", timestamp)
}

// GenerateSecurePassword generates a secure password for testing
func GenerateSecurePassword() string {
	return "TestPassword123!"
}
