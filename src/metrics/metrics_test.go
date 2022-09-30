package metrics

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/newrelic/infra-integrations-sdk/integration"
	"github.com/newrelic/nri-postgresql/src/collection"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestPopulateInstanceMetrics(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")
	testEntity, _ := testIntegration.Entity("testInstance", "instance")

	version := semver.MustParse("9.0.0")

	testConnection, mock := connection.CreateMockSQL(t)
	instanceRows := sqlmock.NewRows([]string{
		"scheduled_checkpoints_performed",
		"requested_checkpoints_performed",
		"buffers_written_during_checkpoint",
		"buffers_written_by_background_writer",
		"background_writer_stops",
		"buffers_written_by_backend",
		"buffers_allocated",
	}).AddRow(1, 2, 3, 4, 5, 6, 7)

	mock.ExpectQuery(".*scheduled_checkpoints_performed.*").
		WillReturnRows(instanceRows)

	PopulateInstanceMetrics(testEntity, &version, testConnection)

	expected := map[string]interface{}{
		"bgwriter.checkpointsScheduledPerSecond":             float64(0),
		"bgwriter.checkpointsRequestedPerSecond":             float64(0),
		"bgwriter.buffersWrittenForCheckpointsPerSecond":     float64(0),
		"bgwriter.buffersWrittenByBackgroundWriterPerSecond": float64(0),
		"bgwriter.backgroundWriterStopsPerSecond":            float64(0),
		"bgwriter.buffersWrittenByBackendPerSecond":          float64(0),
		"bgwriter.buffersAllocatedPerSecond":                 float64(0),
		"displayName":                                        "testInstance",
		"entityName":                                         "instance:testInstance",
		"event_type":                                         "PostgresqlInstanceSample",
	}

	assert.Equal(t, expected, testEntity.Metrics[0].Metrics)
}

func TestPopulateInstanceMetrics_NoRows(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")
	testEntity, _ := testIntegration.Entity("testInstance", "instance")

	version := semver.MustParse("9.0.0")

	testConnection, mock := connection.CreateMockSQL(t)
	instanceRows := sqlmock.NewRows([]string{
		"scheduled_checkpoints_performed",
		"requested_checkpoints_performed",
		"buffers_written_during_checkpoint",
		"buffers_written_by_background_writer",
		"background_writer_stops",
		"buffers_written_by_backend",
		"buffers_allocated",
	})

	mock.ExpectQuery(".*scheduled_checkpoints_performed.*").
		WillReturnRows(instanceRows)

	PopulateInstanceMetrics(testEntity, &version, testConnection)

	expected := map[string]interface{}{
		"displayName": "testInstance",
		"entityName":  "instance:testInstance",
		"event_type":  "PostgresqlInstanceSample",
	}

	assert.Equal(t, expected, testEntity.Metrics[0].Metrics)
}

func TestPopulateDatabaseMetrics(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")

	version := semver.MustParse("9.0.0")
	dbList := collection.DatabaseList{"test1": {}}

	testConnection, mock := connection.CreateMockSQL(t)
	databaseRows := sqlmock.NewRows([]string{
		"database",
		"active_connections",
		"transactions_committed",
		"transactions_rolled_back",
		"block_reads",
		"buffer_hits",
		"rows_returned",
		"rows_fetched",
		"rows_inserted",
		"rows_updated",
		"rows_deleted",
	}).AddRow("testDB", 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	mock.ExpectQuery(".*UNDER91.*").
		WillReturnRows(databaseRows)

	ci := &connection.MockInfo{}
	PopulateDatabaseMetrics(dbList, &version, testIntegration, testConnection, ci)

	expected := map[string]interface{}{

		"db.bufferHitsPerSecond":   float64(0),
		"db.commitsPerSecond":      float64(0),
		"db.connections":           float64(1),
		"db.readsPerSecond":        float64(0),
		"db.rollbacksPerSecond":    float64(0),
		"db.rowsDeletedPerSecond":  float64(0),
		"db.rowsFetchedPerSecond":  float64(0),
		"db.rowsInsertedPerSecond": float64(0),
		"db.rowsReturnedPerSecond": float64(0),
		"db.rowsUpdatedPerSecond":  float64(0),
		"displayName":              "testDB",
		"entityName":               "database:testDB",
		"event_type":               "PostgresqlDatabaseSample",
	}

	dbEntity, err := testIntegration.Entity("testDB", "pg-database", integration.NewIDAttribute("host", "testhost"), integration.NewIDAttribute("port", "1234"))
	assert.Nil(t, err)
	assert.Equal(t, expected, dbEntity.Metrics[0].Metrics)
}

func TestPopulateDatabaseLockMetrics_WithTablefuncExtension(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")

	version := semver.MustParse("9.0.0")
	dbList := collection.DatabaseList{"test1": {}}

	testConnection, mock := connection.CreateMockSQL(t)

	extensionRows := sqlmock.NewRows([]string{
		"schema",
		"extension",
	}).AddRow("public", "tablefunc")
	mock.ExpectQuery(".*EXTENSIONS_LIST.*").WillReturnRows(extensionRows)

	lockRows := sqlmock.NewRows([]string{
		"database",
		"access_exclusive_lock",
		"access_share_lock",
		"exclusive_lock",
		"row_exclusive_lock",
		"row_share_lock",
		"share_lock",
		"share_row_exclusive_lock",
		"share_update_exclusive_lock",
	}).AddRow("testDB", 1, 2, 3, 4, 5, 6, 7, 8)
	mock.ExpectQuery(".*LOCKS_DEFINITION.*").WillReturnRows(lockRows)

	ci := &connection.MockInfo{}
	PopulateDatabaseLockMetrics(dbList, &version, testIntegration, testConnection, ci)

	expected := map[string]interface{}{
		"db.locks.accessExclusiveLock":      float64(1),
		"db.locks.accessShareLock":          float64(2),
		"db.locks.exclusiveLock":            float64(3),
		"db.locks.rowExclusiveLock":         float64(4),
		"db.locks.rowShareLock":             float64(5),
		"db.locks.shareLock":                float64(6),
		"db.locks.shareRowExclusiveLock":    float64(7),
		"db.locks.shareUpdateExclusiveLock": float64(8),
		"displayName":                       "testDB",
		"entityName":                        "database:testDB",
		"event_type":                        "PostgresqlDatabaseSample",
	}

	dbEntity, err := testIntegration.Entity("testDB", "pg-database", integration.NewIDAttribute("host", "testhost"), integration.NewIDAttribute("port", "1234"))

	assert.Nil(t, err)
	assert.Equal(t, expected, dbEntity.Metrics[0].Metrics)
}

func TestPopulateDatabaseLockMetrics_WithoutTablefuncExtension(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")

	version := semver.MustParse("9.0.0")
	dbList := collection.DatabaseList{"test1": {}}

	testConnection, mock := connection.CreateMockSQL(t)
	extensionRows := sqlmock.NewRows([]string{"schema", "extension"})
	mock.ExpectQuery(".*EXTENSIONS_LIST.*").WillReturnRows(extensionRows)

	ci := &connection.MockInfo{}
	PopulateDatabaseLockMetrics(dbList, &version, testIntegration, testConnection, ci)
	dbEntity, err := testIntegration.Entity("testDB", "pg-database", integration.NewIDAttribute("host", "testhost"), integration.NewIDAttribute("port", "1234"))

	assert.Nil(t, err)
	assert.Empty(t, dbEntity.Metrics)
}

func Test_populateTableMetricsForDatabase(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")

	dbList := collection.DatabaseList{
		"db1": collection.SchemaList{
			"schema1": collection.TableList{
				"table1": []string{},
			},
		},
	}

	testConnection, mock := connection.CreateMockSQL(t)
	tableRows := sqlmock.NewRows([]string{
		"database",
		"schema_name",
		"table_name",
		"pg_total_relation_size",
		"pg_indexes_size",
		"idx_blks_read",
		"idx_blks_hit",
		"toast_blks_read",
		"toast_blks_hit",
		"last_vacuum",
		"last_autovacuum",
		"last_analyze",
		"last_autoanalyze",
		"seq_scan",
		"seq_tup_read",
		"idx_scan",
		"idx_tup_fetch",
		"n_tup_ins",
		"n_tup_upd",
		"n_tup_del",
		"n_live_tup",
		"n_dead_tup",
	}).AddRow("db1", "schema1", "table1", 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19)

	bloatRows := sqlmock.NewRows([]string{
		"database",
		"schema_name",
		"table_name",
		"bloat_size",
		"real_size",
		"bloat_ratio",
	}).AddRow("db1", "schema1", "table1", 1.0, 2.0, 0.3)

	mock.ExpectQuery(".*BLOATQUERY.*").
		WillReturnRows(bloatRows)
	mock.ExpectQuery(".*TABLEQUERY.*").
		WillReturnRows(tableRows)

	ci := &connection.MockInfo{}
	version := semver.MustParse("12.0.0")
	populateTableMetricsForDatabase(dbList["db1"], &version, testConnection, testIntegration, ci, true)

	expectedBase := map[string]interface{}{
		"table.totalSizeInBytes":                   float64(1),
		"table.indexSizeInBytes":                   float64(2),
		"table.indexBlocksReadPerSecond":           float64(0),
		"table.indexBlocksHitPerSecond":            float64(0),
		"table.indexToastBlocksReadPerSecond":      float64(0),
		"table.indexToastBlocksHitPerSecond":       float64(0),
		"table.lastVacuum":                         float64(7),
		"table.lastAutoVacuum":                     float64(8),
		"table.lastAnalyze":                        float64(9),
		"table.lastAutoAnalyze":                    float64(10),
		"table.sequentialScansPerSecond":           float64(0),
		"table.sequentialScanRowsFetchedPerSecond": float64(0),
		"table.indexScansPerSecond":                float64(0),
		"table.indexScanRowsFetchedPerSecond":      float64(0),
		"table.rowsInsertedPerSecond":              float64(0),
		"table.rowsUpdatedPerSecond":               float64(0),
		"table.rowsDeletedPerSecond":               float64(0),
		"table.liveRows":                           float64(18),
		"table.deadRows":                           float64(19),
		"database":                                 "db1",
		"schema":                                   "schema1",
		"displayName":                              "table1",
		"entityName":                               "table:table1",
		"event_type":                               "PostgresqlTableSample",
	}

	expectedBloat := map[string]interface{}{
		"table.bloatRatio":       float64(0.3),
		"table.bloatSizeInBytes": float64(1.0),
		"table.dataSizeInBytes":  float64(2.0),
		"database":               "db1",
		"schema":                 "schema1",
		"displayName":            "table1",
		"entityName":             "table:table1",
		"event_type":             "PostgresqlTableSample",
	}

	id1 := integration.NewIDAttribute("pg-database", "db1")
	id2 := integration.NewIDAttribute("pg-schema", "schema1")
	id3 := integration.NewIDAttribute("host", "testhost")
	id4 := integration.NewIDAttribute("port", "1234")
	tableEntity, err := testIntegration.Entity("table1", "pg-table", id1, id2, id3, id4)
	assert.Nil(t, err)
	assert.Equal(t, expectedBloat, tableEntity.Metrics[0].Metrics)
	assert.Equal(t, expectedBase, tableEntity.Metrics[1].Metrics)
}

func Test_populateTableMetricsForDatabase_noTables(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")

	dbList := collection.DatabaseList{
		"db1": collection.SchemaList{
			"schema1": collection.TableList{},
		},
	}

	testConnection, _ := connection.CreateMockSQL(t)

	ci := &connection.MockInfo{}
	version := semver.MustParse("10.0.0")
	populateTableMetricsForDatabase(dbList["db1"], &version, testConnection, testIntegration, ci, true)

	tableEntity, err := testIntegration.Entity("table1", "table")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(tableEntity.Metrics))
}

func Test_populateIndexMetricsForDatabase(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")

	dbList := collection.DatabaseList{
		"db1": collection.SchemaList{
			"schema1": collection.TableList{
				"table1": []string{
					"index1",
					"index2",
				},
			},
		},
	}

	testConnection, mock := connection.CreateMockSQL(t)
	indexRows := sqlmock.NewRows([]string{
		"database",
		"schema_name",
		"table_name",
		"index_name",
		"index_size",
		"tuples_read",
		"tuples_fetched",
	}).AddRow("db1", "schema1", "table1", "index1", 1, 2, 3).AddRow("db1", "schema1", "table1", "index2", 1, 2, 3)

	mock.ExpectQuery(".*INDEXQUERY.*").
		WillReturnRows(indexRows)

	ci := &connection.MockInfo{}
	populateIndexMetricsForDatabase(dbList["db1"], testConnection, testIntegration, ci)

	expected := map[string]interface{}{
		"database":                   "db1",
		"schema":                     "schema1",
		"table":                      "table1",
		"displayName":                "index1",
		"entityName":                 "index:index1",
		"event_type":                 "PostgresqlIndexSample",
		"index.sizeInBytes":          float64(1),
		"index.rowsReadPerSecond":    float64(0),
		"index.rowsFetchedPerSecond": float64(0),
	}

	id1 := integration.NewIDAttribute("pg-database", "db1")
	id2 := integration.NewIDAttribute("pg-schema", "schema1")
	id3 := integration.NewIDAttribute("host", "testhost")
	id4 := integration.NewIDAttribute("port", "1234")
	id5 := integration.NewIDAttribute("pg-table", "table1")
	indexEntity, err := testIntegration.Entity("index1", "pg-index", id1, id2, id3, id4, id5)
	assert.Nil(t, err)
	assert.Equal(t, expected, indexEntity.Metrics[0].Metrics)
}

func Test_populateIndexMetricsForDatabase_noIndexes(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")

	dbList := collection.DatabaseList{
		"db1": collection.SchemaList{
			"schema1": collection.TableList{
				"table1": []string{},
			},
		},
	}

	testConnection, _ := connection.CreateMockSQL(t)

	ci := &connection.MockInfo{}
	populateIndexMetricsForDatabase(dbList["db1"], testConnection, testIntegration, ci)

	indexEntity, err := testIntegration.Entity("index1", "index")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(indexEntity.Metrics))
}

func TestPopulatePgBouncerMetrics(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")

	testConnection, mock := connection.CreateMockSQL(t)
	pgbouncerStatsRows := sqlmock.NewRows([]string{
		"database",
		"total_xact_count",
		"total_query_count",
		"total_received",
		"total_sent",
		"total_xact_time",
		"total_query_time",
		"total_wait_time",
		"avg_xact_count",
		"avg_xact_time",
		"avg_query_count",
		"avg_recv",
		"avg_sent",
		"avg_query_time",
		"avg_wait_time",
	}).AddRow("testDB", 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14)

	pgbouncerPoolsRows := sqlmock.NewRows([]string{
		"database",
		"user",
		"cl_active",
		"cl_waiting",
		"sv_active",
		"sv_idle",
		"sv_used",
		"sv_tested",
		"sv_login",
		"maxwait",
		"maxwait_us",
		"pool_mode",
	}).AddRow("testDB", "testUser", 1, 2, 3, 4, 5, 6, 7, 8, 9, "testMode")

	mock.ExpectQuery("SHOW STATS;").
		WillReturnRows(pgbouncerStatsRows)
	mock.ExpectQuery("SHOW POOLS;").
		WillReturnRows(pgbouncerPoolsRows)

	ci := &connection.MockInfo{}
	PopulatePgBouncerMetrics(testIntegration, testConnection, ci)

	expectedStats := map[string]interface{}{
		"pgbouncer.stats.transactionsPerSecond":                           float64(0),
		"pgbouncer.stats.queriesPerSecond":                                float64(0),
		"pgbouncer.stats.bytesInPerSecond":                                float64(0),
		"pgbouncer.stats.bytesOutPerSecond":                               float64(0),
		"pgbouncer.stats.totalTransactionDurationInMillisecondsPerSecond": float64(0),
		"pgbouncer.stats.totalQueryDurationInMillisecondsPerSecond":       float64(0),
		"pgbouncer.stats.avgTransactionCount":                             float64(8),
		"pgbouncer.stats.avgTransactionDurationInMilliseconds":            float64(9),
		"pgbouncer.stats.avgQueryCount":                                   float64(10),
		"pgbouncer.stats.avgBytesIn":                                      float64(11),
		"pgbouncer.stats.avgBytesOut":                                     float64(12),
		"pgbouncer.stats.avgQueryDurationInMilliseconds":                  float64(13),

		"displayName": "testDB",
		"entityName":  "pgbouncer:testDB",
		"event_type":  "PgBouncerSample",
		"host":        "testhost",
	}

	expectedPool := map[string]interface{}{
		"pgbouncer.pools.clientConnectionsActive":  float64(1),
		"pgbouncer.pools.clientConnectionsWaiting": float64(2),
		"pgbouncer.pools.serverConnectionsActive":  float64(3),
		"pgbouncer.pools.serverConnectionsIdle":    float64(4),
		"pgbouncer.pools.serverConnectionsUsed":    float64(5),
		"pgbouncer.pools.serverConnectionsTested":  float64(6),
		"pgbouncer.pools.serverConnectionsLogin":   float64(7),
		"pgbouncer.pools.maxwaitInMilliseconds":    float64(8),
		"displayName":                              "testDB",
		"entityName":                               "pgbouncer:testDB",
		"event_type":                               "PgBouncerSample",
		"host":                                     "testhost",
	}

	id3 := integration.NewIDAttribute("host", "testhost")
	id4 := integration.NewIDAttribute("port", "1234")
	pbEntity, err := testIntegration.Entity("testDB", "pgbouncer", id3, id4)
	assert.Nil(t, err)
	assert.Equal(t, expectedStats, pbEntity.Metrics[0].Metrics)
	assert.Equal(t, expectedPool, pbEntity.Metrics[1].Metrics)
}

func TestPopulateMetrics(t *testing.T) {
	testIntegration, _ := integration.New("test", "test")

	dbList := collection.DatabaseList{
		"db1": collection.SchemaList{
			"schema1": collection.TableList{
				"table1": []string{
					"index1",
				},
			},
		},
	}

	ci := &connection.MockInfo{}
	testConnection, mock := connection.CreateMockSQL(t)

	versionRows := sqlmock.NewRows([]string{"server_version"}).AddRow("9.2.24")
	mock.ExpectQuery(".*server_version.*").WillReturnRows(versionRows)

	ci.On("NewConnection", tmock.Anything).Return(testConnection, nil)

	instance, _ := testIntegration.Entity("testInstance", "instance")

	PopulateMetrics(ci, dbList, instance, testIntegration, true, true, true, "")
}

func TestPopulateCustomMetricsFromFile(t *testing.T) {
	t.Parallel()

	testIntegration, _ := integration.New("test", "test")

	ci := &connection.MockInfo{}

	testConnection, mock := connection.CreateMockSQL(t)

	ci.On("NewConnection", tmock.Anything).Return(testConnection, nil)

	instanceRows := sqlmock.NewRows([]string{
		"int_metric",
		"float_metric",
		"string_metric",
	}).AddRow(25, 0.064, "test-string")

	mock.ExpectQuery(".*testTable.*").
		WillReturnRows(instanceRows)

	dir := t.TempDir()
	customQueryCfg := []byte(`---
queries:
  - query: >-
      SELECT test FROM testTable;
`)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "customQueryConfig.yaml"), customQueryCfg, 0600))

	PopulateCustomMetricsFromFile(ci, filepath.Join(dir, "customQueryConfig.yaml"), testIntegration)

	assert.Len(t, testIntegration.Entities, 1)
	assert.Len(t, testIntegration.Entities[0].Metrics, 1)
	metricSet := testIntegration.Entities[0].Metrics[0].Metrics

	assert.Equal(t, float64(25), metricSet["int_metric"])
	assert.Equal(t, float64(0.064), metricSet["float_metric"])
	assert.Equal(t, "test-string", metricSet["string_metric"])
}
