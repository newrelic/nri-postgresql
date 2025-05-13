package commonutils_test

import (
	"testing"

	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"

	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/stretchr/testify/assert"
)

func runTestCases(t *testing.T, tests []struct {
	version   uint64
	expected  string
	expectErr bool
}, fetchFunc func(uint64) (string, error)) {
	for _, test := range tests {
		result, err := fetchFunc(test.version)
		if test.expectErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, result)
		}
	}
}

func TestFetchVersionSpecificSlowQueries(t *testing.T) {
	tests := []struct {
		version   uint64
		expected  string
		expectErr bool
	}{
		{commonutils.PostgresVersion12, queries.SlowQueriesForV12, false},
		{commonutils.PostgresVersion13, queries.SlowQueriesForV13AndAbove, false},
		{commonutils.PostgresVersion11, "", true},
	}

	runTestCases(t, tests, commonutils.FetchVersionSpecificSlowQuery)
}

func TestFetchVersionSpecificBlockingQueries(t *testing.T) {
	tests := []struct {
		version   uint64
		expected  string
		expectErr bool
	}{
		{commonutils.PostgresVersion12, queries.BlockingQueriesForV12AndV13, false},
		{commonutils.PostgresVersion13, queries.BlockingQueriesForV12AndV13, false},
		{commonutils.PostgresVersion14, queries.BlockingQueriesForV14AndAbove, false},
		{commonutils.PostgresVersion11, "", true},
	}

	runTestCases(t, tests, commonutils.FetchVersionSpecificBlockingQuery)
}

func TestFetchVersionSpecificIndividualQueries(t *testing.T) {
	tests := []struct {
		version   uint64
		expected  string
		expectErr bool
	}{
		{commonutils.PostgresVersion12, queries.IndividualQuerySearchV12, false},
		{commonutils.PostgresVersion13, queries.IndividualQuerySearchV13AndAbove, false},
		{commonutils.PostgresVersion14, queries.IndividualQuerySearchV13AndAbove, false},
		{commonutils.PostgresVersion11, "", true},
	}

	runTestCases(t, tests, commonutils.FetchVersionSpecificIndividualQueries)
}
