package e2e

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/seventeenthearth/sudal/test/e2e/steps"
)

var (
	godogFormat        = flag.String("godog.format", "pretty", "Set godog format to use")
	godogStopOnFailure = flag.Bool("godog.stop-on-failure", false, "Stop when first step fails")
	godogStrict        = flag.Bool("godog.strict", false, "Fail on pending or undefined steps")
	godogTags          = flag.String("godog.tags", "", "Set godog tags to run")
)

func TestFeatures(t *testing.T) {
	// Get the protocols to test
	healthProtocol := os.Getenv("HEALTH_PROTOCOL")
	if healthProtocol == "" {
		healthProtocol = "rest" // default
	}

	userProtocol := os.Getenv("USER_PROTOCOL")
	if userProtocol == "" {
		userProtocol = "grpc" // default
	}

	// Set up godog options
	opts := godog.Options{
		Format:        *godogFormat,
		Paths:         getFeaturePaths(),
		StopOnFailure: *godogStopOnFailure,
		Strict:        *godogStrict,
		Tags:          *godogTags,
		TestingT:      t,
	}

	// Create test suite
	suite := godog.TestSuite{
		Name:                "sudal-e2e",
		ScenarioInitializer: steps.InitializeScenario,
		Options:             &opts,
	}

	// Run the test suite
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

// getFeaturePaths returns the paths to feature files
func getFeaturePaths() []string {
	// Get the directory where this test file is located
	testDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Warning: Could not get working directory: %v\n", err)
		return []string{"features"}
	}

	// Look for feature files in the godog directory
	featureDir := filepath.Join(testDir, "features")

	// Check if features directory exists
	if _, err := os.Stat(featureDir); os.IsNotExist(err) {
		// If no features directory exists, create a default path
		// This allows the test to run even without feature files
		fmt.Printf("Features directory not found at %s, using default path\n", featureDir)
		return []string{"features"}
	}

	return []string{featureDir}
}

// TestMain handles the test execution for both health and user protocols
func TestMain(m *testing.M) {
	flag.Parse()

	// Check if we should run for specific protocols
	healthProtocol := os.Getenv("HEALTH_PROTOCOL")
	userProtocol := os.Getenv("USER_PROTOCOL")

	if healthProtocol == "" && userProtocol == "" {
		// Run tests for all protocol combinations
		fmt.Println("Running tests for all protocol combinations...")

		var results []int

		// Health REST tests
		fmt.Println("\n=== Running Health REST Protocol Tests ===")
		os.Setenv("HEALTH_PROTOCOL", "rest")
		os.Setenv("USER_PROTOCOL", "")
		results = append(results, m.Run())

		// Health gRPC tests
		fmt.Println("\n=== Running Health gRPC Protocol Tests ===")
		os.Setenv("HEALTH_PROTOCOL", "grpc")
		os.Setenv("USER_PROTOCOL", "")
		results = append(results, m.Run())

		// User gRPC tests
		fmt.Println("\n=== Running User gRPC Protocol Tests ===")
		os.Setenv("HEALTH_PROTOCOL", "")
		os.Setenv("USER_PROTOCOL", "grpc")
		results = append(results, m.Run())

		// User REST tests (negative)
		fmt.Println("\n=== Running User REST Protocol Tests (Negative) ===")
		os.Setenv("HEALTH_PROTOCOL", "")
		os.Setenv("USER_PROTOCOL", "rest")
		results = append(results, m.Run())

		// Check if any tests failed
		for i, result := range results {
			if result != 0 {
				fmt.Printf("Test suite %d failed with code %d\n", i+1, result)
				os.Exit(1)
			}
		}

		fmt.Println("All protocol tests passed!")
		os.Exit(0)
	} else {
		// Run tests for specified protocols only
		fmt.Printf("Running tests for specified protocols - Health: %s, User: %s\n", healthProtocol, userProtocol)
		result := m.Run()
		os.Exit(result)
	}
}

// Example of how to run specific scenarios programmatically
func runHealthScenarios() {
	opts := godog.Options{
		Format: "pretty",
		Paths:  []string{"features"},
		Output: colors.Colored(os.Stdout),
	}

	status := godog.TestSuite{
		Name:                "health-scenarios",
		ScenarioInitializer: steps.InitializeScenario,
		Options:             &opts,
	}.Run()

	if status == 2 {
		fmt.Println("Tests failed due to non-zero status")
		os.Exit(1)
	}

	if status == 1 {
		fmt.Println("Tests failed")
		os.Exit(1)
	}

	fmt.Println("All tests passed!")
}
