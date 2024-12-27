package query_performance_monitoring

// this is the main go file for the query_monitoring package
import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	performanceDbConnection "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/connections"
	"time"

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
	start := time.Now()
	log.Info("Start PopulateSlowRunningMetrics:", start)
	slowRunningQueries := performance_metrics.PopulateSlowRunningMetrics(newConnection, pgIntegration, args)
	log.Info("End PopulateSlowRunningMetrics:", time.Since(start).Seconds())

	start = time.Now()
	log.Info("Start PopulateWaitEventMetrics:", start)
	performance_metrics.PopulateWaitEventMetrics(newConnection, pgIntegration, args)
	log.Info("End PopulateWaitEventMetrics:", time.Since(start).Seconds())

	start = time.Now()
	log.Info("Start PopulateBlockingMetrics:", start)
	performance_metrics.PopulateBlockingMetrics(newConnection, pgIntegration, args)
	log.Info("End PopulateBlockingMetrics:", time.Since(start).Seconds())

	start = time.Now()
	log.Info("Start PopulateIndividualQueryMetrics:", start)
	individualQueries := performance_metrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration, args)
	log.Info("End PopulateIndividualQueryMetrics:", time.Since(start).Seconds())

	start = time.Now()
	log.Info("Start PopulateExecutionPlanMetrics:", start)
	performance_metrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration, args)
	log.Info("End PopulateExecutionPlanMetrics:", time.Since(start).Seconds())

	log.Info("Query analysis completed.")
}
