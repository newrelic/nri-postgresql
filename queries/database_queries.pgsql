-- min version 9.1.0
SELECT 
D.datname,
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
DBC.confl_snapshot AS queries_canceled_due_to_old_snapshots,
DBC.confl_bufferpin AS queries_canceled_due_to_pinned_buffers,
DBC.confl_deadlock AS queries_canceled_due_to_deadlocks
FROM pg_stat_database SD 
INNER JOIN pg_database D ON D.datname = SD.datname 
INNER JOIN pg_stat_database_conflicts DBC ON DBC.datname = D.datname 
LEFT JOIN pg_tablespace TS ON TS.oid = D.dattablespace 
WHERE D.datistemplate = FALSE AND D.datname IS NOT NULL;
-- need to append a 'AND D.datname IN (comma separated DB list)'

-- min version 9.2.0
SELECT 
SD.temp_files AS temporary_files_created,
SD.temp_bytes AS temporary_bytes_written,
SD.deadlocks AS deadlocks,
cast(SD.blk_read_time AS bigint) AS time_spent_reading_data,
cast(SD.blk_write_time AS bigint) AS time_spent_writing_data
FROM pg_stat_database SD 
INNER JOIN pg_database D ON D.datname = SD.datname 
INNER JOIN pg_stat_database_conflicts DBC ON DBC.datname = D.datname 
LEFT JOIN pg_tablespace TS ON TS.oid = D.dattablespace 
WHERE D.datistemplate = FALSE AND D.datname IS NOT NULL;
-- need to append a 'AND D.datname IN (comma separated DB list)'


-- max version 9.0.999
SELECT 
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
WHERE D.datistemplate = FALSE AND D.datname IS NOT NULL;
-- need to append a 'AND D.datname IN (comma separated DB list)'
