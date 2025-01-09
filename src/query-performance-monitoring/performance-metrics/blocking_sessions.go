package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateBlockingMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, args args.ArgumentList, databaseName string) {
	isPgStatStatementEnabled, enableCheckError := validations.CheckBlockingSessionMetricsFetchEligibility(conn)
	if enableCheckError != nil {
		log.Debug("Error executing query: %v in PopulateBlockingMetrics", enableCheckError)
		return
	}
	if !isPgStatStatementEnabled {
		log.Debug("Extension 'pg_stat_statements' is not enabled for the database.")
		return
	}
	blockingQueriesMetricsList, blockQueryFetchErr := GetBlockingMetrics(conn, args, databaseName)
	if blockQueryFetchErr != nil {
		log.Error("Error fetching Blocking queries: %v", blockQueryFetchErr)
		return
	}
	if len(blockingQueriesMetricsList) == 0 {
		log.Debug("No Blocking queries found.")
		return
	}
	commonutils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration, args)
}

func GetBlockingMetrics(conn *performancedbconnection.PGSQLConnection, args args.ArgumentList, databaseName string) ([]interface{}, error) {
	var blockingQueriesMetricsList []interface{}
	versionSpecificBlockingQuery, err := commonutils.FetchVersionSpecificBlockingQueries(conn)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, err
	}
	var query = fmt.Sprintf(versionSpecificBlockingQuery, databaseName, min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Failed to execute query: %v", err)
		return nil, err
	}
	version, versionErr := commonutils.FetchVersion(conn)
	if versionErr != nil {
		log.Error("Failed to fetch version: %v", versionErr)
		return nil, versionErr
	}
	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if scanError := rows.StructScan(&blockingQueryMetric); scanError != nil {
			return nil, scanError
		}
		if version == commonutils.PostgresVersion13 || version == commonutils.PostgresVersion12 {
			*blockingQueryMetric.BlockedQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockedQuery)
			*blockingQueryMetric.BlockingQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockingQuery)
		}
		blockingQueriesMetricsList = append(blockingQueriesMetricsList, blockingQueryMetric)
	}

	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Error closing rows: %v", closeErr)
		return nil, closeErr
	}
	return blockingQueriesMetricsList, nil
}
