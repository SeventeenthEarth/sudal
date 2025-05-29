package steps

import (
	"encoding/json"
	"regexp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Database health specific Then Steps - BDD Style

// ThenJSONResponseShouldContainDatabaseInformation checks for database information in response in BDD style
func ThenJSONResponseShouldContainDatabaseInformation(ctx *TestContext) {
	ctx.TheJSONResponseShouldHaveStructure([]string{"database"})

	if ctx.Response == nil || len(ctx.ResponseBody) == 0 {
		return // Error already reported by TheJSONResponseShouldHaveStructure
	}

	var jsonData map[string]interface{}
	err := json.Unmarshal(ctx.ResponseBody, &jsonData)
	if err != nil {
		return // Error already reported by TheJSONResponseShouldHaveStructure
	}

	database, exists := jsonData["database"]
	if !exists {
		return // Error already reported by TheJSONResponseShouldHaveStructure
	}

	databaseInfo, ok := database.(map[string]interface{})
	if !ok {
		ctx.T.Errorf("Expected database field to be an object, but got %T", database)
		return
	}

	if _, statusExists := databaseInfo["status"]; !statusExists {
		ctx.T.Errorf("Expected database information to contain 'status' field, but it was missing")
	}

	if _, messageExists := databaseInfo["message"]; !messageExists {
		ctx.T.Errorf("Expected database information to contain 'message' field, but it was missing")
	}
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

// ThenAllDatabaseHealthRequestsShouldSucceed checks that all concurrent database health requests succeeded in BDD style
func ThenAllDatabaseHealthRequestsShouldSucceed(ctx *TestContext) {
	if len(ctx.ConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent database health results to exist, but none were found")
		return
	}

	failedCount := 0
	for i, result := range ctx.ConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected concurrent database health request %d to succeed, but got error: %v", i+1, result.Error)
			failedCount++
			continue
		}
		if result.Response == nil {
			ctx.T.Errorf("Expected concurrent database health request %d to have a response, but none was received", i+1)
			failedCount++
			continue
		}
		if result.Response.StatusCode != 200 {
			ctx.T.Errorf("Expected concurrent database health request %d to have status 200, but got %d", i+1, result.Response.StatusCode)
			failedCount++
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent database health requests to succeed, but %d failed", len(ctx.ConcurrentResults), failedCount)
	}
}

// ThenAllResponsesShouldContainValidConnectionStatistics checks all concurrent responses for valid stats in BDD style
func ThenAllResponsesShouldContainValidConnectionStatistics(ctx *TestContext) {
	if len(ctx.ConcurrentResults) == 0 {
		ctx.T.Errorf("Expected concurrent results to exist for connection statistics validation, but none were found")
		return
	}

	failedCount := 0
	for i, result := range ctx.ConcurrentResults {
		if result.Error != nil {
			ctx.T.Errorf("Expected concurrent request %d to succeed for statistics validation, but got error: %v", i+1, result.Error)
			failedCount++
			continue
		}
		if result.Response == nil {
			ctx.T.Errorf("Expected concurrent request %d to have a response for statistics validation, but none was received", i+1)
			failedCount++
			continue
		}
		if len(result.Body) == 0 {
			ctx.T.Errorf("Expected concurrent request %d to have response body for statistics validation, but it was empty", i+1)
			failedCount++
			continue
		}

		var jsonData map[string]interface{}
		err := json.Unmarshal(result.Body, &jsonData)
		if err != nil {
			ctx.T.Errorf("Expected concurrent request %d response to be valid JSON for statistics validation, but got error: %v", i+1, err)
			failedCount++
			continue
		}

		database, exists := jsonData["database"]
		if !exists {
			ctx.T.Errorf("Expected concurrent request %d response to contain 'database' field for statistics validation, but it was missing", i+1)
			failedCount++
			continue
		}

		databaseInfo, ok := database.(map[string]interface{})
		if !ok {
			ctx.T.Errorf("Expected concurrent request %d database field to be an object for statistics validation, but got %T", i+1, database)
			failedCount++
			continue
		}

		stats, statsExists := databaseInfo["stats"]
		if !statsExists {
			ctx.T.Errorf("Expected concurrent request %d database info to contain 'stats' field for statistics validation, but it was missing", i+1)
			failedCount++
			continue
		}

		statsMap, ok := stats.(map[string]interface{})
		if !ok {
			ctx.T.Errorf("Expected concurrent request %d stats field to be an object for statistics validation, but got %T", i+1, stats)
			failedCount++
			continue
		}

		// Basic validation of stats structure
		requiredFields := []string{"max_open_connections", "open_connections", "in_use", "idle"}
		for _, field := range requiredFields {
			value, exists := statsMap[field]
			if !exists {
				ctx.T.Errorf("Expected concurrent request %d stats to contain '%s' field for statistics validation, but it was missing", i+1, field)
				failedCount++
				continue
			}

			// Check if value is a number
			switch value.(type) {
			case int, int64, float64:
				// Valid number type
			default:
				ctx.T.Errorf("Expected concurrent request %d %s to be a number for statistics validation, but got %T", i+1, field, value)
				failedCount++
			}
		}
	}

	if failedCount > 0 {
		ctx.T.Errorf("Expected all %d concurrent responses to contain valid connection statistics, but %d had issues", len(ctx.ConcurrentResults), failedCount)
	}
}
