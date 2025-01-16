package performancemetrics

import (
	"fmt"

	globalvariables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateBlockingMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, gv *globalvariables.GlobalVariables) error {
	isEligible, enableCheckError := validations.CheckBlockingSessionMetricsFetchEligibility(conn, gv.Version)
	if enableCheckError != nil {
		log.Debug("Error executing query: %v in PopulateBlockingMetrics", enableCheckError)
		return commonutils.ErrUnExpectedError
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return commonutils.ErrNotEligible
	}
	blockingQueriesMetricsList, blockQueryFetchErr := GetBlockingMetrics(conn, gv)
	if blockQueryFetchErr != nil {
		log.Error("Error fetching Blocking queries: %v", blockQueryFetchErr)
		return commonutils.ErrUnExpectedError
	}
	if len(blockingQueriesMetricsList) == 0 {
		log.Debug("No Blocking queries found.")
		return nil
	}
	commonutils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration, gv)
	return nil
}

func GetBlockingMetrics(conn *performancedbconnection.PGSQLConnection, gv *globalvariables.GlobalVariables) ([]interface{}, error) {
	var blockingQueriesMetricsList []interface{}
	versionSpecificBlockingQuery, err := commonutils.FetchVersionSpecificBlockingQueries(gv.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, err
	}
	var query = fmt.Sprintf(versionSpecificBlockingQuery, gv.DatabaseString, min(gv.QueryCountThreshold, commonutils.MaxQueryThreshold))
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Failed to execute query: %v", err)
		return nil, err
	}
	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if scanError := rows.StructScan(&blockingQueryMetric); scanError != nil {
			return nil, scanError
		}
		if gv.Version == commonutils.PostgresVersion13 || gv.Version == commonutils.PostgresVersion12 {
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
