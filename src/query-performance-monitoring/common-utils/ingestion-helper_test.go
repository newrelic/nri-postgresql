package commonutils_test

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/stretchr/testify/assert"
)

func TestSetMetric(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	entity, _ := pgIntegration.Entity("test-entity", "test-type")

	metricSet := entity.NewMetricSet("test-event")

	commonutils.SetMetric(metricSet, "testGauge", 123.0, "gauge")
	assert.Equal(t, 123.0, metricSet.Metrics["testGauge"])

	commonutils.SetMetric(metricSet, "testAttribute", "value", "attribute")
	assert.Equal(t, "value", metricSet.Metrics["testAttribute"])

	commonutils.SetMetric(metricSet, "testDefault", 456.0, "unknown")
	assert.Equal(t, 456.0, metricSet.Metrics["testDefault"])
}

func TestIngestMetric(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{
		Hostname: "localhost",
		Port:     "5432",
	}
	metricList := []interface{}{
		struct {
			TestField int `metric_name:"testField" source_type:"gauge"`
		}{TestField: 123},
	}

	commonutils.IngestMetric(metricList, "testEvent", pgIntegration, args)
	assert.NotEmpty(t, pgIntegration.Entities)
}

func TestCreateEntity(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{
		Hostname: "localhost",
		Port:     "5432",
	}

	entity, err := commonutils.CreateEntity(pgIntegration, args)
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, "localhost:5432", entity.Metadata.Name)
}

func TestProcessModel(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	entity, _ := pgIntegration.Entity("test-entity", "test-type")

	metricSet := entity.NewMetricSet("test-event")

	model := struct {
		TestField int `metric_name:"testField" source_type:"gauge"`
	}{TestField: 123}

	err := commonutils.ProcessModel(model, metricSet)
	assert.NoError(t, err)
	assert.Equal(t, 123.0, metricSet.Metrics["testField"])
}

func TestPublishMetrics(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{
		Hostname: "localhost",
		Port:     "5432",
	}
	entity, _ := commonutils.CreateEntity(pgIntegration, args)

	err := commonutils.PublishMetrics(pgIntegration, &entity, args)
	assert.NoError(t, err)
	assert.NotNil(t, entity)
}
