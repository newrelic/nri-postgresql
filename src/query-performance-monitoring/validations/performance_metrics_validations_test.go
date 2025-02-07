package validations_test

import (
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCheckBlockingSessionMetricsFetchEligibilityExtensionNotRequired(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(12)
	enabledExtensions, _ := validations.FetchAllExtensions(conn)
	isExtensionEnabledTest, _ := validations.CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_statements"))
	enabledExtensions, _ := validations.FetchAllExtensions(conn)
	isExtensionEnabledTest, _ := validations.CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_statements"))
	enabledExtensions, _ := validations.FetchAllExtensions(conn)
	isExtensionEnabledTest, _ := validations.CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_monitor"))
	enabledExtensions, _ := validations.FetchAllExtensions(conn)
	isExtensionEnabledTest, _ := validations.CheckIndividualQueryMetricsFetchEligibility(enabledExtensions)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}))
	enabledExtensions, _ := validations.FetchAllExtensions(conn)
	isExtensionEnabledTest, _ := validations.CheckIndividualQueryMetricsFetchEligibility(enabledExtensions)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckWaitEventMetricsFetchEligibility(t *testing.T) {
	validationQuery := "SELECT extname FROM pg_extension"
	testCases := []struct {
		waitExt  string
		statExt  string
		expected bool
	}{
		{"pg_wait_sampling", "pg_stat_statements", true}, // Success
		{"pg_wait_sampling", "", false},                  // Fail V1
		{"", "pg_stat_statements", false},                // Fail V2
	}

	conn, mock := connection.CreateMockSQL(t)
	for _, tc := range testCases {
		mock.ExpectQuery(regexp.QuoteMeta(validationQuery)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow(tc.waitExt).AddRow(tc.statExt))
		enabledExtensions, _ := validations.FetchAllExtensions(conn)
		isExtensionEnabledTest, _ := validations.CheckWaitEventMetricsFetchEligibility(enabledExtensions)
		assert.Equal(t, isExtensionEnabledTest, tc.expected)
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_statements"))
	enabledExtensions, _ := validations.FetchAllExtensions(conn)
	isExtensionEnabledTest, _ := validations.CheckSlowQueryMetricsFetchEligibility(enabledExtensions)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}))
	enabledExtensions, _ := validations.FetchAllExtensions(conn)
	isExtensionEnabledTest, _ := validations.CheckSlowQueryMetricsFetchEligibility(enabledExtensions)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}
