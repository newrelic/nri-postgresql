package commonutils

import (
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
)

func FetchVersionSpecificSlowQueries(version uint64) (string, error) {
	switch {
	case version == PostgresVersion12:
		return queries.SlowQueriesForV12, nil
	case version >= PostgresVersion13:
		return queries.SlowQueriesForV13AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchVersionSpecificBlockingQueries(version uint64) (string, error) {
	switch {
	case version == PostgresVersion12, version == PostgresVersion13:
		return queries.BlockingQueriesForV12AndV13, nil
	case version >= PostgresVersion14:
		return queries.BlockingQueriesForV14AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchVersionSpecificIndividualQueries(version uint64) (string, error) {
	switch {
	case version == PostgresVersion12:
		return queries.IndividualQuerySearchV12, nil
	case version >= PostgresVersion12:
		return queries.IndividualQuerySearchV13AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}
