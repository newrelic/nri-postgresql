package metrics

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"sync"

	"github.com/blang/semver"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/nri-postgresql/src/collection"
	"github.com/newrelic/nri-postgresql/src/connection"
	yaml "gopkg.in/yaml.v2"
)

const (
	versionQuery = `SHOW server_version`
)

// PopulateMetrics collects metrics for each type
func PopulateMetrics(
	ci connection.Info,
	databaseList collection.DatabaseList,
	instance *integration.Entity,
	i *integration.Integration,
	collectPgBouncer, collectDbLocks, collectBloat bool,
	customMetricsQuery string) {

	con, err := ci.NewConnection(ci.DatabaseName())
	if err != nil {
		log.Error("Metrics collection failed: error creating connection to PostgreSQL: %s", err.Error())
		return
	}
	defer con.Close()

	version, err := collectVersion(con)
	if err != nil {
		log.Error("Metrics collection failed: error collecting version number: %s", err.Error())
		return
	}

	PopulateInstanceMetrics(instance, version, con)
	PopulateDatabaseMetrics(databaseList, version, i, con, ci)
	if collectDbLocks {
		PopulateDatabaseLockMetrics(databaseList, version, i, con, ci)
	}
	PopulateTableMetrics(databaseList, version, i, ci, collectBloat)
	PopulateIndexMetrics(databaseList, i, ci)
	if customMetricsQuery != "" {
		PopulateCustomMetrics(customMetricsQuery, i, con, ci, instance)
	}

	if collectPgBouncer {
		con, err = ci.NewConnection("pgbouncer")
		if err != nil {
			log.Error("Error creating connection to pgbouncer database: %s", err)
		} else {
			defer con.Close()
			PopulatePgBouncerMetrics(i, con, ci)
		}
	}
}

// PopulateCustomMetricsFromFile collects metrics defined by a custom config file
func PopulateCustomMetricsFromFile(ci connection.Info, configFile string, psqlIntegration *integration.Integration) {
	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Error("Failed to read custom config file: %s", err)
		return
	}

	var customYAML customMetricsYAML
	err = yaml.Unmarshal(contents, &customYAML)
	if err != nil {
		log.Error("Failed to unmarshal custom config file: %s", err)
		return
	}

	// Semaphore to run 10 custom queries concurrently
	sem := make(chan struct{}, 10)
	wg := sync.WaitGroup{}
	for _, config := range customYAML.Queries {
		sem <- struct{}{}
		wg.Add(1)
		go func(cfg customMetricsConfig) {
			defer wg.Done()
			defer func() {
				<-sem
			}()

			CollectCustomConfig(ci, cfg, psqlIntegration)
		}(config)
	}
	wg.Wait()

}

// CollectCustomConfig collects metrics defined by a custom config
func CollectCustomConfig(ci connection.Info, cfg customMetricsConfig, pgIntegration *integration.Integration) {
	dbName := func() string {
		if cfg.Database == "" {
			return ci.DatabaseName()
		}
		return cfg.Database
	}()

	con, err := ci.NewConnection(dbName)
	if err != nil {
		log.Error("Custom query collection failed: error creating connection to PostgreSQL: %s", err.Error())
		return
	}
	defer con.Close()

	rows, err := con.Queryx(cfg.Query)
	if err != nil {
		log.Error("Could not execute database query: %s", err.Error())
		return
	}
	defer func() {
		_ = rows.Close()
	}()

	host, port := ci.HostPort()
	hostIDAttribute := integration.NewIDAttribute("host", host)
	portIDAttribute := integration.NewIDAttribute("port", port)
	databaseEntity, err := pgIntegration.Entity(dbName, "pg-database", hostIDAttribute, portIDAttribute)
	if err != nil {
		log.Error("Failed to create custom database entity: %s", err)
	}

	sampleName := func() string {
		if cfg.SampleName == "" {
			return "PostgresqlCustomSample"
		}
		return cfg.SampleName
	}()

	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			log.Error("Failed to scan custom query row: %s", err)
			return
		}

		ms := databaseEntity.NewMetricSet(sampleName, metric.Attribute{
			Key: "database", Value: dbName,
		})

		for k, v := range row {
			sanitized := sanitizeValue(v)
			metricType := func() metric.SourceType {
				t, ok := cfg.MetricTypes[k]
				if !ok {
					return inferMetricType(sanitized)
				}
				return metric.SourceType(t)
			}()

			err := ms.SetMetric(k, sanitized, metricType)
			if err != nil {
				log.Warn("Failed to set metric: %s", err)
			}
		}
	}
}

func sanitizeValue(val interface{}) interface{} {
	switch v := val.(type) {
	case string, float32, int, int32, int64:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func inferMetricType(val interface{}) metric.SourceType {
	switch val.(type) {
	case string:
		return metric.ATTRIBUTE
	case float32, float64, int, int32, int64:
		return metric.GAUGE
	default:
		return metric.ATTRIBUTE
	}
}

type metricType metric.SourceType

func (m *metricType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	st, err := metric.SourceTypeForName(raw)
	if err != nil {
		return err
	}

	*m = metricType(st)
	return nil
}

type customMetricsYAML struct {
	Queries []customMetricsConfig
}

type customMetricsConfig struct {
	Query       string                `yaml:"query"`
	Database    string                `yaml:"database"`
	MetricTypes map[string]metricType `yaml:"metric_types"`
	SampleName  string                `yaml:"sample_name"`
}

type serverVersionRow struct {
	Version string `db:"server_version"`
}

func collectVersion(connection *connection.PGSQLConnection) (*semver.Version, error) {
	var versionRows []*serverVersionRow
	if err := connection.Query(&versionRows, versionQuery); err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`[0-9]+\.[0-9]+(\.[0-9])?`)
	version := re.FindString(versionRows[0].Version)

	// special cases for ubuntu/debian parsing
	//version := versionRows[0].Version
	//if strings.Contains(version, "Ubuntu") {
	//return parseSpecialVersion(version, strings.Index(version, " (Ubuntu"))
	//} else if strings.Contains(version, "Debian") {
	//return parseSpecialVersion(version, strings.Index(version, " (Debian"))
	//}

	v, err := semver.ParseTolerant(version)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

//func parseSpecialVersion(version string, specialIndex int) (*semver.Version, error) {
//partialVersion := version[:specialIndex]

//v, err := semver.ParseTolerant(partialVersion)
//if err != nil {
//return nil, err
//}

//return &v, nil
//}

// PopulateInstanceMetrics populates the metrics for an instance
func PopulateInstanceMetrics(instanceEntity *integration.Entity, version *semver.Version, connection *connection.PGSQLConnection) {
	metricSet := instanceEntity.NewMetricSet("PostgresqlInstanceSample",
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

		// Nothing was returned
		if vp.Len() == 0 {
			log.Debug("No data returned from instance query '%s'", queryDef.GetQuery())
			continue
		}

		vpInterface := vp.Index(0).Interface()
		err := metricSet.MarshalMetrics(vpInterface)
		if err != nil {
			log.Error("Could not parse metrics from instance query result: %s", err.Error())
		}
	}
}

// PopulateDatabaseMetrics populates the metrics for a database
func PopulateDatabaseMetrics(databases collection.DatabaseList, version *semver.Version, pgIntegration *integration.Integration, connection *connection.PGSQLConnection, ci connection.Info) {
	databaseDefinitions := generateDatabaseDefinitions(databases, version)
	processDatabaseDefinitions(databaseDefinitions, pgIntegration, connection, ci)
}

// PopulateDatabaseLockMetrics populates the lock metrics for a database
func PopulateDatabaseLockMetrics(databases collection.DatabaseList, version *semver.Version, pgIntegration *integration.Integration, connection *connection.PGSQLConnection, ci connection.Info) {
	if !connection.HaveExtensionInSchema("tablefunc", "public") {
		log.Warn("Crosstab function not available; database lock metric gathering not possible.")
		log.Warn("To enable database lock metrics, enable the 'tablefunc' extension on the public")
		log.Warn("schema of your database. You can do so by:")
		log.Warn("  1. Installing the postgresql contribs package for your OS; and")
		log.Warn("  2. Run the query 'CREATE EXTENSION tablefunc;' against your database's public schema")
		return
	}

	lockDefinitions := generateLockDefinitions(databases, version)

	processDatabaseDefinitions(lockDefinitions, pgIntegration, connection, ci)
}

func processDatabaseDefinitions(definitions []*QueryDefinition, pgIntegration *integration.Integration, connection *connection.PGSQLConnection, ci connection.Info) {
	for _, queryDef := range definitions {
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

			host, port := ci.HostPort()
			hostIDAttribute := integration.NewIDAttribute("host", host)
			portIDAttribute := integration.NewIDAttribute("port", port)
			databaseEntity, err := pgIntegration.Entity(name, "pg-database", hostIDAttribute, portIDAttribute)
			if err != nil {
				log.Error("Failed to get database entity for name %s: %s", name, err.Error())
			}
			metricSet := databaseEntity.NewMetricSet("PostgresqlDatabaseSample",
				metric.Attribute{Key: "displayName", Value: databaseEntity.Metadata.Name},
				metric.Attribute{Key: "entityName", Value: "database:" + databaseEntity.Metadata.Name},
			)

			if err := metricSet.MarshalMetrics(db); err != nil {
				log.Error("Failed to database entity with metrics: %s", err.Error())
			}

		}
	}
}

// PopulateTableMetrics populates the metrics for a table
func PopulateTableMetrics(databases collection.DatabaseList, version *semver.Version, pgIntegration *integration.Integration, ci connection.Info, collectBloat bool) {
	for database, schemaList := range databases {
		if len(schemaList) == 0 {
			return
		}

		// Create a new connection to the database
		con, err := ci.NewConnection(database)
		defer con.Close()
		if err != nil {
			log.Error("Failed to connect to database %s: %s", database, err.Error())
			continue
		}

		populateTableMetricsForDatabase(schemaList, version, con, pgIntegration, ci, collectBloat)
	}
}

func populateTableMetricsForDatabase(schemaList collection.SchemaList, version *semver.Version, con *connection.PGSQLConnection, pgIntegration *integration.Integration, ci connection.Info, collectBloat bool) {

	tableDefinitions := generateTableDefinitions(schemaList, version, collectBloat)

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

			host, port := ci.HostPort()
			hostIDAttribute := integration.NewIDAttribute("host", host)
			portIDAttribute := integration.NewIDAttribute("port", port)
			databaseIDAttribute := integration.NewIDAttribute("pg-database", dbName)
			schemaIDAttribute := integration.NewIDAttribute("pg-schema", schemaName)
			tableEntity, err := pgIntegration.Entity(tableName, "pg-table", hostIDAttribute, portIDAttribute, databaseIDAttribute, schemaIDAttribute)
			if err != nil {
				log.Error("Failed to get table entity for table %s: %s", tableName, err.Error())
			}
			metricSet := tableEntity.NewMetricSet("PostgresqlTableSample",
				metric.Attribute{Key: "displayName", Value: tableEntity.Metadata.Name},
				metric.Attribute{Key: "entityName", Value: "table:" + tableEntity.Metadata.Name},
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
func PopulateIndexMetrics(databases collection.DatabaseList, pgIntegration *integration.Integration, ci connection.Info) {
	for database, schemaList := range databases {
		con, err := ci.NewConnection(database)
		if err != nil {
			log.Error("Failed to create new connection to database %s: %s", database, err.Error())
			continue
		}
		defer con.Close()
		populateIndexMetricsForDatabase(schemaList, con, pgIntegration, ci)
	}
}

func populateIndexMetricsForDatabase(schemaList collection.SchemaList, con *connection.PGSQLConnection, pgIntegration *integration.Integration, ci connection.Info) {
	indexDefinitions := generateIndexDefinitions(schemaList)

	for _, definition := range indexDefinitions {

		// collect into model
		dataModels := definition.GetDataModels()
		if err := con.Query(dataModels, definition.GetQuery()); err != nil {
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

			host, port := ci.HostPort()
			hostIDAttribute := integration.NewIDAttribute("host", host)
			portIDAttribute := integration.NewIDAttribute("port", port)
			databaseIDAttribute := integration.NewIDAttribute("pg-database", dbName)
			schemaIDAttribute := integration.NewIDAttribute("pg-schema", schemaName)
			tableIDAttribute := integration.NewIDAttribute("pg-table", tableName)
			indexEntity, err := pgIntegration.Entity(indexName, "pg-index", hostIDAttribute, portIDAttribute, databaseIDAttribute, schemaIDAttribute, tableIDAttribute)
			if err != nil {
				log.Error("Failed to get table entity for index %s: %s", indexName, err.Error())
			}
			metricSet := indexEntity.NewMetricSet("PostgresqlIndexSample",
				metric.Attribute{Key: "displayName", Value: indexEntity.Metadata.Name},
				metric.Attribute{Key: "entityName", Value: "index:" + indexEntity.Metadata.Name},
				metric.Attribute{Key: "database", Value: dbName},
				metric.Attribute{Key: "schema", Value: schemaName},
				metric.Attribute{Key: "table", Value: tableName},
			)

			if err := metricSet.MarshalMetrics(row); err != nil {
				log.Error("Failed to populate index entity with metrics: %s", err.Error())
			}

		}

	}
}

// PopulatePgBouncerMetrics populates pgbouncer metrics
func PopulatePgBouncerMetrics(pgIntegration *integration.Integration, con *connection.PGSQLConnection, ci connection.Info) {
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

			host, port := ci.HostPort()
			hostIDAttribute := integration.NewIDAttribute("host", host)
			portIDAttribute := integration.NewIDAttribute("port", port)
			pgEntity, err := pgIntegration.Entity(name, "pgbouncer", hostIDAttribute, portIDAttribute)
			if err != nil {
				log.Error("Failed to get database entity for name %s: %s", name, err.Error())
			}
			metricSet := pgEntity.NewMetricSet("PgBouncerSample",
				metric.Attribute{Key: "displayName", Value: name},
				metric.Attribute{Key: "entityName", Value: "pgbouncer:" + name},
				metric.Attribute{Key: "host", Value: host},
			)

			if err := metricSet.MarshalMetrics(db); err != nil {
				log.Error("Failed to populate pgbouncer entity with metrics: %s", err.Error())
			}

		}
	}
}

// PopulateCustomMetrics collects metrics from a custom query
func PopulateCustomMetrics(customMetricsQuery string, pgIntegration *integration.Integration, con *connection.PGSQLConnection, ci connection.Info, instance *integration.Entity) {

	rows, err := con.Queryx(customMetricsQuery)
	if err != nil {
		log.Error("Could not execute database query: %s", err.Error())
		return
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		row := make(map[string]interface{})
		err := rows.MapScan(row)
		if err != nil {
			log.Error("Failed to scan custom query row: %s", err)
			return
		}

		nameInterface, ok := row["metric_name"]
		if !ok {
			log.Error("Missing required column 'metric_name' in custom query")
			return
		}
		name, ok := nameInterface.(string)
		if !ok {
			log.Error("Non-string type %T for custom query 'metric_name' column", nameInterface)
			continue
		}

		metricTypeInterface, ok := row["metric_type"]
		if !ok {
			log.Error("Missing required column 'metric_type' in custom query")
			return
		}
		metricTypeString, ok := metricTypeInterface.(string)
		if !ok {
			log.Error("Non-string type %T for custom query 'metric_type' column", metricTypeInterface)
			continue
		}
		metricType, err := metric.SourceTypeForName(metricTypeString)
		if err != nil {
			log.Error("Invalid metric type %s: %s", metricTypeString, err)
			continue
		}

		value, ok := row["metric_value"]
		if !ok {
			log.Error("Missing required column 'metric_type' in custom query")
			return
		}

		attributes := []metric.Attribute{
			{Key: "displayName", Value: instance.Metadata.Name},
			{Key: "entityName", Value: instance.Metadata.Namespace + ":" + instance.Metadata.Name},
		}
		for k, v := range row {
			if k == "metric_name" || k == "metric_type" || k == "metric_value" {
				continue
			}

			var valString string
			switch v := v.(type) {
			case []byte:
				valString = string(v)
			case string:
				valString = v
			default:
				valString = fmt.Sprint(v)
			}

			attributes = append(attributes, metric.Attribute{Key: k, Value: valString})
		}

		ms := instance.NewMetricSet("PgCustomQuerySample", attributes...)
		err = ms.SetMetric(name, value, metricType)
		if err != nil {
			log.Error("Failed to set metric: %s", err)
			continue
		}
	}
}
