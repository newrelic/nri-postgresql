package queryperformancemonitoring

// this is the main go file for the query_monitoring package
import (
	"time"

	global_variables "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/global-variables"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/collection"
	performancedbconnection "github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/metrics"
	commonutils "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-utils"
	performancemetrics "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/performance-metrics"
)

func QueryPerformanceMain(args args.ArgumentList, pgIntegration *integration.Integration, databaseMap collection.DatabaseList) {
	connectionInfo := performancedbconnection.DefaultConnectionInfo(&args)
	if len(databaseMap) == 0 {
		log.Debug("No databases found")
		return
	}
	newConnection, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
	if err != nil {
		log.Error("Error creating connection: ", err)
		return
	}
	defer newConnection.Close()

	version, versionErr := metrics.CollectVersion(newConnection)
	versionInt := version.Major
	if versionErr != nil {
		log.Error("Error fetching version: ", versionErr)
		return
	}
	gv := global_variables.SetGlobalVariables(args, versionInt, commonutils.GetDatabaseListInString(databaseMap))
	populateQueryPerformanceMetrics(newConnection, pgIntegration, gv)
}

func populateQueryPerformanceMetrics(newConnection *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, gv *global_variables.GlobalVariables) {
	start := time.Now()
	log.Debug("Starting PopulateSlowRunningMetrics at ", start)
	slowRunningQueries := performancemetrics.PopulateSlowRunningMetrics(newConnection, pgIntegration, gv)
	log.Debug("PopulateSlowRunningMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateWaitEventMetrics at ", start)
	_ = performancemetrics.PopulateWaitEventMetrics(newConnection, pgIntegration, gv)
	log.Debug("PopulateWaitEventMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateBlockingMetrics at ", start)
	performancemetrics.PopulateBlockingMetrics(newConnection, pgIntegration, gv)
	log.Debug("PopulateBlockingMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateIndividualQueryMetrics at ", start)
	individualQueries := performancemetrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration, gv)
	log.Debug("PopulateIndividualQueryMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateExecutionPlanMetrics at ", start)
	performancemetrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration, gv)
	log.Debug("PopulateExecutionPlanMetrics completed in ", time.Since(start))
}
