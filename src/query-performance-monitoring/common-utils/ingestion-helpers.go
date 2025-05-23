package commonutils

import (
	"fmt"
	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
)

// IngestMetric is a util by which we publish data in batches .Reason for this is to avoid publishing large data in one go and its a limitation for NewRelic.
func IngestMetric(metricList []interface{}, eventName string, pgIntegration *integration.Integration, cp *commonparameters.CommonParameters) error {
	instanceEntity, err := CreateEntity(pgIntegration, cp)
	if err != nil {
		log.Error("Error creating entity: %v", err)
		return err
	}
	metricCount := 0
	for _, model := range metricList {
		if model == nil {
			continue
		}
		metricCount += 1
		metricSet := instanceEntity.NewMetricSet(eventName)
		marshalErr := metricSet.MarshalMetrics(model)
		if marshalErr != nil {
			log.Error("Error processing model: %v", marshalErr)
			continue
		}
		if metricCount == PublishThreshold {
			metricCount = 0
			if err := PublishMetrics(pgIntegration, &instanceEntity, cp); err != nil {
				log.Error("Error publishing metrics: %v", err)
				return err
			}
		}
	}
	if metricCount > 0 {
		if err := PublishMetrics(pgIntegration, &instanceEntity, cp); err != nil {
			log.Error("Error publishing metrics: %v", err)
			return err
		}
	}
	return nil
}

func CreateEntity(pgIntegration *integration.Integration, cp *commonparameters.CommonParameters) (*integration.Entity, error) {
	return pgIntegration.Entity(fmt.Sprintf("%s:%s", cp.Host, cp.Port), "pg-instance")
}

func PublishMetrics(pgIntegration *integration.Integration, instanceEntity **integration.Entity, cp *commonparameters.CommonParameters) error {
	if err := pgIntegration.Publish(); err != nil {
		log.Error("Error publishing query performance metrics")
		return err
	}
	var err error
	*instanceEntity, err = CreateEntity(pgIntegration, cp)
	return err
}
