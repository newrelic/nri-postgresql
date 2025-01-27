package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	globalvariables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection, gv *globalvariables.GlobalVariables) ([]datamodels.SlowRunningQueryMetrics, []interface{}, error) {
	var slowQueryMetricsList []datamodels.SlowRunningQueryMetrics
	var slowQueryMetricsListInterface []interface{}
	versionSpecificSlowQuery, err := commonutils.FetchVersionSpecificSlowQueries(gv.Version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, nil, err
	}
	var query = fmt.Sprintf(versionSpecificSlowQuery, gv.DatabaseString, min(gv.QueryCountThreshold, commonutils.MaxQueryThreshold))
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

func PopulateSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, gv *globalvariables.GlobalVariables) []datamodels.SlowRunningQueryMetrics {
	isEligible, err := validations.CheckSlowQueryMetricsFetchEligibility(conn, gv.Version)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isEligible {
		log.Debug("Extension 'pg_stat_statements' is not enabled or unsupported version.")
		return nil
	}

	slowQueryMetricsList, slowQueryMetricsListInterface, err := GetSlowRunningMetrics(conn, gv)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil
	}

	if len(slowQueryMetricsList) == 0 {
		log.Debug("No slow-running queries found.")
		return nil
	}
	commonutils.IngestMetric(slowQueryMetricsListInterface, "PostgresSlowQueries", pgIntegration, gv)
	return slowQueryMetricsList
}
