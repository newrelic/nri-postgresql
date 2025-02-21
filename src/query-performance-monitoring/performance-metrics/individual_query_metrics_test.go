package performancemetrics

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
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
	cp := commonparameters.SetCommonParameters(args, version, databaseName)

	// Mock the individual query
	query := fmt.Sprintf(queries.IndividualQuerySearchV13AndAbove, mockQueryID, databaseName, args.QueryMonitoringResponseTimeThreshold, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"query", "queryid", "datname", "planid", "cpu_time_ms", "exec_time_ms",
	}).AddRow(
		"SELECT 1", "queryid1", "testdb", "planid1", 10.0, 20.0,
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

func TestIndividualQueriesInEligibility(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	argumentList := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	cp := commonparameters.SetCommonParameters(argumentList, version, databaseName)
	queryText := "SELECT 1"
	queryID := "123"
	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{
		{
			QueryID:      &queryID,
			QueryText:    &queryText,
			DatabaseName: &databaseName,
		},
	}
	enabledExtensions := map[string]bool{"pg_stat_monitor": false}
	individualQueriesList := PopulateIndividualQueryMetrics(conn, slowRunningQueries, nil, cp, enabledExtensions)
	assert.Len(t, individualQueriesList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPopulateIndividualQueryMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	enabledExtensions := map[string]bool{"pg_stat_monitor": true}
	databaseName := "testdb"
	version := uint64(13)
	queryText := "SELECT 1"
	queryID := "123"
	expectedQuery := queries.IndividualQuerySearchV13AndAbove
	query := fmt.Sprintf(expectedQuery, "123", databaseName, args.QueryMonitoringResponseTimeThreshold, args.QueryMonitoringCountThreshold)
	rowData := []driver.Value{
		"SELECT 1", "queryid1", "testdb", 10.0, 20.0,
	}
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"query", "queryid", "datname", "cpu_time_ms", "exec_time_ms",
	}).AddRow(rowData...).AddRow(rowData...))
	slowRunningQueries := []datamodels.SlowRunningQueryMetrics{
		{
			QueryID:      &queryID,
			QueryText:    &queryText,
			DatabaseName: &databaseName,
		},
	}
	cp := commonparameters.SetCommonParameters(args, version, databaseName)
	individualQueriesList := PopulateIndividualQueryMetrics(conn, slowRunningQueries, pgIntegration, cp, enabledExtensions)
	expectedRows := [][]driver.Value{
		rowData, rowData,
	}
	compareMockRowsWithIndividualQueryMetrics(t, expectedRows, individualQueriesList)
	assert.Len(t, individualQueriesList, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func compareMockRowsWithIndividualQueryMetrics(t *testing.T, expectedRows [][]driver.Value, individualQueryMetricsList []datamodels.IndividualQueryInfo) {
	for index := range individualQueryMetricsList {
		metric := individualQueryMetricsList[index]
		assert.Equal(t, expectedRows[index][0], *metric.RealQueryText)
		assert.Equal(t, expectedRows[index][1], *metric.QueryID)
		assert.Equal(t, expectedRows[index][2], *metric.DatabaseName)
	}
}
