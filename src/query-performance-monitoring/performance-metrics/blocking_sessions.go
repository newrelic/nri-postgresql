package performancemetrics

import (
	"fmt"

	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
)

func PopulateBlockingMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) {
	isEligible, enableCheckError := validations.CheckBlockingSessionMetricsFetchEligibility(enabledExtensions, cp.Version)
	if enableCheckError != nil {
		log.Error("Error executing query for eligibility check in PopulateBlockingMetrics: %v", enableCheckError)
		return
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return
	}
	blockingQueriesMetricsList, blockQueryMetricsFetchErr := getBlockingMetrics(conn, cp)
	if blockQueryMetricsFetchErr != nil {
		log.Error("Error fetching blocking queries: %v", blockQueryMetricsFetchErr)
		return
	}
	if len(blockingQueriesMetricsList) == 0 {
		log.Debug("No Blocking queries found.")
		return
	}
	err := commonutils.IngestMetric(blockingQueriesMetricsList, "PostgresBlockingSessions", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting blocking queries: %v", err)
		return
	}
	log.Debug("Successfully ingested blocking metrics ")
}

func getBlockingMetrics(conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters) ([]interface{}, error) {
	versionSpecificBlockingQuery, err := commonutils.FetchVersionSpecificBlockingQueries(cp.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, err
	}
	query := fmt.Sprintf(versionSpecificBlockingQuery, cp.Databases, cp.QueryMonitoringCountThreshold)
	blockingQueriesMetricsList, _, err := fetchMetrics[datamodels.BlockingSessionMetrics](conn, query, "Blocking Query")
	if err != nil {
		log.Error("Error fetching blocking queries: %v", err)
		return nil, commonutils.ErrUnExpectedError
	}
	if cp.Version == commonutils.PostgresVersion13 || cp.Version == commonutils.PostgresVersion12 {
		for i := range blockingQueriesMetricsList {
			*blockingQueriesMetricsList[i].BlockedQuery = commonutils.AnonymizeQueryText(*blockingQueriesMetricsList[i].BlockedQuery)
			*blockingQueriesMetricsList[i].BlockingQuery = commonutils.AnonymizeQueryText(*blockingQueriesMetricsList[i].BlockingQuery)
		}
	}
	var blockingQueriesMetricsListInterface = make([]interface{}, 0)
	for _, metric := range blockingQueriesMetricsList {
		blockingQueriesMetricsListInterface = append(blockingQueriesMetricsListInterface, metric)
	}
	return blockingQueriesMetricsListInterface, nil
}
