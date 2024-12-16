package query_performance_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(args args.ArgumentList, pgIntegration *integration.Integration) {
	connectionInfo := performanceDbConnection.DefaultConnectionInfo(&args)
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		fmt.Println("Error creating connection: ", err)
		return
	}
	slowRunningQueries := performance_metrics.PopulateSlowRunningMetrics(newConnection, pgIntegration)
	performance_metrics.PopulateWaitEventMetrics(newConnection, pgIntegration)
	performance_metrics.PopulateBlockingMetrics(newConnection, pgIntegration)
	individualQueries := performance_metrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration)
	performance_metrics.PopulateExecutionPlanMetrics(individualQueries, args, pgIntegration)
	fmt.Println("Query analysis completed.")
}
