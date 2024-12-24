package performance_metrics

import (
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPopulateExecutionPlanMetrics(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}
	results := []datamodels.IndividualQueryMetrics{
		{
			QueryId:       int64Ptr(1),
			DatabaseName:  strPtr("testdb"),
			QueryText:     strPtr("SELECT * FROM test_table"),
			RealQueryText: strPtr("SELECT * FROM test_table"),
		},
	}

	PopulateExecutionPlanMetrics(results, pgIntegration, args)

	// Add assertions to verify the expected behavior
	assert.NotNil(t, pgIntegration)
}

func TestGetExecutionPlanMetrics(t *testing.T) {
	args := args.ArgumentList{}
	results := []datamodels.IndividualQueryMetrics{
		{
			QueryId:       int64Ptr(1),
			DatabaseName:  strPtr("testdb"),
			QueryText:     strPtr("SELECT * FROM test_table"),
			RealQueryText: strPtr("SELECT * FROM test_table"),
		},
	}

	executionPlanMetrics := GetExecutionPlanMetrics(results, args)

	// Add assertions to verify the expected behavior
	assert.NotNil(t, executionPlanMetrics)
	assert.Greater(t, len(executionPlanMetrics), 0)
}

// Helper functions to create pointers for test data
func int64Ptr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}
