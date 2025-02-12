package queryperformancemonitoring

// this is the main go file for the query_monitoring package
import (
	"time"

	"github.com/newrelic/nri-postgresql/src/query-performance-monitoring/validations"

	common_parameters "github.com/newrelic/nri-postgresql/src/query-performance-monitoring/common-parameters"

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
	if versionErr != nil {
		log.Error("Error fetching version: ", versionErr)
		return
	}
	versionInt := version.Major
	if !validations.CheckPostgresVersionSupportForQueryMonitoring(versionInt) {
		log.Debug("Postgres version: %d is not supported for query monitoring", versionInt)
		return
	}
	cp := common_parameters.SetCommonParameters(args, versionInt, commonutils.GetDatabaseListInString(databaseMap))

	populateQueryPerformanceMetrics(newConnection, pgIntegration, cp, connectionInfo)
}

func populateQueryPerformanceMetrics(newConnection *performancedbconnection.PGSQLConnection, pgIntegration *integration.Integration, cp *common_parameters.CommonParameters, connectionInfo performancedbconnection.Info) {
	enabledExtensions, err := validations.FetchAllExtensions(newConnection)
	if err != nil {
		log.Error("Error fetching extensions: ", err)
		return
	}
	start := time.Now()
	log.Debug("Starting PopulateSlowRunningMetrics at ", start)
	slowRunningQueries := performancemetrics.PopulateSlowRunningMetrics(newConnection, pgIntegration, cp, enabledExtensions)
	log.Debug("PopulateSlowRunningMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateWaitEventMetrics at ", start)
	_ = performancemetrics.PopulateWaitEventMetrics(newConnection, pgIntegration, cp, enabledExtensions)
	log.Debug("PopulateWaitEventMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateBlockingMetrics at ", start)
	performancemetrics.PopulateBlockingMetrics(newConnection, pgIntegration, cp, enabledExtensions)
	log.Debug("PopulateBlockingMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateIndividualQueryMetrics at ", start)
	individualQueries := performancemetrics.PopulateIndividualQueryMetrics(newConnection, slowRunningQueries, pgIntegration, cp, enabledExtensions)
	log.Debug("PopulateIndividualQueryMetrics completed in ", time.Since(start))

	start = time.Now()
	log.Debug("Starting PopulateExecutionPlanMetrics at ", start)
	performancemetrics.PopulateExecutionPlanMetrics(individualQueries, pgIntegration, cp, connectionInfo)
	log.Debug("PopulateExecutionPlanMetrics completed in ", time.Since(start))
}
