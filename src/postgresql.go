package main

import (
	"os"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/inventory"
	"github.com/newrelic/nri-postgresql/src/metrics"
)

const (
	integrationName    = "com.newrelic.postgresql"
	integrationVersion = "1.1.0"
)

func main() {
	var args args.ArgumentList
	// Create Integration
	postgresIntegration, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Setup logging with verbose
	log.SetupLogging(args.Verbose)

	// Validate arguments
	if err := args.Validate(); err != nil {
		log.Error("Configuration error for args %v: %s", args, err.Error())
		os.Exit(1)
	}

	databaseList := args.GetCollectionList()
	connectionInfo := connection.DefaultConnectionInfo(&args)

	instance, err := postgresIntegration.Entity(args.Hostname, "instance")
	if err != nil {
		log.Error("Error creating instance entity: %s", err.Error())
		os.Exit(1)
	}

	if args.HasMetrics() {
		metrics.PopulateMetrics(connectionInfo, databaseList, instance, postgresIntegration, args.Pgbouncer)
	}

	if args.HasInventory() {
		con, err := connectionInfo.NewConnection(connectionInfo.Databasename())
		if err != nil {
			log.Error("Inventory collection failed: error creating connection to PostgreSQL: %s", err.Error())
		} else {
			defer con.Close()
			inventory.PopulateInventory(instance, con)
		}
	}

	if err = postgresIntegration.Publish(); err != nil {
		log.Error(err.Error())
	}
}
