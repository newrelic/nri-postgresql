package performancemetrics

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"

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

func TestPopulateIndividualQueryMetricsPgStat(t *testing.T) {
	slowQueries := []datamodels.SlowRunningQueryMetrics{
		{
			QueryID:         stringPtr("query1"),
			DatabaseName:    stringPtr("testdb"),
			QueryText:       stringPtr("SELECT * FROM test where id = $1"),
			IndividualQuery: stringPtr("SELECT * FROM test where id = 1"),
		},
		{
			QueryID:         stringPtr("query2"),
			DatabaseName:    stringPtr("testdb"),
			QueryText:       stringPtr("SELECT * FROM users where id = $1"),
			IndividualQuery: stringPtr("SELECT * FROM users where id = 1"),
		},
	}
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	version := uint64(13)
	databaseName := "testdb"
	cp := common_parameters.SetCommonParameters(args, version, databaseName)
	result := PopulateIndividualQueryMetricsPgStat(slowQueries, pgIntegration, cp)
	assert.NotEmpty(t, pgIntegration.Entities)
	assert.Len(t, result, 2)
	assert.Equal(t, "query1", *result[0].QueryID)
	assert.Equal(t, "testdb", *result[0].DatabaseName)
	assert.Equal(t, "SELECT * FROM test where id = $1", *result[0].QueryText)
	assert.Equal(t, "SELECT * FROM test where id = 1", *result[0].RealQueryText)
}

func TestPopulateIndividualQueryMetricsPgStatEmpty(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	version := uint64(13)
	databaseName := "testdb"
	cp := common_parameters.SetCommonParameters(args, version, databaseName)
	result := PopulateIndividualQueryMetricsPgStat(nil, pgIntegration, cp)
	assert.NotEmpty(t, pgIntegration.Entities)
	assert.Len(t, result, 0)
}

func stringPtr(s string) *string {
	return &s
}
