package performancemetrics

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	common_parameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetIndividualQueryMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	mockQueryID := "-123"
	mockQueryText := "SELECT 1"
	cp := common_parameters.SetCommonParameters(args, version, databaseName)

	// Mock the individual query
	query := fmt.Sprintf(queries.IndividualQuerySearchV13AndAbove, mockQueryID, databaseName, args.QueryMonitoringResponseTimeThreshold, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "query", "queryid", "datname", "planid", "cpu_time_ms", "exec_time_ms",
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

	individualQueryMetricsInterface, individualQueryMetrics := getIndividualQueryMetrics(conn, slowRunningQueries, cp)

	assert.Len(t, individualQueryMetricsInterface, 1)
	assert.Len(t, individualQueryMetrics, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
