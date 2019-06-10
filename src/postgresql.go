package main

import (
	"fmt"
	"os"

	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/collection"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/inventory"
	"github.com/newrelic/nri-postgresql/src/metrics"
)

const (
	integrationName    = "com.newrelic.postgresql"
	integrationVersion = "2.1.1"
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

	connectionInfo := connection.DefaultConnectionInfo(&args)
	collectionList, err := collection.BuildCollectionList(args, connectionInfo)
	if err != nil {
		log.Error("Error creating list of entities to collect: %s", err)
		os.Exit(1)
	}

	instance, err := postgresIntegration.Entity(fmt.Sprintf("%s:%s", args.Hostname, args.Port), "pg-instance")
	if err != nil {
		log.Error("Error creating instance entity: %s", err.Error())
		os.Exit(1)
	}

	if args.HasMetrics() {
		metrics.PopulateMetrics(connectionInfo, collectionList, instance, postgresIntegration, args.Pgbouncer)
	}

	if args.HasInventory() {
		con, err := connectionInfo.NewConnection(connectionInfo.DatabaseName())
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
