package performancemetrics_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
	"github.com/stretchr/testify/assert"
)

func TestPopulateIndividualQueryMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	// Mock the extension check
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM pg_extension WHERE extname = 'pg_stat_monitor'")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mockQueryID := "-123"
	mockQueryText := "SELECT 1"
	// Mock the individual query
	query := fmt.Sprintf(queries.IndividualQuerySearchV13AndAbove, mockQueryID, databaseName, args.QueryResponseTimeThreshold, min(args.QueryCountThreshold, commonutils.MaxIndividualQueryThreshold))
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

	individualQueryMetrics := performancemetrics.PopulateIndividualQueryMetrics(conn, slowRunningQueries, pgIntegration, args, databaseName, version)

	assert.Len(t, individualQueryMetrics, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConstructIndividualQuery(t *testing.T) {
	mockQueryID := "-123"
	slowRunningQuery := datamodels.SlowRunningQueryMetrics{
		QueryID: &mockQueryID,
	}
	args := args.ArgumentList{QueryResponseTimeThreshold: 100}
	databaseName := "testdb"
	versionSpecificQuery := queries.IndividualQuerySearchV13AndAbove

	actual := performancemetrics.ConstructIndividualQuery(slowRunningQuery, args, databaseName, versionSpecificQuery)
	expected := fmt.Sprintf(versionSpecificQuery, *slowRunningQuery.QueryID, databaseName, args.QueryResponseTimeThreshold, min(args.QueryCountThreshold, commonutils.MaxIndividualQueryThreshold))

	assert.Equal(t, actual, expected)
}

func TestGetIndividualQueryMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	mockQueryID := "-123"
	mockQueryText := "SELECT 1"

	// Mock the individual query
	query := fmt.Sprintf(queries.IndividualQuerySearchV13AndAbove, mockQueryID, databaseName, args.QueryResponseTimeThreshold, min(args.QueryCountThreshold, commonutils.MaxIndividualQueryThreshold))
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

	individualQueryMetricsInterface, individualQueryMetrics := performancemetrics.GetIndividualQueryMetrics(conn, slowRunningQueries, args, databaseName, version)

	assert.Len(t, individualQueryMetricsInterface, 1)
	assert.Len(t, individualQueryMetrics, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
