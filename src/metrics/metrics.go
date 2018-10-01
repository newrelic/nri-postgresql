package metrics

import (
	"reflect"

	"github.com/blang/semver"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/newrelic/nri-postgresql/src/connection"
)

// PopulateInstanceMetrics populates the metrics for an instance
func PopulateInstanceMetrics(instanceEntity *integration.Entity, version semver.Version, connection *connection.PGSQLConnection) {
	metricSet := instanceEntity.NewMetricSet("PostgreSQLInstanceSample",
		metric.Attribute{Key: "displayName", Value: instanceEntity.Metadata.Name},
		metric.Attribute{Key: "entityName", Value: instanceEntity.Metadata.Namespace + ":" + instanceEntity.Metadata.Name},
	)

	for _, queryDef := range generateInstanceDefinitions(version) {
		dataModels := queryDef.GetDataModels()
		if err := connection.Query(dataModels, queryDef.GetQuery()); err != nil {
			log.Error("Could not execute instance query: %s", err.Error())
			continue
		}

		vp := reflect.Indirect(reflect.ValueOf(dataModels))
		vpInterface := vp.Index(0).Interface()
		err := metricSet.MarshalMetrics(vpInterface)
		if err != nil {
			log.Error("Could not parse metrics from instance query result: %s", err.Error())
		}
	}
}

// PopulateDatabaseMetrics populates the metrics for a database
func PopulateDatabaseMetrics(databases args.DatabaseList, version semver.Version, integration *integration.Integration, connection *connection.PGSQLConnection) {
	databaseDefinitions := generateDatabaseDefinitions(databases, version)

	for _, queryDef := range databaseDefinitions {
		// collect into model
		dataModels := queryDef.GetDataModels()
		if err := connection.Query(dataModels, queryDef.GetQuery()); err != nil {
			log.Error("Could not execute database query: %s", err.Error())
			continue
		}

		// for each row in the response
		v := reflect.Indirect(reflect.ValueOf(dataModels))
		for i := 0; i < v.Len(); i++ {
			db := v.Index(i).Interface()
			name, err := GetDatabaseName(db)
			if err != nil {
				log.Error("Unable to get database name: %s", err.Error())
			}

			databaseEntity, err := integration.Entity(name, "database")
			if err != nil {
				log.Error("Failed to get database entity for name %s: %s", name, err.Error())
			}
			metricSet := databaseEntity.NewMetricSet("PostgreSQLDatabaseSample",
				metric.Attribute{Key: "displayName", Value: databaseEntity.Metadata.Name},
				metric.Attribute{Key: "entityName", Value: databaseEntity.Metadata.Namespace + ":" + databaseEntity.Metadata.Name},
			)

			if err := metricSet.MarshalMetrics(db); err != nil {
				log.Error("Failed to database entity with metrics: %s", err.Error())
			}

		}
	}
}

// PopulateTableMetrics populates the metrics for a table
func PopulateTableMetrics(databases args.DatabaseList, integration *integration.Integration, ci *connection.Info) {
	for database, schemaList := range databases {
		// Create a new connection to the database
		ci.Database = database
		con, err := ci.NewConnection()
		if err != nil {
			log.Error("Failed to connect to database %s: %s", database, err.Error())
			continue
		}

		populateTableMetricsForDatabase(schemaList, con, integration)
	}
}

func populateTableMetricsForDatabase(schemaList args.SchemaList, con *connection.PGSQLConnection, integration *integration.Integration) {
	tableDefinitions := generateTableDefinitions(schemaList)

	// collect into model
	for _, definition := range tableDefinitions {

		dataModels := definition.GetDataModels()
		if err := con.Query(dataModels, definition.GetQuery()); err != nil {
			log.Error("Could not execute table query: %s", err.Error())
			return
		}

		// for each row in the response
		v := reflect.Indirect(reflect.ValueOf(dataModels))
		for i := 0; i < v.Len(); i++ {
			row := v.Index(i).Interface()
			dbName, err := GetDatabaseName(row)
			if err != nil {
				log.Error("Unable to get database name: %s", err.Error())
			}
			schemaName, err := GetSchemaName(row)
			if err != nil {
				log.Error("Unable to get schema name: %s", err.Error())
			}
			tableName, err := GetTableName(row)
			if err != nil {
				log.Error("Unable to get table name: %s", err.Error())
			}

			tableEntity, err := integration.Entity(tableName, "table")
			if err != nil {
				log.Error("Failed to get table entity for table %s: %s", tableName, err.Error())
			}
			metricSet := tableEntity.NewMetricSet("PostgreSQLTableSample",
				metric.Attribute{Key: "displayName", Value: tableEntity.Metadata.Name},
				metric.Attribute{Key: "entityName", Value: tableEntity.Metadata.Namespace + ":" + tableEntity.Metadata.Name},
				metric.Attribute{Key: "database", Value: dbName},
				metric.Attribute{Key: "schema", Value: schemaName},
			)

			if err := metricSet.MarshalMetrics(row); err != nil {
				log.Error("Failed to populate table entity with metrics: %s", err.Error())
			}

		}
	}
}

// PopulateIndexMetrics populates the metrics for an index
func PopulateIndexMetrics(databases args.DatabaseList, integration *integration.Integration, ci *connection.Info) {
	for database, schemaList := range databases {
		ci.Database = database
		con, err := ci.NewConnection()
		if err != nil {
			log.Error("Failed to create new connection to database %s: %s", database, err.Error())
			continue
		}
		populateIndexMetricsForDatabase(schemaList, con, integration)
	}
}

func populateIndexMetricsForDatabase(schemaList args.SchemaList, con *connection.PGSQLConnection, integration *integration.Integration) {
	indexDefinition := generateIndexDefinitions(schemaList)

	// collect into model
	dataModels := indexDefinition.GetDataModels()
	if err := con.Query(dataModels, indexDefinition.GetQuery()); err != nil {
		log.Error("Could not execute index query: %s", err.Error())
		return
	}

	// for each row in the response
	v := reflect.Indirect(reflect.ValueOf(dataModels))
	for i := 0; i < v.Len(); i++ {
		row := v.Index(i).Interface()
		dbName, err := GetDatabaseName(row)
		if err != nil {
			log.Error("Unable to get database name: %s", err.Error())
		}
		schemaName, err := GetSchemaName(row)
		if err != nil {
			log.Error("Unable to get schema name: %s", err.Error())
		}
		tableName, err := GetTableName(row)
		if err != nil {
			log.Error("Unable to get table name: %s", err.Error())
		}
		indexName, err := GetIndexName(row)
		if err != nil {
			log.Error("Unable to get index name: %s", err.Error())
		}

		indexEntity, err := integration.Entity(indexName, "index")
		if err != nil {
			log.Error("Failed to get table entity for index %s: %s", indexName, err.Error())
		}
		metricSet := indexEntity.NewMetricSet("PostgreSQLIndexSample",
			metric.Attribute{Key: "displayName", Value: indexEntity.Metadata.Name},
			metric.Attribute{Key: "entityName", Value: indexEntity.Metadata.Namespace + ":" + indexEntity.Metadata.Name},
			metric.Attribute{Key: "database", Value: dbName},
			metric.Attribute{Key: "schema", Value: schemaName},
			metric.Attribute{Key: "table", Value: tableName},
		)

		if err := metricSet.MarshalMetrics(row); err != nil {
			log.Error("Failed to populate index entity with metrics: %s", err.Error())
		}

	}

}

// PopulatePgBouncerMetrics populates pgbouncer metrics
func PopulatePgBouncerMetrics(integration *integration.Integration, con *connection.PGSQLConnection) {
	pgbouncerDefs := generatePgBouncerDefinitions()

	for _, definition := range pgbouncerDefs {
		dataModels := definition.GetDataModels()
		if err := con.Query(dataModels, definition.GetQuery()); err != nil {
			log.Error("Could not execute index query: %s", err.Error())
			return
		}

		// for each row in the response
		v := reflect.Indirect(reflect.ValueOf(dataModels))
		for i := 0; i < v.Len(); i++ {
			db := v.Index(i).Interface()
			name, err := GetDatabaseName(db)
			if err != nil {
				log.Error("Unable to get database name: %s", err.Error())
				continue
			}

			pgEntity, err := integration.Entity(name, "pgbouncer")
			if err != nil {
				log.Error("Failed to get database entity for name %s: %s", name, err.Error())
			}
			metricSet := pgEntity.NewMetricSet("PgBouncerSample",
				metric.Attribute{Key: "displayName", Value: name},
				metric.Attribute{Key: "entityName", Value: "pgbouncer:" + name},
			)

			if err := metricSet.MarshalMetrics(db); err != nil {
				log.Error("Failed to populate pgbouncer entity with metrics: %s", err.Error())
			}

		}
	}
}
