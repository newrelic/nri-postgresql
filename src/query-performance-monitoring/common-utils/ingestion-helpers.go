package commonutils

import (
	"fmt"
	"reflect"

	globalvariables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"

	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
)

func SetMetric(metricSet *metric.Set, name string, value interface{}, sourceType string) {
	switch sourceType {
	case `gauge`:
		err := metricSet.SetMetric(name, value, metric.GAUGE)
		if err != nil {
			log.Error("Error setting metric: %v", err)
			return
		}
	case `attribute`:
		err := metricSet.SetMetric(name, value, metric.ATTRIBUTE)
		if err != nil {
			log.Error("Error setting metric: %v", err)
			return
		}
	default:
		err := metricSet.SetMetric(name, value, metric.GAUGE)
		if err != nil {
			log.Error("Error setting metric: %v", err)
			return
		}
	}
}

// IngestMetric is a util by which we publish data in batches .Reason for this is to avoid publishing large data in one go and its a limitation for NewRelic.
func IngestMetric(metricList []interface{}, eventName string, pgIntegration *integration.Integration, gv *globalvariables.GlobalVariables) {
	instanceEntity, err := CreateEntity(pgIntegration, gv)
	if err != nil {
		log.Error("Error creating entity: %v", err)
		return
	}

	metricCount := 0
	lenOfMetricList := len(metricList)

	for _, model := range metricList {
		if model == nil {
			continue
		}
		metricCount += 1
		metricSet := instanceEntity.NewMetricSet(eventName)

		processErr := ProcessModel(model, metricSet)
		if processErr != nil {
			log.Error("Error processing model: %v", processErr)
			continue
		}

		if metricCount == PublishThreshold || metricCount == lenOfMetricList {
			metricCount = 0
			if err := PublishMetrics(pgIntegration, &instanceEntity, gv); err != nil {
				log.Error("Error publishing metrics: %v", err)
				return
			}
		}
	}
	if metricCount > 0 {
		if err := PublishMetrics(pgIntegration, &instanceEntity, gv); err != nil {
			log.Error("Error publishing metrics: %v", err)
			return
		}
	}
}

func CreateEntity(pgIntegration *integration.Integration, gv *globalvariables.GlobalVariables) (*integration.Entity, error) {
	return pgIntegration.Entity(fmt.Sprintf("%s:%s", gv.Hostname, gv.Port), "pg-instance")
}

func ProcessModel(model interface{}, metricSet *metric.Set) error {
	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
	}
	if !modelValue.IsValid() || modelValue.Kind() != reflect.Struct {
		log.Error("Invalid model type: %v", modelValue.Kind())
		return ErrInvalidModelType
	}

	modelType := reflect.TypeOf(model)

	for i := 0; i < modelValue.NumField(); i++ {
		field := modelValue.Field(i)
		fieldType := modelType.Field(i)
		metricName := fieldType.Tag.Get("metric_name")
		sourceType := fieldType.Tag.Get("source_type")
		ingestData := fieldType.Tag.Get("ingest_data")

		if ingestData == "false" {
			continue
		}

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			SetMetric(metricSet, metricName, field.Elem().Interface(), sourceType)
		} else if field.Kind() != reflect.Ptr {
			SetMetric(metricSet, metricName, field.Interface(), sourceType)
		}
	}
	return nil
}

func PublishMetrics(pgIntegration *integration.Integration, instanceEntity **integration.Entity, gv *globalvariables.GlobalVariables) error {
	if err := pgIntegration.Publish(); err != nil {
		return err
	}
	var err error
	*instanceEntity, err = CreateEntity(pgIntegration, gv)
	return err
}
