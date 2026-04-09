//go:generate goversioninfo
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	queryperformancemonitoring "github.com/newrelic/nri-postgresql/src/query-performance-monitoring"

	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/collection"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/inventory"
	"github.com/newrelic/nri-postgresql/src/metrics"
)

const (
	integrationName = "com.newrelic.postgresql"
)

var (
	integrationVersion = "0.0.0"
	gitCommit          = ""
	buildDate          = ""
)

func main() {

	var args args.ArgumentList
	// Create Integration
	pgIntegration, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if args.ShowVersion {
		fmt.Printf(
			"New Relic %s integration Version: %s, Platform: %s, GoVersion: %s, GitCommit: %s, BuildDate: %s\n",
			strings.Title(strings.Replace(integrationName, "com.newrelic.", "", 1)),
			integrationVersion,
			fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			runtime.Version(),
			gitCommit,
			buildDate)
		os.Exit(0)
	}

	// Setup logging with verbose
	log.SetupLogging(args.Verbose)

	// Validate arguments
	if err := args.Validate(); err != nil {
		log.Error("Configuration error for args %v: %s", args, err.Error())
		os.Exit(1)
	}

	connectionInfo := connection.DefaultConnectionInfo(&args)

	// Try to build collection list - if it fails, we might be connected to PgBouncer admin console
	// The detection will happen later in PopulateMetrics
	collectionList, err := collection.BuildCollectionList(args, connectionInfo)
	if err != nil {
		// Log as debug since this might be expected for PgBouncer admin console
		log.Debug("Unable to build collection list (this is normal for PgBouncer admin console): %s", err)
		// Initialize empty collection list - PopulateMetrics will handle detection
		collectionList = collection.DatabaseList{}
	}
	instance, err := pgIntegration.Entity(fmt.Sprintf("%s:%s", args.Hostname, args.Port), "pg-instance")
	if err != nil {
		log.Error("Error creating instance entity: %s", err.Error())
		os.Exit(1)
	}
	if args.HasMetrics() {
		metrics.PopulateMetrics(connectionInfo, collectionList, instance, pgIntegration, args.Pgbouncer, args.CollectDbLockMetrics, args.CollectBloatMetrics, args.CustomMetricsQuery)
		if args.CustomMetricsConfig != "" {
			metrics.PopulateCustomMetricsFromFile(connectionInfo, args.CustomMetricsConfig, pgIntegration)
		}
	}

	if args.HasInventory() {
		con, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
		if err != nil {
			log.Error("Inventory collection failed: error creating connection: %s", err.Error())
		} else {
			defer con.Close()

			// Try to detect database type - inventory only works for PostgreSQL
			connInfo, detectErr := metrics.DetectDatabaseType(con)
			if detectErr != nil {
				log.Warn("Unable to determine database type for inventory collection: %s", detectErr.Error())
			} else if connInfo.Type == metrics.DatabaseTypePgBouncerAdmin {
				log.Debug("Skipping inventory collection - connected to PgBouncer admin console")
			} else {
				inventory.PopulateInventory(instance, con)
			}
		}
	}

	if err = pgIntegration.Publish(); err != nil {
		log.Error(err.Error())
	}

	if args.EnableQueryMonitoring {
		queryperformancemonitoring.QueryPerformanceMain(args, pgIntegration, collectionList)
	}

}
