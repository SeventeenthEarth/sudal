package steps

import (
	"encoding/json"
	"regexp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Database health specific Then Steps

// ThenJSONResponseShouldContainDatabaseInformation checks for database information in response
func ThenJSONResponseShouldContainDatabaseInformation(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Response is not valid JSON")

	database, exists := jsonData["database"]
	require.True(ctx.T, exists, "Response does not contain 'database' field")

	databaseInfo, ok := database.(map[string]interface{})
	require.True(ctx.T, ok, "Database field should be an object")

	_, statusExists := databaseInfo["status"]
	assert.True(ctx.T, statusExists, "Database info does not contain 'status' field")

	_, messageExists := databaseInfo["message"]
	assert.True(ctx.T, messageExists, "Database info does not contain 'message' field")
}

// ThenJSONResponseShouldContainConnectionStatistics checks for connection statistics
func ThenJSONResponseShouldContainConnectionStatistics(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Response is not valid JSON")

	database, exists := jsonData["database"]
	require.True(ctx.T, exists, "Response does not contain 'database' field")

	databaseInfo, ok := database.(map[string]interface{})
	require.True(ctx.T, ok, "Database field should be an object")

	stats, statsExists := databaseInfo["stats"]
	require.True(ctx.T, statsExists, "Database info does not contain 'stats' field")

	statsMap, ok := stats.(map[string]interface{})
	require.True(ctx.T, ok, "Stats field should be an object")

	// Check for expected statistics fields
	expectedStats := []string{
		"max_open_connections",
		"open_connections",
		"in_use",
		"idle",
		"wait_count",
		"wait_duration",
		"max_idle_closed",
		"max_lifetime_closed",
	}

	for _, statField := range expectedStats {
		value, exists := statsMap[statField]
		assert.True(ctx.T, exists, "Stats does not contain '%s' field", statField)

		// Check if value is a number (int or float)
		switch v := value.(type) {
		case int, int64, float64:
			// Valid number type
		default:
			ctx.T.Errorf("Stat '%s' should be a number, got %T", statField, v)
		}
	}
}

// ThenJSONResponseShouldContainTimestampField checks for timestamp field
func ThenJSONResponseShouldContainTimestampField(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Response is not valid JSON")

	timestamp, exists := jsonData["timestamp"]
	require.True(ctx.T, exists, "Response does not contain 'timestamp' field")

	timestampStr, ok := timestamp.(string)
	require.True(ctx.T, ok, "Timestamp should be a string")
	assert.NotEmpty(ctx.T, timestampStr, "Timestamp should not be empty")

	// Basic ISO 8601 format check (YYYY-MM-DDTHH:MM:SSZ)
	isoPattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`
	matched, err := regexp.MatchString(isoPattern, timestampStr)
	require.NoError(ctx.T, err, "Failed to compile regex pattern")
	assert.True(ctx.T, matched,
		"Timestamp should be in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ), got: %s", timestampStr)
}

// ThenDatabaseConnectionPoolShouldBeHealthy checks that connection pool is healthy
func ThenDatabaseConnectionPoolShouldBeHealthy(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Response is not valid JSON")

	database, exists := jsonData["database"]
	require.True(ctx.T, exists, "Response does not contain 'database' field")

	databaseInfo, ok := database.(map[string]interface{})
	require.True(ctx.T, ok, "Database field should be an object")

	status, statusExists := databaseInfo["status"]
	require.True(ctx.T, statusExists, "Database info does not contain 'status' field")
	assert.Equal(ctx.T, "healthy", status, "Expected database status to be 'healthy', got '%v'", status)

	stats, statsExists := databaseInfo["stats"]
	require.True(ctx.T, statsExists, "Database info does not contain 'stats' field")

	statsMap, ok := stats.(map[string]interface{})
	require.True(ctx.T, ok, "Stats field should be an object")

	// Check that we have at least some open connections configured
	maxOpenConns, exists := statsMap["max_open_connections"]
	require.True(ctx.T, exists, "Stats does not contain 'max_open_connections'")

	maxOpenConnsFloat, ok := maxOpenConns.(float64)
	if !ok {
		if maxOpenConnsInt, ok := maxOpenConns.(int); ok {
			maxOpenConnsFloat = float64(maxOpenConnsInt)
		} else {
			ctx.T.Fatalf("max_open_connections should be a number, got %T", maxOpenConns)
		}
	}
	assert.Greater(ctx.T, maxOpenConnsFloat, float64(0), "Max open connections should be greater than 0")

	// Check that open connections doesn't exceed max
	openConns, exists := statsMap["open_connections"]
	require.True(ctx.T, exists, "Stats does not contain 'open_connections'")

	openConnsFloat, ok := openConns.(float64)
	if !ok {
		if openConnsInt, ok := openConns.(int); ok {
			openConnsFloat = float64(openConnsInt)
		} else {
			ctx.T.Fatalf("open_connections should be a number, got %T", openConns)
		}
	}
	assert.LessOrEqual(ctx.T, openConnsFloat, maxOpenConnsFloat,
		"Open connections should not exceed max open connections")
}

// ThenConnectionStatisticsShouldBeValid checks that connection statistics are valid and consistent
func ThenConnectionStatisticsShouldBeValid(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Response is not valid JSON")

	database, exists := jsonData["database"]
	require.True(ctx.T, exists, "Response does not contain 'database' field")

	databaseInfo, ok := database.(map[string]interface{})
	require.True(ctx.T, ok, "Database field should be an object")

	stats, statsExists := databaseInfo["stats"]
	require.True(ctx.T, statsExists, "Database info does not contain 'stats' field")

	statsMap, ok := stats.(map[string]interface{})
	require.True(ctx.T, ok, "Stats field should be an object")

	// Helper function to convert interface{} to float64
	toFloat64 := func(v interface{}) float64 {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		default:
			ctx.T.Fatalf("Expected number, got %T", v)
			return 0
		}
	}

	// Validate that statistics are consistent
	openConns := toFloat64(statsMap["open_connections"])
	inUse := toFloat64(statsMap["in_use"])
	idle := toFloat64(statsMap["idle"])

	// Open connections should equal in_use + idle
	assert.Equal(ctx.T, openConns, inUse+idle,
		"Open connections (%.0f) should equal in_use (%.0f) + idle (%.0f)", openConns, inUse, idle)

	// All counts should be non-negative
	nonNegativeFields := []string{
		"open_connections",
		"in_use",
		"idle",
		"wait_count",
		"max_idle_closed",
		"max_lifetime_closed",
	}

	for _, field := range nonNegativeFields {
		value := toFloat64(statsMap[field])
		assert.GreaterOrEqual(ctx.T, value, float64(0), "%s should be non-negative, got %.0f", field, value)
	}

	// Wait duration should be non-negative
	waitDuration := toFloat64(statsMap["wait_duration"])
	assert.GreaterOrEqual(ctx.T, waitDuration, float64(0),
		"Wait duration should be non-negative, got %.0f", waitDuration)
}

// ThenConnectionStatisticsShouldIncludeMaxOpenConnections checks for max_open_connections stat
func ThenConnectionStatisticsShouldIncludeMaxOpenConnections(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Response is not valid JSON")

	database, exists := jsonData["database"]
	require.True(ctx.T, exists, "Response does not contain 'database' field")

	databaseInfo, ok := database.(map[string]interface{})
	require.True(ctx.T, ok, "Database field should be an object")

	stats, statsExists := databaseInfo["stats"]
	require.True(ctx.T, statsExists, "Database info does not contain 'stats' field")

	statsMap, ok := stats.(map[string]interface{})
	require.True(ctx.T, ok, "Stats field should be an object")

	maxOpenConns, exists := statsMap["max_open_connections"]
	require.True(ctx.T, exists, "Stats does not contain 'max_open_connections'")

	// Check if value is a number
	switch v := maxOpenConns.(type) {
	case int, int64, float64:
		// Valid number type
		var value float64
		if intVal, ok := v.(int); ok {
			value = float64(intVal)
		} else if int64Val, ok := v.(int64); ok {
			value = float64(int64Val)
		} else {
			value = v.(float64)
		}
		assert.Greater(ctx.T, value, float64(0), "max_open_connections should be greater than 0")
	default:
		ctx.T.Errorf("max_open_connections should be a number, got %T", v)
	}
}

// ThenConnectionStatisticsShouldIncludeCurrentUsageMetrics checks for current usage metrics
func ThenConnectionStatisticsShouldIncludeCurrentUsageMetrics(ctx *TestContext) {
	require.NotNil(ctx.T, ctx.Response, "No response received")
	require.NotEmpty(ctx.T, ctx.ResponseBody, "Response body is empty")

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	require.NoError(ctx.T, err, "Response is not valid JSON")

	database, exists := jsonData["database"]
	require.True(ctx.T, exists, "Response does not contain 'database' field")

	databaseInfo, ok := database.(map[string]interface{})
	require.True(ctx.T, ok, "Database field should be an object")

	stats, statsExists := databaseInfo["stats"]
	require.True(ctx.T, statsExists, "Database info does not contain 'stats' field")

	statsMap, ok := stats.(map[string]interface{})
	require.True(ctx.T, ok, "Stats field should be an object")

	// Check for current usage metrics
	usageMetrics := []string{"open_connections", "in_use", "idle"}
	for _, metric := range usageMetrics {
		value, exists := statsMap[metric]
		assert.True(ctx.T, exists, "Stats does not contain '%s'", metric)

		// Check if value is a number and non-negative
		switch v := value.(type) {
		case int, int64, float64:
			var numValue float64
			if intVal, ok := v.(int); ok {
				numValue = float64(intVal)
			} else if int64Val, ok := v.(int64); ok {
				numValue = float64(int64Val)
			} else {
				numValue = v.(float64)
			}
			assert.GreaterOrEqual(ctx.T, numValue, float64(0), "%s should be non-negative", metric)
		default:
			ctx.T.Errorf("%s should be a number, got %T", metric, v)
		}
	}
}

// ThenAllDatabaseHealthRequestsShouldSucceed checks that all concurrent database health requests succeeded
func ThenAllDatabaseHealthRequestsShouldSucceed(ctx *TestContext) {
	require.NotEmpty(ctx.T, ctx.ConcurrentResults, "No concurrent results found")

	for i, result := range ctx.ConcurrentResults {
		assert.NoError(ctx.T, result.Error, "Request %d failed with error", i+1)
		assert.NotNil(ctx.T, result.Response, "Request %d has no response", i+1)
		assert.Equal(ctx.T, 200, result.Response.StatusCode,
			"Request %d expected status 200, got %d", i+1, result.Response.StatusCode)
	}
}

// ThenAllResponsesShouldContainValidConnectionStatistics checks all concurrent responses for valid stats
func ThenAllResponsesShouldContainValidConnectionStatistics(ctx *TestContext) {
	require.NotEmpty(ctx.T, ctx.ConcurrentResults, "No concurrent results found")

	for i, result := range ctx.ConcurrentResults {
		assert.NoError(ctx.T, result.Error, "Request %d failed with error", i+1)
		assert.NotNil(ctx.T, result.Response, "Request %d has no response", i+1)
		assert.NotEmpty(ctx.T, result.Body, "Request %d has empty response body", i+1)

		var jsonData map[string]interface{}
		err := json.Unmarshal(result.Body, &jsonData)
		assert.NoError(ctx.T, err, "Request %d response is not valid JSON", i+1)

		database, exists := jsonData["database"]
		assert.True(ctx.T, exists, "Request %d response does not contain 'database' field", i+1)

		databaseInfo, ok := database.(map[string]interface{})
		assert.True(ctx.T, ok, "Request %d database field should be an object", i+1)

		stats, statsExists := databaseInfo["stats"]
		assert.True(ctx.T, statsExists, "Request %d database info does not contain 'stats' field", i+1)

		statsMap, ok := stats.(map[string]interface{})
		assert.True(ctx.T, ok, "Request %d stats field should be an object", i+1)

		// Basic validation of stats structure
		requiredFields := []string{"max_open_connections", "open_connections", "in_use", "idle"}
		for _, field := range requiredFields {
			value, exists := statsMap[field]
			assert.True(ctx.T, exists, "Request %d stats does not contain '%s' field", i+1, field)

			// Check if value is a number
			switch value.(type) {
			case int, int64, float64:
				// Valid number type
			default:
				ctx.T.Errorf("Request %d %s should be a number, got %T", i+1, field, value)
			}
		}
	}
}
