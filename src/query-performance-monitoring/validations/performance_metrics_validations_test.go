package validations

import (
	"regexp"
	"testing"

	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestCheckBlockingSessionMetricsFetchEligibilityExtensionNotRequired(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(12)
	enabledExtensions, _ := FetchAllExtensions(conn)
	isExtensionEnabledTest := CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_statements"))
	enabledExtensions, _ := FetchAllExtensions(conn)
	isExtensionEnabledTest := CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckBlockingSessionMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	version := uint64(14)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_statements"))
	enabledExtensions, _ := FetchAllExtensions(conn)
	isExtensionEnabledTest := CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, version)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_monitor"))
	enabledExtensions, _ := FetchAllExtensions(conn)
	isExtensionEnabledTest := CheckIndividualQueryMetricsFetchEligibility(enabledExtensions)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIndividualQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}))
	enabledExtensions, _ := FetchAllExtensions(conn)
	isExtensionEnabledTest := CheckIndividualQueryMetricsFetchEligibility(enabledExtensions)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckWaitEventMetricsFetchEligibility(t *testing.T) {
	testCases := []struct {
		name              string
		enabledExtensions map[string]bool
		expected          bool
	}{
		{
			name: "Both pg_wait_sampling and pg_stat_statements enabled",
			enabledExtensions: map[string]bool{
				"pg_wait_sampling":   true,
				"pg_stat_statements": true,
			},
			expected: true,
		},
		{
			name: "Only pg_stat_statements enabled",
			enabledExtensions: map[string]bool{
				"pg_wait_sampling":   false,
				"pg_stat_statements": true,
			},
			expected: true,
		},
		{
			name: "Neither pg_wait_sampling nor pg_stat_statements enabled",
			enabledExtensions: map[string]bool{
				"pg_wait_sampling":   false,
				"pg_stat_statements": false,
			},
			expected: false,
		},
		{
			name: "Only pg_wait_sampling enabled",
			enabledExtensions: map[string]bool{
				"pg_wait_sampling":   true,
				"pg_stat_statements": false,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CheckWaitEventMetricsFetchEligibility(tc.enabledExtensions)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionSuccess(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}).AddRow("pg_stat_statements"))
	enabledExtensions, _ := FetchAllExtensions(conn)
	isExtensionEnabledTest := CheckSlowQueryMetricsFetchEligibility(enabledExtensions)
	assert.Equal(t, isExtensionEnabledTest, true)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckSlowQueryMetricsFetchEligibilitySupportedVersionFail(t *testing.T) {
	conn, mock := connection.CreateMockSQL(t)
	validationQueryStatStatements := "SELECT extname FROM pg_extension"
	mock.ExpectQuery(regexp.QuoteMeta(validationQueryStatStatements)).WillReturnRows(sqlmock.NewRows([]string{"extname"}))
	enabledExtensions, _ := FetchAllExtensions(conn)
	isExtensionEnabledTest := CheckSlowQueryMetricsFetchEligibility(enabledExtensions)
	assert.Equal(t, isExtensionEnabledTest, false)
	assert.NoError(t, mock.ExpectationsWereMet())
}
