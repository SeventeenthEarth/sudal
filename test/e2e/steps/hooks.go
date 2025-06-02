package steps

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

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

	// Create shared HTTP response holder
	sharedResponse := &SharedHTTPResponse{}

	// Set up hooks
	sc.Before(func(ctx context.Context, scn *godog.Scenario) (context.Context, error) {
		// Log scenario start
		fmt.Printf("Starting scenario: %s\n", scn.Name)

		// Reset context state for each scenario by reinitializing the same instances
		// instead of creating new ones
		*healthCtx = *NewHealthCtx()
		*userCtx = *NewUserCtx()

		// Reset shared response
		*sharedResponse = SharedHTTPResponse{}

		// Set shared response in both contexts
		healthCtx.sharedResponse = sharedResponse
		userCtx.sharedResponse = sharedResponse

		return ctx, nil
	})

	sc.After(func(ctx context.Context, scn *godog.Scenario, err error) (context.Context, error) {
		// Cleanup resources
		healthCtx.Cleanup()
		userCtx.Cleanup()

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
		os.Setenv("BASE_URL", "http://localhost:8080")
	}

	if os.Getenv("GRPC_ADDR") == "" {
		os.Setenv("GRPC_ADDR", "localhost:8080")
	}
}
