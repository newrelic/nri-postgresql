package commonutils_test

import (
	"testing"

	commonparameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	"github.com/stretchr/testify/assert"
)

func TestIngestMetric(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{
		Hostname: "localhost",
		Port:     "5432",
	}
	cp := commonparameters.SetCommonParameters(args, uint64(14), "testdb")
	metricList := []interface{}{
		struct {
			TestField int `metric_name:"testField" source_type:"gauge"`
		}{TestField: 123},
	}
	err := commonutils.IngestMetric(metricList, "testEvent", pgIntegration, cp)
	if err != nil {
		t.Error(err)
		return
	}
	assert.NotEmpty(t, pgIntegration.Entities)
}

func TestCreateEntity(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{
		Hostname: "localhost",
		Port:     "5432",
	}
	cp := commonparameters.SetCommonParameters(args, uint64(14), "testdb")

	entity, err := commonutils.CreateEntity(pgIntegration, cp)
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, "localhost:5432", entity.Metadata.Name)
}

func TestPublishMetrics(t *testing.T) {
	pgIntegration, _ := integration.New("test", "1.0.0")
	args := args.ArgumentList{
		Hostname: "localhost",
		Port:     "5432",
	}
	cp := commonparameters.SetCommonParameters(args, uint64(14), "testdb")
	entity, _ := commonutils.CreateEntity(pgIntegration, cp)

	err := commonutils.PublishMetrics(pgIntegration, &entity, cp)
	assert.NoError(t, err)
	assert.NotNil(t, entity)
}
