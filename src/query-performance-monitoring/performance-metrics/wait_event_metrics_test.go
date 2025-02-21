package performancemetrics

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"

	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetWaitEventMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	cp := commonparameters.SetCommonParameters(args, uint64(14), databaseName)

	var query = fmt.Sprintf(queries.WaitEvents, databaseName, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}).AddRow(
		"Locks:Lock", "Locks", 1000.0, "2023-01-01T00:00:00Z", "queryid1", "SELECT 1", "testdb",
	))
	waitEventsList, err := getWaitEventMetrics(conn, cp)

	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetWaitEventEmptyMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	args := args.ArgumentList{QueryMonitoringCountThreshold: 10}
	databaseName := "testdb"
	cp := commonparameters.SetCommonParameters(args, uint64(14), databaseName)

	var query = fmt.Sprintf(queries.WaitEvents, databaseName, args.QueryMonitoringCountThreshold)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}))
	waitEventsList, err := getWaitEventMetrics(conn, cp)
	assert.NoError(t, err)
	assert.Len(t, waitEventsList, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInEligibilityWaitEvents(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	argumentList := args.ArgumentList{QueryMonitoringCountThreshold: 10, Hostname: "localhost", Port: "5432"}
	databaseName := "testdb"
	cp := commonparameters.SetCommonParameters(argumentList, 14, databaseName)
	enabledExtensions := map[string]bool{"pg_wait_sampling": false}
	err := PopulateWaitEventMetrics(conn, pgIntegration, cp, enabledExtensions)
	assert.Equal(t, err, commonutils.ErrNotEligible)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPopulateWaitEventMetricsErr(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	argumentList := args.ArgumentList{QueryMonitoringCountThreshold: 10, Hostname: "localhost", Port: "5432"}
	databaseName := "testdb"
	cp := commonparameters.SetCommonParameters(argumentList, 14, databaseName)
	enabledExtensions := map[string]bool{"pg_wait_sampling": true, "pg_stat_statements": true}
	err := PopulateWaitEventMetrics(conn, pgIntegration, cp, enabledExtensions)
	assert.Equal(t, err, commonutils.ErrUnExpectedError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPopulateWaitEventMetrics(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	pgIntegration, _ := integration.New("test", "1.0.0")
	argumentList := args.ArgumentList{QueryMonitoringCountThreshold: 10, Hostname: "localhost", Port: "5432"}
	databaseName := "testdb"
	cp := commonparameters.SetCommonParameters(argumentList, 14, databaseName)
	enabledExtensions := map[string]bool{"pg_wait_sampling": true, "pg_stat_statements": true}
	query := fmt.Sprintf(queries.WaitEvents, databaseName, argumentList.QueryMonitoringCountThreshold)
	rowData := []driver.Value{
		"Locks:Lock", "Locks", 1000.0, "2023-01-01T00:00:00Z", "queryid1", "SELECT 1", "testdb",
	}
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{
		"wait_event_name", "wait_category", "total_wait_time_ms", "collection_timestamp", "query_id", "query_text", "database_name",
	}).AddRow(rowData...).AddRow(rowData...))
	err := PopulateWaitEventMetrics(conn, pgIntegration, cp, enabledExtensions)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
