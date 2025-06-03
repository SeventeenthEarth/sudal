package e2e

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"

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

	// For user tests, force sequential execution to avoid Firebase rate limiting
	if userProtocol != "" {
		opts.Concurrency = 1 // Force sequential execution for user tests
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

	// Simply run all tests once - the script handles protocol separation
	fmt.Println("Running E2E tests...")
	result := m.Run()
	os.Exit(result)
}
