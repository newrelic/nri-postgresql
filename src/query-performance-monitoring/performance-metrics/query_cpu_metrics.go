package performance_metrics

import (
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	common_utils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
	"reflect"
)

func PopulateQueryCpuMetrics(instanceEntity *integration.Entity, conn *performanceDbConnection.PGSQLConnection) {
	isExtensionEnabled, err := validations.CheckPgStatMonitorExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_monitor' is not enabled.")
		return
	}
	log.Info("Extension 'pg_stat_monitor' enabled.")

	queryCpuMetricsList, err := GetQueryCpuMetrics(conn)
	if err != nil {
		log.Error("Error fetching Query CPU metrics: %v", err)
		return
	}
	if len(queryCpuMetricsList) == 0 {
		log.Info("No Query CPU metrics found.")
		return
	}
	for _, model := range queryCpuMetricsList {
		metricSet := instanceEntity.NewMetricSet("PostgresQueryCpuMetricsSample")

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
}

func GetQueryCpuMetrics(conn *performanceDbConnection.PGSQLConnection) ([]datamodels.QueryCpuMetrics, error) {
	var queryCpuMetricsList []datamodels.QueryCpuMetrics
	var query = queries.QueryCpuMetrics
	rows, err := conn.Queryx(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var queryCpuMetrics datamodels.QueryCpuMetrics
		if err := rows.StructScan(&queryCpuMetrics); err != nil {
			return nil, err
		}
		queryCpuMetricsList = append(queryCpuMetricsList, queryCpuMetrics)
	}

	return queryCpuMetricsList, nil
}
