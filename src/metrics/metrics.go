package metrics

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"sync"

	"github.com/blang/semver/v4"
	"github.com/newrelic/infra-integrations-sdk/v3/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/infra-integrations-sdk/v3/log"
	"github.com/newrelic/nri-postgresql/src/collection"
	"github.com/newrelic/nri-postgresql/src/connection"
	yaml "gopkg.in/yaml.v3"
)

const (
	versionQuery          = `SHOW server_version`
	pgbouncerVersionQuery = `SHOW VERSION`
)

// DatabaseType represents the type of database we're connected to
type DatabaseType int

const (
	DatabaseTypePostgreSQL DatabaseType = iota
	DatabaseTypePgBouncerAdmin
	DatabaseTypeUnknown
)

// ConnectionInfo contains information about the connected database
type ConnectionInfo struct {
	Type              DatabaseType
	PostgreSQLVersion *semver.Version
	PgBouncerVersion  string
}

// PopulateMetrics collects metrics for each type
// It automatically detects whether we're connected to PostgreSQL or PgBouncer admin console
func PopulateMetrics(
	ci connection.Info,
	databaseList collection.DatabaseList,
	instance *integration.Entity,
	i *integration.Integration,
	collectPgBouncer, collectDbLocks, collectBloat bool,
	customMetricsQuery string) {

	// Create connection to the configured database
	con, err := ci.NewConnection(ci.DatabaseName())
	if err != nil {
		log.Error("Metrics collection failed: error creating connection: %s", err.Error())
		return
	}
	defer con.Close()

	// Detect what type of database we're connected to by trying queries
	connInfo, err := DetectDatabaseType(con)
	if err != nil {
		log.Error("Metrics collection failed: unable to determine database type: %s", err.Error())
		return
	}

	// Handle based on detected database type
	switch connInfo.Type {
	case DatabaseTypePgBouncerAdmin:
		// Connected to PgBouncer admin console - only collect PgBouncer metrics
		log.Info("Detected PgBouncer admin console connection")
		log.Info("PgBouncer version: %s", connInfo.PgBouncerVersion)

		// Add version to instance entity
		addPgBouncerVersionToInstance(instance, connInfo.PgBouncerVersion)

		// Collect PgBouncer metrics
		PopulatePgBouncerMetrics(i, con, ci)

	case DatabaseTypePostgreSQL:
		// Connected to PostgreSQL (either direct or through PgBouncer proxy)
		// Collect all standard PostgreSQL metrics
		log.Info("Detected PostgreSQL connection, version: %s", connInfo.PostgreSQLVersion)

		PopulateInstanceMetrics(instance, connInfo.PostgreSQLVersion, con)
		PopulateDatabaseMetrics(databaseList, connInfo.PostgreSQLVersion, i, con, ci)
		if collectDbLocks {
			PopulateDatabaseLockMetrics(databaseList, connInfo.PostgreSQLVersion, i, con, ci)
		}
		PopulateTableMetrics(databaseList, connInfo.PostgreSQLVersion, i, ci, collectBloat)
		PopulateIndexMetrics(databaseList, i, ci)
		if customMetricsQuery != "" {
			PopulateCustomMetrics(customMetricsQuery, i, con, ci, instance)
		}

		// If PgBouncer flag is set, also collect PgBouncer pool metrics
		// by connecting to the "pgbouncer" virtual database
		if collectPgBouncer {
			pbCon, pbErr := ci.NewConnection("pgbouncer")
			if pbErr != nil {
				log.Error("Error creating connection to pgbouncer database: %s", pbErr)
			} else {
				defer pbCon.Close()

				// Try to get PgBouncer version and add to metrics
				pbVersion, pbVerErr := tryPgBouncerVersion(pbCon)
				if pbVerErr != nil {
					log.Warn("Unable to collect PgBouncer version: %s", pbVerErr.Error())
				} else {
					log.Info("PgBouncer version: %s", pbVersion)
					addPgBouncerVersionToInstance(instance, pbVersion)
				}

				PopulatePgBouncerMetrics(i, pbCon, ci)
			}
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

		ms := databaseEntity.NewMetricSet(sampleName, attribute.Attribute{
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
	case string, float32, float64, int, int32, int64:
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

type pgbouncerVersionRow struct {
	Version string `db:"version"`
}

// addPgBouncerVersionToInstance adds PgBouncer version metrics to the instance entity
func addPgBouncerVersionToInstance(instance *integration.Entity, version string) {
	metricSet := instance.NewMetricSet("PostgresqlInstanceSample",
		attribute.Attribute{Key: "displayName", Value: instance.Metadata.Name},
		attribute.Attribute{Key: "entityName", Value: instance.Metadata.Namespace + ":" + instance.Metadata.Name},
		attribute.Attribute{Key: "pgbouncerVersion", Value: version},
	)
	_ = metricSet.SetMetric("pgbouncer.version", version, metric.ATTRIBUTE)
}

// DetectDatabaseType determines what type of database we're connected to
// by attempting queries and checking which ones succeed
func DetectDatabaseType(connection *connection.PGSQLConnection) (*ConnectionInfo, error) {
	connInfo := &ConnectionInfo{Type: DatabaseTypeUnknown}

	// First, try PostgreSQL version query
	pgVersion, pgErr := tryPostgreSQLVersion(connection)
	if pgErr == nil {
		// Successfully connected to PostgreSQL (either direct or through PgBouncer proxy)
		connInfo.Type = DatabaseTypePostgreSQL
		connInfo.PostgreSQLVersion = pgVersion
		return connInfo, nil
	}

	// PostgreSQL query failed, try PgBouncer admin console query
	pbVersion, pbErr := tryPgBouncerVersion(connection)
	if pbErr == nil {
		// Successfully connected to PgBouncer admin console
		connInfo.Type = DatabaseTypePgBouncerAdmin
		connInfo.PgBouncerVersion = pbVersion
		return connInfo, nil
	}

	// Neither worked - return both errors
	return nil, fmt.Errorf("unable to determine database type - PostgreSQL error: %v, PgBouncer error: %v", pgErr, pbErr)
}

// tryPostgreSQLVersion attempts to get PostgreSQL version
func tryPostgreSQLVersion(connection *connection.PGSQLConnection) (*semver.Version, error) {
	var versionRows []*serverVersionRow
	if err := connection.Query(&versionRows, versionQuery); err != nil {
		return nil, err
	}

	if len(versionRows) == 0 {
		return nil, fmt.Errorf("no version information returned")
	}

	re := regexp.MustCompile(`[0-9]+\.[0-9]+(\.[0-9])?`)
	version := re.FindString(versionRows[0].Version)

	v, err := semver.ParseTolerant(version)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// tryPgBouncerVersion attempts to get PgBouncer version
func tryPgBouncerVersion(connection *connection.PGSQLConnection) (string, error) {
	var versionRows []*pgbouncerVersionRow
	if err := connection.Query(&versionRows, pgbouncerVersionQuery); err != nil {
		return "", err
	}

	if len(versionRows) == 0 {
		return "", fmt.Errorf("no version information returned from PgBouncer")
	}

	return versionRows[0].Version, nil
}

// CollectVersion is kept for backward compatibility
// It only works for PostgreSQL connections
func CollectVersion(connection *connection.PGSQLConnection) (*semver.Version, error) {
	return tryPostgreSQLVersion(connection)
}


// func parseSpecialVersion(version string, specialIndex int) (*semver.Version, error) {
// partialVersion := version[:specialIndex]

//v, err := semver.ParseTolerant(partialVersion)
//if err != nil {
//return nil, err
//}

//return &v, nil
//}

// PopulateInstanceMetrics populates the metrics for an instance
func PopulateInstanceMetrics(instanceEntity *integration.Entity, version *semver.Version, connection *connection.PGSQLConnection) {
	metricSet := instanceEntity.NewMetricSet("PostgresqlInstanceSample",
		attribute.Attribute{Key: "displayName", Value: instanceEntity.Metadata.Name},
		attribute.Attribute{Key: "entityName", Value: instanceEntity.Metadata.Namespace + ":" + instanceEntity.Metadata.Name},
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

	lockDefinitions := generateLockDefinitions(databases)

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
				attribute.Attribute{Key: "displayName", Value: databaseEntity.Metadata.Name},
				attribute.Attribute{Key: "entityName", Value: "database:" + databaseEntity.Metadata.Name},
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
		if err != nil {
			log.Error("Failed to connect to database %s: %s", database, err.Error())
			continue
		}
		defer con.Close()
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
				attribute.Attribute{Key: "displayName", Value: tableEntity.Metadata.Name},
				attribute.Attribute{Key: "entityName", Value: "table:" + tableEntity.Metadata.Name},
				attribute.Attribute{Key: "database", Value: dbName},
				attribute.Attribute{Key: "schema", Value: schemaName},
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
				attribute.Attribute{Key: "displayName", Value: indexEntity.Metadata.Name},
				attribute.Attribute{Key: "entityName", Value: "index:" + indexEntity.Metadata.Name},
				attribute.Attribute{Key: "database", Value: dbName},
				attribute.Attribute{Key: "schema", Value: schemaName},
				attribute.Attribute{Key: "table", Value: tableName},
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
		// Use QueryUnsafe to support different PgBouncer versions with varying column sets
		if err := con.QueryUnsafe(dataModels, definition.GetQuery()); err != nil {
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
				attribute.Attribute{Key: "displayName", Value: name},
				attribute.Attribute{Key: "entityName", Value: "pgbouncer:" + name},
				attribute.Attribute{Key: "host", Value: host},
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

		attributes := []attribute.Attribute{
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

			attributes = append(attributes, attribute.Attribute{Key: k, Value: valString})
		}

		ms := instance.NewMetricSet("PgCustomQuerySample", attributes...)
		err = ms.SetMetric(name, value, metricType)
		if err != nil {
			log.Error("Failed to set metric: %s", err)
			continue
		}
	}
}
