package performancemetrics_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	global_variables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestPopulateWaitEventMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	gv := global_variables.SetGlobalVariables(args, version, databaseName)

	validationQuery := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_wait_sampling")
	mock.ExpectQuery(regexp.QuoteMeta(validationQuery)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	var query = fmt.Sprintf(queries.WaitEvents, databaseName, min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}).AddRow(
		"Locks:Lock", "Locks", 1000.0, "2023-01-01T00:00:00Z", "queryid1", "SELECT 1", "testdb",
	))

	err := performancemetrics.PopulateWaitEventMetrics(conn, pgIntegration, gv)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPopulateWaitEventMetricsInEligibility(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(11)
	gv := global_variables.SetGlobalVariables(args, version, databaseName)

	err := performancemetrics.PopulateWaitEventMetrics(conn, pgIntegration, gv)
	assert.EqualError(t, err, commonutils.ErrNotEligible.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPopulateWaitEventMetricsExtensionsNotEnable(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	version := uint64(13)
	gv := global_variables.SetGlobalVariables(args, version, databaseName)

	validationQuery := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_wait_sampling")
	mock.ExpectQuery(regexp.QuoteMeta(validationQuery)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	validationQueryStatStatements := fmt.Sprintf("SELECT count(*) FROM pg_extension WHERE extname = '%s'", "pg_stat_statements")
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	err := performancemetrics.PopulateWaitEventMetrics(conn, pgIntegration, gv)
	assert.EqualError(t, err, commonutils.ErrNotEligible.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWaitEventMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	gv := global_variables.SetGlobalVariables(args, uint64(14), databaseName)

	var query = fmt.Sprintf(queries.WaitEvents, databaseName, min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}).AddRow(
		"Locks:Lock", "Locks", 1000.0, "2023-01-01T00:00:00Z", "queryid1", "SELECT 1", "testdb",
	))
	waitEventsList, err := performancemetrics.GetWaitEventMetrics(conn, gv)

	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWaitEventEmptyMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryCountThreshold: 10}
	databaseName := "testdb"
	gv := global_variables.SetGlobalVariables(args, uint64(14), databaseName)

	var query = fmt.Sprintf(queries.WaitEvents, databaseName, min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}))
	waitEventsList, err := performancemetrics.GetWaitEventMetrics(conn, gv)
	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}