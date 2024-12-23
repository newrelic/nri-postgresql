package metrics

import (
	"github.com/blang/semver/v4"
	"github.com/newrelic/nri-postgresql/src/collection"
)

func generateDatabaseDefinitions(databases collection.DatabaseList, version *semver.Version) []*QueryDefinition {
	log.info("test")
	queryDefinitions := make([]*QueryDefinition, 0, 2)
	if len(databases) == 0 {
		return queryDefinitions
	}

	v91 := semver.MustParse("9.1.0")
	v92 := semver.MustParse("9.2.0")

	if version.LT(v91) {
		queryDefinitions = append(queryDefinitions, databaseDefinitionUnder91.insertDatabaseNames(databases))
	} else {
		queryDefinitions = append(queryDefinitions, databaseDefinitionOver91.insertDatabaseNames(databases))
	}

	if version.GE(v92) {
		queryDefinitions = append(queryDefinitions, databaseDefinitionOver92.insertDatabaseNames(databases))
	}

	return queryDefinitions
}

// databaseDefinitionUnder91 is the query used to fetch metrics from Postgres below version 9.1.
// As a special case, max_connections is obtained from the pg_settings table rather than from pg_stat_database
var databaseDefinitionUnder91 = &QueryDefinition{
	query: `SELECT -- UNDER91
		D.datname AS database,
		(SELECT setting::INTEGER FROM pg_settings WHERE  name = 'max_connections') AS max_connections,
		SD.numbackends AS active_connections,
		SD.xact_commit AS transactions_committed,
		SD.xact_rollback AS transactions_rolled_back,
		SD.blks_read AS block_reads,
		SD.blks_hit AS buffer_hits,
		SD.tup_returned AS rows_returned,
		SD.tup_fetched AS rows_fetched,
		SD.tup_inserted AS rows_inserted,
		SD.tup_updated AS rows_updated,
		SD.tup_deleted AS rows_deleted
		FROM pg_stat_database SD 
		INNER JOIN pg_database D ON D.datname = SD.datname 
		LEFT JOIN pg_tablespace TS ON TS.oid = D.dattablespace 
		WHERE D.datistemplate = FALSE 
			AND D.datname IS NOT NULL
			AND D.datname IN (%DATABASES%);`,

	dataModels: []struct {
		databaseBase
		MaxConnections         *int64 `db:"max_connections"          metric_name:"db.maxconnections"        source_type:"gauge"`
		ActiveConnections      *int64 `db:"active_connections"       metric_name:"db.connections"           source_type:"gauge"`
		TransactionsCommitted  *int64 `db:"transactions_committed"   metric_name:"db.commitsPerSecond"      source_type:"rate"`
		TransactionsRolledBack *int64 `db:"transactions_rolled_back" metric_name:"db.rollbacksPerSecond"    source_type:"rate"`
		BlockReads             *int64 `db:"block_reads"              metric_name:"db.readsPerSecond"        source_type:"rate"`
		BufferHits             *int64 `db:"buffer_hits"              metric_name:"db.bufferHitsPerSecond"   source_type:"rate"`
		RowsReturned           *int64 `db:"rows_returned"            metric_name:"db.rowsReturnedPerSecond" source_type:"rate"`
		RowsFetched            *int64 `db:"rows_fetched"             metric_name:"db.rowsFetchedPerSecond"  source_type:"rate"`
		RowsInserted           *int64 `db:"rows_inserted"            metric_name:"db.rowsInsertedPerSecond" source_type:"rate"`
		RowsUpdated            *int64 `db:"rows_updated"             metric_name:"db.rowsUpdatedPerSecond"  source_type:"rate"`
		RowsDeleted            *int64 `db:"rows_deleted"             metric_name:"db.rowsDeletedPerSecond"  source_type:"rate"`
	}{},
}

// databaseDefinitionOver91 is the query used to fetch metrics from Postgres version 9.1 and above.
// As a special case, max_connections is obtained from the pg_settings table rather than from pg_stat_database
var databaseDefinitionOver91 = &QueryDefinition{
	query: `SELECT 
		D.datname AS database,
		(SELECT setting::INTEGER FROM pg_settings WHERE  name = 'max_connections') AS max_connections,
		SD.numbackends AS active_connections,
		SD.xact_commit AS transactions_committed,
		SD.xact_rollback AS transactions_rolled_back,
		SD.blks_read AS block_reads,
		SD.blks_hit AS buffer_hits,
		SD.tup_returned AS rows_returned,
		SD.tup_fetched AS rows_fetched,
		SD.tup_inserted AS rows_inserted,
		SD.tup_updated AS rows_updated,
		SD.tup_deleted AS rows_deleted,
		DBC.confl_tablespace AS queries_canceled_due_to_dropped_tablespaces,
		DBC.confl_lock AS queries_canceled_due_to_lock_timeouts,
		DBC.confl_snapshot AS queries_canceled_due_to_old_snapshots,
		DBC.confl_bufferpin AS queries_canceled_due_to_pinned_buffers,
		DBC.confl_deadlock AS queries_canceled_due_to_deadlocks
		FROM pg_stat_database SD 
		INNER JOIN pg_database D ON D.datname = SD.datname 
		INNER JOIN pg_stat_database_conflicts DBC ON DBC.datname = D.datname 
		LEFT JOIN pg_tablespace TS ON TS.oid = D.dattablespace 
		WHERE D.datistemplate = FALSE 
			AND D.datname IS NOT NULL
			AND D.datname IN (%DATABASES%);`,

	dataModels: []struct {
		databaseBase
		MaxConnections                    *int64 `db:"max_connections"                             metric_name:"db.maxconnections"                source_type:"gauge"`
		ActiveConnections                 *int64 `db:"active_connections"                          metric_name:"db.connections"                   source_type:"gauge"`
		TransactionsCommitted             *int64 `db:"transactions_committed"                      metric_name:"db.commitsPerSecond"              source_type:"rate"`
		TransactionsRolledBack            *int64 `db:"transactions_rolled_back"                    metric_name:"db.rollbacksPerSecond"            source_type:"rate"`
		BlockReads                        *int64 `db:"block_reads"                                 metric_name:"db.readsPerSecond"                source_type:"rate"`
		BufferHits                        *int64 `db:"buffer_hits"                                 metric_name:"db.bufferHitsPerSecond"           source_type:"rate"`
		RowsReturned                      *int64 `db:"rows_returned"                               metric_name:"db.rowsReturnedPerSecond"         source_type:"rate"`
		RowsFetched                       *int64 `db:"rows_fetched"                                metric_name:"db.rowsFetchedPerSecond"          source_type:"rate"`
		RowsInserted                      *int64 `db:"rows_inserted"                               metric_name:"db.rowsInsertedPerSecond"         source_type:"rate"`
		RowsUpdated                       *int64 `db:"rows_updated"                                metric_name:"db.rowsUpdatedPerSecond"          source_type:"rate"`
		RowsDeleted                       *int64 `db:"rows_deleted"                                metric_name:"db.rowsDeletedPerSecond"          source_type:"rate"`
		CanceledQueriesDroppedTablespaces *int64 `db:"queries_canceled_due_to_dropped_tablespaces" metric_name:"db.conflicts.tablespacePerSecond" source_type:"rate"`
		CanceledQueriesLockTimeouts       *int64 `db:"queries_canceled_due_to_lock_timeouts"       metric_name:"db.conflicts.locksPerSecond"      source_type:"rate"`
		CanceledQueriesOldSnapshots       *int64 `db:"queries_canceled_due_to_old_snapshots"       metric_name:"db.conflicts.snapshotPerSecond"   source_type:"rate"`
		CanceledQueriesPinnedBuffers      *int64 `db:"queries_canceled_due_to_pinned_buffers"      metric_name:"db.conflicts.bufferpinPerSecond"  source_type:"rate"`
		CanceledQueriesDeadlocks          *int64 `db:"queries_canceled_due_to_deadlocks"           metric_name:"db.conflicts.deadlockPerSecond"   source_type:"rate"`
	}{},
}

// databaseDefinitionOver92 is the query used to fetch extra metrics from Postgres version 9.2 and above.
var databaseDefinitionOver92 = &QueryDefinition{
	query: `SELECT 
		D.datname AS database,
		SD.temp_files AS temporary_files_created,
		SD.temp_bytes AS temporary_bytes_written,
		SD.deadlocks AS deadlocks,
		cast(SD.blk_read_time AS bigint) AS time_spent_reading_data,
		cast(SD.blk_write_time AS bigint) AS time_spent_writing_data
		FROM pg_stat_database SD 
		INNER JOIN pg_database D ON D.datname = SD.datname 
		INNER JOIN pg_stat_database_conflicts DBC ON DBC.datname = D.datname 
		LEFT JOIN pg_tablespace TS ON TS.oid = D.dattablespace 
		WHERE D.datistemplate = FALSE 
			AND D.datname IS NOT NULL
			AND D.datname IN (%DATABASES%);`,

	dataModels: []struct {
		databaseBase
		TempFilesCreated   *int64 `db:"temporary_files_created" metric_name:"db.tempFilesCreatedPerSecond"        source_type:"rate"`
		TempWrittenInBytes *int64 `db:"temporary_bytes_written" metric_name:"db.tempWrittenInBytesPerSecond"      source_type:"rate"`
		Deadlocks          *int64 `db:"deadlocks"               metric_name:"db.deadlocksPerSecond"               source_type:"rate"`
		TimeSpentReading   *int64 `db:"time_spent_reading_data" metric_name:"db.readTimeInMillisecondsPerSecond"  source_type:"rate"`
		TimeSpentWriting   *int64 `db:"time_spent_writing_data" metric_name:"db.writeTimeInMillisecondsPerSecond" source_type:"rate"`
	}{},
}
