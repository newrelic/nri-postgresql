package performancemetrics_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	global_variables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetIndividualQueryMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	mockQueryID := "-123"
	mockQueryText := "SELECT 1"
	gv := global_variables.SetGlobalVariables(args, version, databaseName)

	// Mock the individual query
	query := fmt.Sprintf(queries.IndividualQuerySearchV13AndAbove, mockQueryID, databaseName, args.QueryResponseTimeThreshold, min(args.QueryCountThreshold, commonutils.MaxIndividualQueryCountThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "query", "queryid", "datname", "planid", "avg_cpu_time_ms", "avg_exec_time_ms",
	}).AddRow(
		"newrelic_value", "SELECT 1", "queryid1", "testdb", "planid1", 10.0, 20.0,
	))

	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{
		{
			QueryID:      &mockQueryID,
			QueryText:    &mockQueryText,
			DatabaseName: &databaseName,
		},
	}

	individualQueryMetricsInterface, individualQueryMetrics := performancemetrics.GetIndividualQueryMetrics(conn, slowRunningQueries, gv)

	assert.Len(t, individualQueryMetricsInterface, 1)
	assert.Len(t, individualQueryMetrics, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
