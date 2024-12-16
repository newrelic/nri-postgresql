package query_performance_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(instanceEntity *integration.Entity, args args.ArgumentList, pgIntegration *integration.Integration) {
	connectionInfo := performanceDbConnection.DefaultConnectionInfo(&args)
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		fmt.Println("Error creating connection: ", err)
		return
	}
	slowRunningQueries := performance_metrics.PopulateSlowRunningMetrics(instanceEntity, newConnection, pgIntegration)
	performance_metrics.PopulateWaitEventMetrics(instanceEntity, newConnection, pgIntegration)
	performance_metrics.PopulateBlockingMetrics(instanceEntity, newConnection, pgIntegration)
	individualQueries := performance_metrics.PopulateIndividualQueryMetrics(instanceEntity, newConnection, slowRunningQueries, pgIntegration)
	performance_metrics.PopulateExecutionPlanMetrics(instanceEntity, individualQueries, args, pgIntegration)
	fmt.Println("Query analysis completed.")
}
