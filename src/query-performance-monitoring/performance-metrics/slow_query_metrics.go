package performancemetrics

import (
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection, args args.ArgumentList, databaseNames string, version uint64) ([]datamodels.SlowRunningQueryMetrics, []interface{}, error) {
	var slowQueryMetricsList []datamodels.SlowRunningQueryMetrics
	var slowQueryMetricsListInterface []interface{}
	versionSpecificSlowQuery, err := commonutils.FetchVersionSpecificSlowQueries(version)
	if err != nil {
		log.Error("Unsupported postgres version: %v", err)
		return nil, nil, err
	}
	var query = fmt.Sprintf(versionSpecificSlowQuery, databaseNames, min(args.QueryCountThreshold, commonutils.MaxQueryThreshold))
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var slowQuery datamodels.SlowRunningQueryMetrics
		if scanErr := rows.StructScan(&slowQuery); scanErr != nil {
			return nil, nil, err
		}
		slowQueryMetricsList = append(slowQueryMetricsList, slowQuery)
		slowQueryMetricsListInterface = append(slowQueryMetricsListInterface, slowQuery)
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Error("Error closing rows: %v", closeErr)
		return nil, nil, closeErr
	}
	return slowQueryMetricsList, slowQueryMetricsListInterface, nil
}

func PopulateSlowRunningMetrics(conn *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, args args.ArgumentList, databaseNames string, version uint64) []datamodels.SlowRunningQueryMetrics {
	isExtensionEnabled, err := validations.CheckSlowQueryMetricsFetchEligibility(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isExtensionEnabled {
		log.Debug("Extension 'pg_stat_statements' is not enabled.")
		return nil
	}

	slowQueryMetricsList, slowQueryMetricsListInterface, err := GetSlowRunningMetrics(conn, args, databaseNames, version)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil
	}

	if len(slowQueryMetricsList) == 0 {
		log.Debug("No slow-running queries found.")
		return nil
	}
	commonutils.IngestMetric(slowQueryMetricsListInterface, "PostgresSlowQueries", pgIntegration, args)
	return slowQueryMetricsList
}
