package performance_metrics

import (
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"reflect"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func GetSlowRunningMetrics(conn *performanceDbConnection.PGSQLConnection, queryCpuMetricsMap map[string]map[int64]float64) ([]datamodels.SlowRunningQuery, error) {
	var slowQueries []datamodels.SlowRunningQuery
	var query = queries.SlowQueries
	rows, err := conn.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var slowQuery datamodels.SlowRunningQuery
		if err := rows.StructScan(&slowQuery); err != nil {
			return nil, err
		}
		slowQueries = append(slowQueries, slowQuery)
	}

	for _, query := range slowQueries {
		log.Info("Slow Query: %+v", query)
	}
	return slowQueries, nil
}

func PopulateSlowRunningMetrics(instanceEntity *integration.Entity, conn *performanceDbConnection.PGSQLConnection, queryCpuMetricsMap map[string]map[int64]float64) []datamodels.SlowRunningQuery {
	isExtensionEnabled, err := validations.CheckPgStatStatementsExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_statements' is not enabled.")
		return nil
	}

	log.Info("Extension 'pg_stat_statements' enabled.")
	log.Info("queryCpuMetricsMap: %+v", queryCpuMetricsMap)
	slowQueries, err := GetSlowRunningMetrics(conn, queryCpuMetricsMap)
	if err != nil {
		log.Error("Error fetching slow-running queries: %v", err)
		return nil
	}

	if len(slowQueries) == 0 {
		log.Info("No slow-running queries found.")
		return nil
	}
	log.Info("Populate-slow running: %+v", slowQueries)

	for _, model := range slowQueries {
		metricSet := instanceEntity.NewMetricSet("PostgresSlowQueries")

		modelValue := reflect.ValueOf(model)
		modelType := reflect.TypeOf(model)

		for i := 0; i < modelValue.NumField(); i++ {
			field := modelValue.Field(i)
			fieldType := modelType.Field(i)
			metricName := fieldType.Tag.Get("metric_name")
			sourceType := fieldType.Tag.Get("source_type")

			if field.Kind() == reflect.Ptr && !field.IsNil() {
				common_utils.SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
			} else if field.Kind() != reflect.Ptr {
				common_utils.SetMetric(metricSet, metricName, field.Interface(), sourceType)
			}
		}

	}
	return slowQueries

}
