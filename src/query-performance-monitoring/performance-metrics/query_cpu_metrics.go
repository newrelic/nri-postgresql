package performance_metrics

import (
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/queries"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"
)

func PopulateQueryCpuMetrics(conn *performanceDbConnection.PGSQLConnection) map[string]map[int64]float64 {
	isExtensionEnabled, err := validations.CheckPgStatMonitorExtensionEnabled(conn)
	if err != nil {
		log.Error("Error executing query: %v", err)
		return nil
	}
	if !isExtensionEnabled {
		log.Info("Extension 'pg_stat_monitor' is not enabled.")
		return nil
	}
	log.Info("Extension 'pg_stat_monitor' enabled.")

	queryCpuMetricsList, err := GetQueryCpuMetrics(conn)
	if err != nil {
		log.Error("Error fetching Query CPU metrics: %v", err)
		return nil
	}
	if len(queryCpuMetricsList) == 0 {
		log.Info("No Query CPU metrics found.")
		return nil
	}
	//for _, model := range queryCpuMetricsList {
	//	metricSet := instanceEntity.NewMetricSet("PostgresQueryCpuMetrics")
	//
	//	modelValue := reflect.ValueOf(model)
	//	modelType := reflect.TypeOf(model)
	//
	//	for i := 0; i < modelValue.NumField(); i++ {
	//		field := modelValue.Field(i)
	//		fieldType := modelType.Field(i)
	//		metricName := fieldType.Tag.Get("metric_name")
	//		sourceType := fieldType.Tag.Get("source_type")
	//
	//		if field.Kind() == reflect.Ptr && !field.IsNil() {
	//			common_utils.SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
	//		} else if field.Kind() != reflect.Ptr {
	//			common_utils.SetMetric(metricSet, metricName, field.Interface(), sourceType)
	//		}
	//	}
	//}
	queryCpuMetricsMap := ConvertToCpuMetricsMap(queryCpuMetricsList)
	return queryCpuMetricsMap
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

func ConvertToCpuMetricsMap(queryCpuMetricsList []datamodels.QueryCpuMetrics) map[string]map[int64]float64 {
	cpuMetricsMap := make(map[string]map[int64]float64)

	for _, metric := range queryCpuMetricsList {
		dbName := *metric.DatabaseName
		queryID := *metric.QueryId
		avgCpuTime := *metric.AvgCpuTime

		if _, exists := cpuMetricsMap[dbName]; !exists {
			cpuMetricsMap[dbName] = make(map[int64]float64)
		}
		cpuMetricsMap[dbName][queryID] = avgCpuTime
	}

	return cpuMetricsMap
}
