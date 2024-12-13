package query_performance_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(instanceEntity *integration.Entity, args args.ArgumentList) {
	connectionInfo := performanceDbConnection.DefaultConnectionInfo(&args)
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		fmt.Println("Error creating connection: ", err)
		return
	}
	slowRunningQueries := performance_metrics.PopulateSlowRunningMetrics(instanceEntity, newConnection)
	performance_metrics.PopulateWaitEventMetrics(instanceEntity, newConnection)
	performance_metrics.PopulateBlockingMetrics(instanceEntity, newConnection)
	individualQueries := performance_metrics.PopulateIndividualQueryMetrics(instanceEntity, newConnection, slowRunningQueries)
	performance_metrics.PopulateExecutionPlanMetrics(instanceEntity, individualQueries, args)
	fmt.Println("Query analysis completed.")
}