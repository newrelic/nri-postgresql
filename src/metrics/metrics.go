package metrics

import (
	"github.com/blang/semver"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/nri-postgresql/src/connection"
)

func PopulateInstanceMetrics(instanceEntity *integration.Entity, version semver.Version, connection *connection.PGSQLConnection) {
	metricSet := instanceEntity.NewMetricSet("PostgreSQLInstanceSample",
		metric.Attribute{Key: "displayName", Value: instanceEntity.Metadata.Name},
		metric.Attribute{Key: "entityName", Value: instanceEntity.Metadata.Namespace + ":" + instanceEntity.Metadata.Name},
	)

	println(len(metricSet.Metrics))
	// TODO collect instance metrics based on version
}
