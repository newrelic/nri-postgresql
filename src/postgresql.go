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
	integrationVersion = "0.1.0"
)

func main() {
	var args args.ArgumentList
	// Create Integration
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
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

	// Create a new connection
	ci := connection.DefaultConnectionInfo(&args)
	con, err := ci.NewConnection()
	if err != nil {
		log.Error("Error creating connection to SQL Server: %s", err.Error())
		os.Exit(1)
	}

	instance, err := i.Entity(args.Hostname, "instance")
	if err != nil {
		log.Error("Error creating instance entity: %s", err.Error())
		os.Exit(1)
	}
	version, err := collectVersion(con)
	if err != nil {
		log.Error("Error collecting version number: %s", err.Error())
		os.Exit(1)
	}

	if args.HasInventory() {
		inventory.PopulateInventory(instance, con)
	}

	if args.HasMetrics() {
		metrics.PopulateInstanceMetrics(instance, version, con)
		metrics.PopulateDatabaseMetrics(databaseList, version, i, con)
		metrics.PopulateTableMetrics(databaseList, i, ci)
		metrics.PopulateIndexMetrics(databaseList, i, ci)
		if args.Pgbouncer {
			ci.Database = "pgbouncer"
			con, err = ci.NewConnection()
			if err != nil {
				log.Error("Error creating connection to pgbouncer database: %s", err)
			} else {
				metrics.PopulatePgBouncerMetrics(i, con)
			}
		}
	}

	if err = i.Publish(); err != nil {
		log.Error(err.Error())
	}
}
