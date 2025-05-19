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
	var slowQueryMetricsList []datamodels.SlowRunningQueryMetrics
	var slowQueryMetricsListInterface []interface{}
	versionSpecificSlowQuery, err := commonutils.FetchVersionSpecificSlowQuery(cp.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, nil, err
	}
	var query = fmt.Sprintf(versionSpecificSlowQuery, cp.Databases, cp.QueryMonitoringCountThreshold)
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var slowQuery datamodels.SlowRunningQueryMetrics
		if scanErr := rows.StructScan(&slowQuery); scanErr != nil {
			return nil, nil, err
		}
		slowQueryMetricsList = append(slowQueryMetricsList, slowQuery)
		slowQueryMetricsListInterface = append(slowQueryMetricsListInterface, slowQuery)
	}
	return slowQueryMetricsList, slowQueryMetricsListInterface, nil
}

func PopulateSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) []datamodels.SlowRunningQueryMetrics {
	isEligible := validations.CheckSlowQueryMetricsFetchEligibility(enabledExtensions)
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return nil
	}

	slowQueryMetricsList, slowQueryMetricsListInterface, err := getSlowRunningMetrics(conn, cp)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil
	}

	if len(slowQueryMetricsList) == 0 {
		log.Debug("No slow-running queries found.")
		return nil
	}
	err = commonutils.IngestMetric(slowQueryMetricsListInterface, "PostgresSlowQueries", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting slow-running queries: %v", err)
		return nil
	}
	return slowQueryMetricsList
}

func PopulateSlowRunningMetricsPgStat(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters, enabledExtensions map[string]bool) []datamodels.SlowRunningQueryMetrics {
	isEligible := validations.CheckSlowQueryMetricsFetchEligibility(enabledExtensions)
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return nil
	}
	individualQueries := getIndividualQueriesFromPgStat(conn)
	slowQueryMetricsList, _, err := getSlowRunningMetrics(conn, cp)
	filteredSlowQueryMetrics, filteredSlowQueryMetricsInterface := getFilteredSlowMetrics(individualQueries, slowQueryMetricsList)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil
	}

	if len(slowQueryMetricsList) == 0 {
		log.Debug("No slow-running queries found.")
		return nil
	}
	err = commonutils.IngestMetric(filteredSlowQueryMetricsInterface, "PostgresSlowQueries", pgIntegration, cp)
	if err != nil {
		log.Error("Error ingesting slow-running queries: %v", err)
		return nil
	}
	return filteredSlowQueryMetrics
}

func getFilteredSlowMetrics(individualQueries []string, slowQueryMetrics []datamodels.SlowRunningQueryMetrics) ([]datamodels.SlowRunningQueryMetrics, []interface{}) {
	filteredSlowQueryMetrics := make([]datamodels.SlowRunningQueryMetrics, 0)
	filteredSlowQueryMetricsInterface := make([]interface{}, 0)
	individualQueryMap := make(map[string]string)
	for _, individualQuery := range individualQueries {
		individualQueryMap[commonutils.AnonymizeAndNormalize(individualQuery)] = individualQuery
	}
	for _, slowQueryMetric := range slowQueryMetrics {
		normalizedSlowQueryText := commonutils.AnonymizeAndNormalize(*slowQueryMetric.QueryText)
		if _, exists := individualQueryMap[normalizedSlowQueryText]; exists {
			individualQuerySample := individualQueryMap[normalizedSlowQueryText]
			slowQueryMetric.IndividualQuery = &individualQuerySample
			filteredSlowQueryMetricsInterface = append(filteredSlowQueryMetricsInterface, slowQueryMetric)
			filteredSlowQueryMetrics = append(filteredSlowQueryMetrics, slowQueryMetric)
		}
	}
	return filteredSlowQueryMetrics, filteredSlowQueryMetricsInterface
}
