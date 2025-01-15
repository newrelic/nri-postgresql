package performancemetrics_test

import (
	"fmt"
	"regexp"
	"testing"

	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
	"github.com/stretchr/testify/assert"
)

func TestPopulateBlockingMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	expectedQuery := queries.BlockingQueriesForV14AndAbove
	query := fmt.Sprintf(expectedQuery, databaseName, min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "blocked_pid", "blocked_query", "blocked_query_id", "blocked_query_start", "database_name",
		"blocking_pid", "blocking_query", "blocking_query_id", "blocking_query_start",
	}).AddRow(
		"newrelic_value", 123, "SELECT 1", 1233444, "2023-01-01 00:00:00", "testdb",
		456, "SELECT 2", 4566, "2023-01-01 00:00:00",
	))

	err := performancemetrics.PopulateBlockingMetrics(conn, pgIntegration, args, databaseName, version)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPopulateBlockingMetricsSupportedVersionExtensionNotRequired(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(12)
	expectedQuery := queries.BlockingQueriesForV12AndV13
	query := fmt.Sprintf(expectedQuery, databaseName, min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "blocked_pid", "blocked_query", "blocked_query_id", "blocked_query_start", "database_name",
		"blocking_pid", "blocking_query", "blocking_query_id", "blocking_query_start",
	}).AddRow(
		"newrelic_value", 123, "SELECT 1", 1233444, "2023-01-01 00:00:00", "testdb",
		456, "SELECT 2", 4566, "2023-01-01 00:00:00",
	))
	err := performancemetrics.PopulateBlockingMetrics(conn, pgIntegration, args, databaseName, version)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPopulateBlockingMetricsUnSupportedVersion(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(11)
	err := performancemetrics.PopulateBlockingMetrics(conn, pgIntegration, args, databaseName, version)
	assert.EqualError(t, err, commonutils.ErrNotEligible.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPopulateBlockingMetricsExtensionsNotEnabled(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(14)
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	err := performancemetrics.PopulateBlockingMetrics(conn, pgIntegration, args, databaseName, version)
	assert.EqualError(t, err, commonutils.ErrNotEligible.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlockingMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	expectedQuery := queries.BlockingQueriesForV12AndV13
	query := fmt.Sprintf(expectedQuery, databaseName, min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"newrelic", "blocked_pid", "blocked_query", "blocked_query_id", "blocked_query_start", "database_name",
		"blocking_pid", "blocking_query", "blocking_query_id", "blocking_query_start",
	}).AddRow(
		"newrelic_value", 123, "SELECT 1", 1233444, "2023-01-01 00:00:00", "testdb",
		456, "SELECT 2", 4566, "2023-01-01 00:00:00",
	))
	blockingQueriesMetricsList, err := performancemetrics.GetBlockingMetrics(conn, args, databaseName, version)

	assert.NoError(t, err)
	assert.Len(t, blockingQueriesMetricsList, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
