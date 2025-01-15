package queryperformancemonitoring

// this is the main go file for the query_monitoring package
import (
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/collection"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/metrics"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(args args.ArgumentList, pgIntegration *integration.Integration, databaseList collection.DatabaseList) {
	connectionInfo := performancedbconnection.DefaultConnectionInfo(&args)
	databaseStringList := commonutils.GetDatabaseListInString(databaseList)
	if len(databaseList) == 0 {
		log.Debug("No databases found")
		return
	}

	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		log.Debug("Error creating connection: ", err)
		return
	}

	version, versionErr := metrics.CollectVersion(newConnection)
	versionInt := version.Major
	if versionErr != nil {
		log.Debug("Error fetching version: ", versionErr)
		return
	}

	start := time.Now()
	log.Debug("Starting PopulateSlowRunningMetrics at ", start)
	slowRunningQueries := performancemetrics.PopulateSlowRunningMetrics(newConnection, pgIntegration, args, databaseStringList, versionInt)
	log.Debug("PopulateSlowRunningMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateWaitEventMetrics at ", start)
	_ = performancemetrics.PopulateWaitEventMetrics(newConnection, pgIntegration, args, databaseStringList, versionInt)
	log.Debug("PopulateWaitEventMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateBlockingMetrics at ", start)
	_ = performancemetrics.PopulateBlockingMetrics(newConnection, pgIntegration, args, databaseStringList, versionInt)
	log.Debug("PopulateBlockingMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateIndividualQueryMetrics at ", start)
	individualQueries := performancemetrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration, args, databaseStringList, versionInt)
	log.Debug("PopulateIndividualQueryMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateExecutionPlanMetrics at ", start)
	performancemetrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration, args)
	log.Debug("PopulateExecutionPlanMetrics completed in ", time.Since(start))

	log.Debug("Query analysis completed.")
}
