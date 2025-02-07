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

func GetSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection, cp *commonparameters.CommonParameters) ([]datamodels.SlowRunningQueryMetrics, []interface{}, error) {
	var slowQueryMetricsList []datamodels.SlowRunningQueryMetrics
	var slowQueryMetricsListInterface []interface{}
	versionSpecificSlowQuery, err := commonutils.FetchVersionSpecificSlowQueries(cp.Version)
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
	isEligible, err := validations.CheckSlowQueryMetricsFetchEligibility(enabledExtensions)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return nil
	}

	slowQueryMetricsList, slowQueryMetricsListInterface, err := GetSlowRunningMetrics(conn, cp)
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
