package common_utils

import (
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/datamodels"
	"reflect"
)

func SetMetric(metricSet *metric.Set, name string, value interface{}, sourceType string) {
	switch sourceType {
	case `gauge`:
		metricSet.SetMetric(name, value, metric.GAUGE)
	case `attribute`:
		metricSet.SetMetric(name, value, metric.ATTRIBUTE)
	default:
		metricSet.SetMetric(name, value, metric.GAUGE)
	}
}

func IngestMetric(slowQueries []interface{}, instanceEntity *integration.Entity, eventName string) {

	for _, model := range slowQueries {
		if model == nil {
			continue
		}
		metricSet := instanceEntity.NewMetricSet(eventName)

		modelValue := reflect.ValueOf(model)
		modelType := reflect.TypeOf(model)

		for i := 0; i < modelValue.NumField(); i++ {
			field := modelValue.Field(i)
			fieldType := modelType.Field(i)
			metricName := fieldType.Tag.Get("metric_name")
			sourceType := fieldType.Tag.Get("source_type")

			if field.Kind() == reflect.Ptr && !field.IsNil() {
				SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
			} else if field.Kind() != reflect.Ptr {
				SetMetric(metricSet, metricName, field.Interface(), sourceType)
			}
		}
	}
}

func IngestSlowQueryMetrics(slowQueryMetrics []datamodels.SlowRunningQuery, instanceEntity *integration.Entity) {
	slowQueriesInterface := make([]interface{}, len(slowQueryMetrics))
	for i, v := range slowQueryMetrics {
		slowQueriesInterface[i] = v
	}
	IngestMetric(slowQueriesInterface, instanceEntity, "PostgresSlowQueries")
}

func IngestWaitEventMetrics(waitEventMetrics []datamodels.WaitEventQuery, instanceEntity *integration.Entity) {
	waitEventMetricsInterface := make([]interface{}, len(waitEventMetrics))
	for i, v := range waitEventMetricsInterface {
		waitEventMetricsInterface[i] = v
	}
	IngestMetric(waitEventMetricsInterface, instanceEntity, "PostgresWaitEvents")
}

func IngestIndividualQueryMetrics(individualQueryMetrics []datamodels.IndividualQuerySearch, instanceEntity *integration.Entity) {
	individualQueryMetricsInterface := make([]interface{}, len(individualQueryMetrics))
	for i, v := range individualQueryMetricsInterface {
		individualQueryMetricsInterface[i] = v
	}
	IngestMetric(individualQueryMetricsInterface, instanceEntity, "PostgresIndividualQueries")
}

func IngestExecutionPlanMetrics(executionPlanMetrics []datamodels.QueryExecutionPlanMetrics, instanceEntity *integration.Entity) {
	executionPlanMetricsInterface := make([]interface{}, len(executionPlanMetrics))
	for i, v := range executionPlanMetricsInterface {
		executionPlanMetricsInterface[i] = v
	}
	IngestMetric(executionPlanMetricsInterface, instanceEntity, "PostgresExecutionPlanMetrics")
}

func IngestBlockSessionMetrics(blockIoMetrics []datamodels.BlockingQuery, instanceEntity *integration.Entity) {
	blockIoMetricsInterface := make([]interface{}, len(blockIoMetrics))
	for i, v := range blockIoMetricsInterface {
		blockIoMetricsInterface[i] = v
	}
	IngestMetric(blockIoMetricsInterface, instanceEntity, "PostgresBlockingQueries")
}
