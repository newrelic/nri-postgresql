package commonutils

import (
	"regexp"
	"strconv"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
)

func FetchVersion(conn *performancedbconnection.PGSQLConnection) (int, error) {
	var versionStr string
	rows, err := conn.Queryx("SELECT version()")
	if err != nil {
		log.Error("Error executing query: %v", err)
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, ErrVersionFetchError
	}
	if scanErr := rows.Scan(&versionStr); scanErr != nil {
		log.Error("Error scanning version: %v", err)
		return 0, err
	}
	re := regexp.MustCompile(VersionRegex)
	matches := re.FindStringSubmatch(versionStr)
	if len(matches) < VersionIndex {
		log.Error("Unable to parse PostgreSQL version from string: %s", versionStr)
		return 0, ErrParseVersion
	}

	version, err := strconv.Atoi(matches[1])
	log.Debug("version", version)
	if err != nil {
		log.Error("Error converting version to integer: %v", err)
		return 0, err
	}
	return version, nil
}

func FetchVersionSpecificSlowQueries(conn *performancedbconnection.PGSQLConnection) (string, error) {
	version, err := FetchVersion(conn)
	if err != nil {
		return "", err
	}
	switch {
	case version == PostgresVersion12:
		return queries.SlowQueriesForV12, nil
	case version >= PostgresVersion13:
		return queries.SlowQueriesForV13AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchVersionSpecificBlockingQueries(conn *performancedbconnection.PGSQLConnection) (string, error) {
	version, err := FetchVersion(conn)
	if err != nil {
		return "", err
	}
	switch {
	case version == PostgresVersion12, version == PostgresVersion13:
		return queries.BlockingQueriesForV12AndV13, nil
	case version >= PostgresVersion14:
		return queries.BlockingQueriesForV14AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}

func FetchVersionSpecificIndividualQueries(conn *performancedbconnection.PGSQLConnection) (string, error) {
	version, err := FetchVersion(conn)
	if err != nil {
		return "", err
	}
	switch {
	case version == PostgresVersion12:
		return queries.IndividualQuerySearchV12, nil
	case version >= PostgresVersion12:
		return queries.IndividualQuerySearchV13AndAbove, nil
	default:
		return "", ErrUnsupportedVersion
	}
}
