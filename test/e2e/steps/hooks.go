package steps

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

// SharedHTTPResponse holds HTTP response data that can be shared between contexts
type SharedHTTPResponse struct {
	Response *http.Response
	Body     []byte
	Error    error
}

var (
	healthProtocol = flag.String("health_protocol", "rest", "Protocol to test (rest or grpc)")
	userProtocol   = flag.String("user_protocol", "grpc", "Protocol to test (rest or grpc)")
)

// InitializeScenario initializes the scenario context with health and user steps
func InitializeScenario(sc *godog.ScenarioContext) {
	// Create contexts
	healthCtx := NewHealthCtx()
	userCtx := NewUserCtx()
	userAuthCtx := NewUserAuthCtx()
	quizCtx := NewQuizCtx()

	// Create shared HTTP response holder
	sharedResponse := &SharedHTTPResponse{}

	// Set up hooks
	sc.Before(func(ctx context.Context, scn *godog.Scenario) (context.Context, error) {
		// Log scenario start
		fmt.Printf("Starting scenario: %s\n", scn.Name)

		// Add delay between scenarios only for user tests that actually create Firebase users
		// Health tests and some user tests don't need delays
		needsFirebaseDelay := false
		isUserTest := false

		for _, tag := range scn.Tags {
			if tag.Name == "@user" || tag.Name == "@user_auth" {
				isUserTest = true
			}
			// Skip delay for scenarios that don't create Firebase users
			if tag.Name == "@firebase_rate_limit" || tag.Name == "@negative" {
				needsFirebaseDelay = false
				break
			}
		}

		// Only apply delay for positive user tests that actually CREATE Firebase users
		if isUserTest && !needsFirebaseDelay {
			// Check if scenario involves Firebase user CREATION (not just usage)
			scenarioText := strings.ToLower(scn.Name)

			// Scenarios that definitely create Firebase users
			if strings.Contains(scenarioText, "registration") ||
				strings.Contains(scenarioText, "register") ||
				strings.Contains(scenarioText, "sign up") ||
				strings.Contains(scenarioText, "signup") {
				needsFirebaseDelay = true
			}

			// Special case: "existing user is registered" creates a user in the background
			// but "get user profile" or "update user profile" just use existing users
			if strings.Contains(scenarioText, "existing user") && strings.Contains(scenarioText, "registered") {
				needsFirebaseDelay = true
			}
		}

		if needsFirebaseDelay {
			fmt.Printf("Adding Firebase rate limiting delay for user scenario...\n")
			time.Sleep(1 * time.Second) // Reduced from 2 seconds to 1 second
		}

		// Reset context state for each scenario by reinitializing the same instances
		// instead of creating new ones
		*healthCtx = *NewHealthCtx()
		*userCtx = *NewUserCtx()
		*userAuthCtx = *NewUserAuthCtx()
		*quizCtx = *NewQuizCtx()

		// Reset shared response
		*sharedResponse = SharedHTTPResponse{}

		// Set shared response in both contexts
		healthCtx.sharedResponse = sharedResponse
		userCtx.sharedResponse = sharedResponse
		// quizCtx does not currently use sharedResponse

		return ctx, nil
	})

	sc.After(func(ctx context.Context, scn *godog.Scenario, err error) (context.Context, error) {
		// Cleanup resources
		healthCtx.Cleanup()
		userCtx.Cleanup()
		userAuthCtx.Cleanup()
		quizCtx.Cleanup()

		// Log scenario completion
		if err != nil {
			fmt.Printf("Scenario failed: %s - %v\n", scn.Name, err)
		} else {
			fmt.Printf("Scenario passed: %s\n", scn.Name)
		}

		return ctx, nil
	})

	// Register step definitions
	healthCtx.Register(sc)
	userCtx.Register(sc)
	userAuthCtx.Register(sc)
	quizCtx.Register(sc)
}

// GetHealthProtocol returns the current health protocol being tested
func GetHealthProtocol() string {
	if healthProtocol == nil {
		return "rest"
	}
	return *healthProtocol
}

// GetUserProtocol returns the current user protocol being tested
func GetUserProtocol() string {
	if userProtocol == nil {
		return "grpc"
	}
	return *userProtocol
}

// SetupTestEnvironment sets up the test environment variables
func SetupTestEnvironment() {
	// Set default environment variables if not already set
	if os.Getenv("BASE_URL") == "" {
		_ = os.Setenv("BASE_URL", "http://localhost:8080")
	}

	if os.Getenv("GRPC_ADDR") == "" {
		_ = os.Setenv("GRPC_ADDR", "localhost:8080")
	}
}
