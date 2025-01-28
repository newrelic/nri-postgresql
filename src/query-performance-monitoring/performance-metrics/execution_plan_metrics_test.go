package performancemetrics_test

import (
	"testing"

	global_variables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
	"github.com/stretchr/testify/assert"
)

func TestPopulateExecutionPlanMetrics(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{}
	results := []datamodels.IndividualQueryMetrics{}
	gv := global_variables.SetGlobalVariables(args, uint64(13), "testdb")

	performancemetrics.PopulateExecutionPlanMetrics(results, pgIntegration, gv)
	assert.Empty(t, pgIntegration.Entities)
}

func TestGroupQueriesByDatabase(t *testing.T) {
	databaseName := "testdb"
	queryID := "queryid1"
	queryText := "SELECT 1"
	results := []datamodels.IndividualQueryMetrics{
		{
			QueryID:      &queryID,
			QueryText:    &queryText,
			DatabaseName: &databaseName,
		},
	}

	groupedQueries := performancemetrics.GroupQueriesByDatabase(results)
	assert.Len(t, groupedQueries, 1)
	assert.Contains(t, groupedQueries, databaseName)
	assert.Len(t, groupedQueries[databaseName], 1)
}

func TestFetchNestedExecutionPlanDetails(t *testing.T) {
	queryID := "queryid1"
	queryText := "SELECT 1"
	databaseName := "testdb"
	planID := "planid1"
	individualQuery := datamodels.IndividualQueryMetrics{
		QueryID:      &queryID,
		QueryText:    &queryText,
		DatabaseName: &databaseName,
		PlanID:       &planID,
	}
	execPlan := map[string]interface{}{
		"Node Type":     "Seq Scan",
		"Relation Name": "test_table",
		"Alias":         "test_table",
		"Startup Cost":  0.00,
		"Total Cost":    1000.00,
		"Plan Rows":     100000,
		"Plan Width":    4,
	}
	execPlanLevel2 := map[string]interface{}{
		"Node Type":     "Seq Scan",
		"Relation Name": "test_table",
		"Alias":         "test_table",
		"Startup Cost":  0.00,
		"Total Cost":    1000.00,
		"Plan Rows":     100000,
		"Plan Width":    4,
		"Plans":         []interface{}{execPlan},
	}
	execPlanLevel3 := map[string]interface{}{
		"Node Type":     "Seq Scan",
		"Relation Name": "test_table",
		"Alias":         "test_table",
		"Startup Cost":  0.00,
		"Total Cost":    1000.00,
		"Plan Rows":     100000,
		"Plan Width":    4,
		"Plans":         []interface{}{execPlanLevel2},
	}
	var executionPlanMetricsList []interface{}
	level := 0

	performancemetrics.FetchNestedExecutionPlanDetails(individualQuery, &level, execPlanLevel3, &executionPlanMetricsList)
	assert.Len(t, executionPlanMetricsList, 3)
}
