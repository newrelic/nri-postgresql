package performancemetrics

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func getSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters) ([]datamodels.SlowRunningQueryMetrics, []interface{}, error) {
	versionSpecificSlowQuery, err := commonutils.FetchVersionSpecificSlowQueries(cp.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, nil, err
	}
	query := fmt.Sprintf(versionSpecificSlowQuery, cp.Databases, cp.QueryMonitoringCountThreshold)
	slowQueryMetricsList, slowQueryMetricsListInterface, err := fetchMetrics[datamodels.SlowRunningQueryMetrics](conn, query, "Slow Running")
	return slowQueryMetricsList, slowQueryMetricsListInterface, err
}

func PopulateSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) []datamodels.SlowRunningQueryMetrics {
	isEligible, err := validations.CheckSlowQueryMetricsFetchEligibility(enabledExtensions)
	if err != nil {
		log.Error("Error executing query for eligibility check: %v", err)
		return []datamodels.SlowRunningQueryMetrics{}
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return []datamodels.SlowRunningQueryMetrics{}
	}

	slowQueryMetricsList, slowQueryMetricsListInterface, err := getSlowRunningMetrics(conn, cp)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return []datamodels.SlowRunningQueryMetrics{}
	}

	if len(slowQueryMetricsList) == 0 {
		log.Debug("No slow-running queries found.")
		return []datamodels.SlowRunningQueryMetrics{}
	}
	err = commonutils.IngestMetric(slowQueryMetricsListInterface, "PostgresSlowQueries", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting slow-running queries: %v", err)
		return []datamodels.SlowRunningQueryMetrics{}
	}
	log.Debug("Successfully ingested slow running metrics for databases")
	return slowQueryMetricsList
}
