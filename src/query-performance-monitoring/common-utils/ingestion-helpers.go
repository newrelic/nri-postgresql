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
		err := metricSet.SetMetric(name, value, metric.GAUGE)
		if err != nil {
			return
		}
	case `attribute`:
		err := metricSet.SetMetric(name, value, metric.ATTRIBUTE)
		if err != nil {
			return
		}
	default:
		err := metricSet.SetMetric(name, value, metric.GAUGE)
		if err != nil {
			return
		}
	}
}

func IngestMetric(slowQueries []interface{}, instanceEntity *integration.Entity, eventName string) {

	for _, model := range slowQueries {
		if model == nil {
			continue
		}
		metricSet := instanceEntity.NewMetricSet(eventName)

		modelValue := reflect.ValueOf(model)
		if modelValue.Kind() == reflect.Ptr {
			modelValue = modelValue.Elem()
		}
		if !modelValue.IsValid() || modelValue.Kind() != reflect.Struct {
			continue
		}

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
	slowQueriesInterface := make([]interface{}, 0)
	for _, v := range slowQueryMetrics {
		slowQueriesInterface = append(slowQueriesInterface, v)
	}
	IngestMetric(slowQueriesInterface, instanceEntity, "PostgresSlowQueries")
}

func IngestWaitEventMetrics(waitEventMetrics []datamodels.WaitEventQuery, instanceEntity *integration.Entity) {
	waitEventMetricsInterface := make([]interface{}, 0)
	for _, v := range waitEventMetrics {
		waitEventMetricsInterface = append(waitEventMetricsInterface, v)
	}
	IngestMetric(waitEventMetricsInterface, instanceEntity, "PostgresWaitEvents")
}

func IngestIndividualQueryMetrics(individualQueryMetrics []datamodels.IndividualQuerySearch, instanceEntity *integration.Entity) {
	individualQueryMetricsInterface := make([]interface{}, 0)
	for _, v := range individualQueryMetrics {
		individualQueryMetricsInterface = append(individualQueryMetricsInterface, v)
	}
	IngestMetric(individualQueryMetricsInterface, instanceEntity, "PostgresIndividualQueries")
}

func IngestExecutionPlanMetrics(executionPlanMetrics []datamodels.QueryExecutionPlanMetrics, instanceEntity *integration.Entity) {
	executionPlanMetricsInterface := make([]interface{}, 0)
	for _, v := range executionPlanMetrics {
		executionPlanMetricsInterface = append(executionPlanMetricsInterface, v)
	}
	IngestMetric(executionPlanMetricsInterface, instanceEntity, "PostgresExecutionPlanMetrics")
}

func IngestBlockSessionMetrics(blockIoMetrics []datamodels.BlockingQuery, instanceEntity *integration.Entity) {
	blockIoMetricsInterface := make([]interface{}, 0)
	for _, v := range blockIoMetrics {
		blockIoMetricsInterface = append(blockIoMetricsInterface, v)
	}
	IngestMetric(blockIoMetricsInterface, instanceEntity, "PostgresBlockingQueries")
}
