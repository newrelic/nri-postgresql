package main

import (
	"os"

	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
)

type argumentList struct {
	sdkArgs.DefaultArgumentList
}

const (
	integrationName    = "com.New Relic.postgresql"
	integrationVersion = "0.1.0"
)

var (
	args argumentList
)

func main() {
	// Create Integration
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Create Entity, entities name must be unique
	e1, err := i.Entity("instance-1", "custom")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	// Add Event
	if args.All() || args.Events {
		err = e1.AddEvent(event.New("restart", "status"))
		if err != nil {
			log.Error(err.Error())
		}
	}

	// Add Inventory item
	if args.All() || args.Inventory {
		err = e1.SetInventoryItem("instance", "version", "3.0.1")
		if err != nil {
			log.Error(err.Error())
		}
	}

	// Add Metric
	if args.All() || args.Metrics {
		m1 := e1.NewMetricSet("CustomSample")
		err = m1.SetMetric("some-data", 1000, metric.GAUGE)
		if err != nil {
			log.Error(err.Error())
		}
	}

	// Create another Entity
	e2, err := i.Entity("instance-2", "custom")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	if args.All() || args.Inventory {
		err = e2.SetInventoryItem("instance", "version", "3.0.4")
		if err != nil {
			log.Error(err.Error())
		}
	}

	if args.All() || args.Metrics {
		m2 := e2.NewMetricSet("CustomSample")
		err = m2.SetMetric("some-data", 2000, metric.GAUGE)
		if err != nil {
			log.Error(err.Error())
		}
	}

	if err = i.Publish(); err != nil {
		log.Error(err.Error())
	}
}
