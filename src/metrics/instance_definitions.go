package metrics

import (
	"github.com/blang/semver/v4"
)

func generateInstanceDefinitions(version *semver.Version) []*QueryDefinition {
	queryDefinitions := make([]*QueryDefinition, 1, 3)
	v91 := semver.MustParse("9.1.0")
	v92 := semver.MustParse("9.2.0")
	v170 := semver.MustParse("17.0.0")

	if !version.GE(v170) {
		queryDefinitions[0] = instanceDefinitionBase

		if version.GE(v91) {
			queryDefinitions = append(queryDefinitions, instanceDefinition91)
		}

		if version.GE(v92) {
			queryDefinitions = append(queryDefinitions, instanceDefinition92)
		}

	} else {
		queryDefinitions[0] = instanceDefinitionBase170
		queryDefinitions = append(queryDefinitions, instanceDefinition170)
	}

	return queryDefinitions
}

var instanceDefinitionBase = &QueryDefinition{
	query: `SELECT
		BG.checkpoints_timed AS scheduled_checkpoints_performed,
		BG.checkpoints_req AS requested_checkpoints_performed,
		BG.buffers_checkpoint AS buffers_written_during_checkpoint,
		BG.buffers_clean AS buffers_written_by_background_writer,
		BG.maxwritten_clean AS background_writer_stops,
		BG.buffers_backend AS buffers_written_by_backend,
		BG.buffers_alloc AS buffers_allocated
		FROM pg_stat_bgwriter BG;`,

	dataModels: []struct {
		ScheduledCheckpointsPerformed    *int64 `db:"scheduled_checkpoints_performed"      metric_name:"bgwriter.checkpointsScheduledPerSecond"             source_type:"rate"`
		RequestedCheckpointsPerformed    *int64 `db:"requested_checkpoints_performed"      metric_name:"bgwriter.checkpointsRequestedPerSecond"             source_type:"rate"`
		BuffersWrittenDuringCheckpoint   *int64 `db:"buffers_written_during_checkpoint"    metric_name:"bgwriter.buffersWrittenForCheckpointsPerSecond"     source_type:"rate"`
		BuffersWrittenByBackgroundWriter *int64 `db:"buffers_written_by_background_writer" metric_name:"bgwriter.buffersWrittenByBackgroundWriterPerSecond" source_type:"rate"`
		BackgroundWriterStops            *int64 `db:"background_writer_stops"              metric_name:"bgwriter.backgroundWriterStopsPerSecond"            source_type:"rate"`
		BuffersWrittenByBackend          *int64 `db:"buffers_written_by_backend"           metric_name:"bgwriter.buffersWrittenByBackendPerSecond"          source_type:"rate"`
		BuffersAllocated                 *int64 `db:"buffers_allocated"                    metric_name:"bgwriter.buffersAllocatedPerSecond"                 source_type:"rate"`
	}{},
}

var instanceDefinition91 = &QueryDefinition{
	query: `SELECT 
		BG.buffers_backend_fsync AS times_backend_executed_own_fsync
		FROM pg_stat_bgwriter BG;`,

	dataModels: []struct {
		BackendExecutedOwnFsync *int64 `db:"times_backend_executed_own_fsync" metric_name:"bgwriter.backendFsyncCallsPerSecond" source_type:"rate"`
	}{},
}

var instanceDefinition92 = &QueryDefinition{
	query: `SELECT 
		cast(BG.checkpoint_write_time AS bigint) AS time_writing_checkpoint_files_to_disk,
		cast(BG.checkpoint_sync_time AS bigint) AS time_synchronizing_checkpoint_files_to_disk
		FROM pg_stat_bgwriter BG;`,

	dataModels: []struct {
		CheckpointWriteTime *int64 `db:"time_writing_checkpoint_files_to_disk"       metric_name:"bgwriter.checkpointWriteTimeInMillisecondsPerSecond" source_type:"rate"`
		CheckpointSyncTime  *int64 `db:"time_synchronizing_checkpoint_files_to_disk" metric_name:"bgwriter.checkpointSyncTimeInMillisecondsPerSecond"  source_type:"rate"`
	}{},
}

var instanceDefinitionBase170 = &QueryDefinition{
	query: `SELECT
		BG.buffers_clean AS buffers_written_by_background_writer,
		BG.maxwritten_clean AS background_writer_stops,
		BG.buffers_alloc AS buffers_allocated
		FROM pg_stat_bgwriter BG;`,

	dataModels: []struct {
		BuffersWrittenByBackgroundWriter *int64 `db:"buffers_written_by_background_writer" metric_name:"bgwriter.buffersWrittenByBackgroundWriterPerSecond" source_type:"rate"`
		BackgroundWriterStops            *int64 `db:"background_writer_stops"              metric_name:"bgwriter.backgroundWriterStopsPerSecond"            source_type:"rate"`
		BuffersAllocated                 *int64 `db:"buffers_allocated"                    metric_name:"bgwriter.buffersAllocatedPerSecond"                 source_type:"rate"`
	}{},
}

var instanceDefinition170 = &QueryDefinition{
	query: `SELECT 
		BG.num_timed AS scheduled_checkpoints_performed,
		BG.num_requested AS requested_checkpoints_performed,
		BG.buffers_written AS buffers_written_during_checkpoint,
		cast(BG.write_time AS bigint) AS time_writing_checkpoint_files_to_disk,
		cast(BG.sync_time AS bigint) AS time_synchronizing_checkpoint_files_to_disk
		FROM pg_stat_checkpointer BG;`,

	dataModels: []struct {
		ScheduledCheckpointsPerformed  *int64 `db:"scheduled_checkpoints_performed"             metric_name:"checkpointer.checkpointsScheduledPerSecond"                  source_type:"rate"`
		RequestedCheckpointsPerformed  *int64 `db:"requested_checkpoints_performed"             metric_name:"checkpointer.checkpointsRequestedPerSecond"              source_type:"rate"`
		BuffersWrittenDuringCheckpoint *int64 `db:"buffers_written_during_checkpoint"           metric_name:"checkpointer.buffersWrittenForCheckpointsPerSecond"      source_type:"rate"`
		CheckpointWriteTime            *int64 `db:"time_writing_checkpoint_files_to_disk"       metric_name:"checkpointer.checkpointWriteTimeInMillisecondsPerSecond" source_type:"rate"`
		CheckpointSyncTime             *int64 `db:"time_synchronizing_checkpoint_files_to_disk" metric_name:"checkpointer.checkpointSyncTimeInMillisecondsPerSecond"  source_type:"rate"`
	}{},
}
