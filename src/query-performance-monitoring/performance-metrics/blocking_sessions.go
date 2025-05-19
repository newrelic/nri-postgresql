package performancemetrics

import (
	"fmt"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"

	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateBlockingMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) {
	isEligible := validations.CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, cp.Version)
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return
	}
	blockingQueriesMetricsList, blockQueryFetchErr := getBlockingMetrics(conn, cp)
	if blockQueryFetchErr != nil {
		log.Error("Error fetching Blocking queries: %v", blockQueryFetchErr)
		return
	}
	if len(blockingQueriesMetricsList) == 0 {
		log.Debug("No Blocking queries found.")
		return
	}
	err := commonutils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting Blocking queries: %v", err)
		return
	}
}

func getBlockingMetrics(conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters) ([]interface{}, error) {
	var blockingQueriesMetricsList []interface{}
	versionSpecificBlockingQuery, err := commonutils.FetchVersionSpecificBlockingQuery(cp.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, err
	}
	var query = fmt.Sprintf(versionSpecificBlockingQuery, cp.Databases, cp.QueryMonitoringCountThreshold)
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Failed to execute query: %v", err)
		return nil, commonutils.ErrUnExpectedError
	}
	defer rows.Close()
	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if scanError := rows.StructScan(&blockingQueryMetric); scanError != nil {
			return nil, scanError
		}
		// For PostgreSQL versions 13 and 12, anonymization of queries does not occur for blocking sessions, so it's necessary to explicitly anonymize them.
		if cp.Version == commonutils.PostgresVersion13 || cp.Version == commonutils.PostgresVersion12 {
			*blockingQueryMetric.BlockedQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockedQuery)
			*blockingQueryMetric.BlockingQuery = commonutils.AnonymizeQueryText(*blockingQueryMetric.BlockingQuery)
		}
		blockingQueriesMetricsList = append(blockingQueriesMetricsList, blockingQueryMetric)
	}

	return blockingQueriesMetricsList, nil
}

func PopulateBlockingMetricsPgStat(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool, slowQueryMetrics []datamodels.SlowRunningQueryMetrics) {
	isEligible := validations.CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, cp.Version)
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return
	}
	blockingQueriesMetricsList, blockQueryFetchErr := getBlockingMetricsPgStat(conn, cp)
	if blockQueryFetchErr != nil {
		log.Error("Error fetching Blocking queries: %v", blockQueryFetchErr)
		return
	}
	if len(blockingQueriesMetricsList) == 0 {
		log.Debug("No Blocking queries found.")
		return
	}
	blockingSessionMetricInterface := getFilteredBlockingSessions(blockingQueriesMetricsList, slowQueryMetrics)
	err := commonutils.IngestMetric(blockingSessionMetricInterface, "PostgresBlockingSessions", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting Blocking queries: %v", err)
		return
	}
}

func getBlockingMetricsPgStat(conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters) ([]datamodels.BlockingSessionMetrics, error) {
	var blockingQueriesMetricsList []datamodels.BlockingSessionMetrics
	var query = fmt.Sprintf(queries.RDSPostgresBlockingQuery, cp.Databases, cp.QueryMonitoringCountThreshold)
	rows, err := conn.Queryx(query)
	if err != nil {
		log.Error("Failed to execute query: %v", err)
		return nil, commonutils.ErrUnExpectedError
	}
	defer rows.Close()
	for rows.Next() {
		var blockingQueryMetric datamodels.BlockingSessionMetrics
		if scanError := rows.StructScan(&blockingQueryMetric); scanError != nil {
			return nil, scanError
		}
		blockingQueriesMetricsList = append(blockingQueriesMetricsList, blockingQueryMetric)
	}
	return blockingQueriesMetricsList, nil
}

func getFilteredBlockingSessions(blockingSessionMetrics []datamodels.BlockingSessionMetrics, slowQueryMetrics []datamodels.SlowRunningQueryMetrics) []interface{} {
	filteredBlockingSessionMetricList := make([]interface{}, 0)
	slowQueryTextMap := make(map[string]datamodels.SlowRunningQueryMetrics)
	for _, metric := range slowQueryMetrics {
		slowQueryTextMap[commonutils.AnonymizeAndNormalize(*metric.QueryText)] = metric
	}
	for _, blockingSessionMetric := range blockingSessionMetrics {
		normalizedBlockingQuery := commonutils.AnonymizeAndNormalize(*blockingSessionMetric.BlockingQuery)
		normalizedBlockedQuery := commonutils.AnonymizeAndNormalize(*blockingSessionMetric.BlockedQuery)
		_, blockingQueryMetricExists := slowQueryTextMap[normalizedBlockingQuery]
		_, blockedQueryMetricExists := slowQueryTextMap[normalizedBlockedQuery]

		if blockingQueryMetricExists && blockedQueryMetricExists {
			blockingSessionMetric.BlockingQuery = slowQueryTextMap[normalizedBlockingQuery].QueryText
			blockingSessionMetric.BlockingQueryID = slowQueryTextMap[normalizedBlockingQuery].QueryID
			blockingSessionMetric.BlockedQuery = slowQueryTextMap[normalizedBlockedQuery].QueryText
			blockingSessionMetric.BlockedQueryID = slowQueryTextMap[normalizedBlockedQuery].QueryID
			filteredBlockingSessionMetricList = append(filteredBlockingSessionMetricList, blockingSessionMetric)
		}
	}
	return filteredBlockingSessionMetricList
}
