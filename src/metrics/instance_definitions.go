package metrics

import (
	"github.com/blang/semver"
)

func generateInstanceDefinitions(version semver.Version) []*QueryDefinition {
	queryDefinitions := make([]*QueryDefinition, 1, 3)
	v91 := semver.MustParse("9.1.0")
	v92 := semver.MustParse("9.2.0")

	queryDefinitions[0] = instanceDefinitionBase

	if version.GE(v91) {
		queryDefinitions = append(queryDefinitions, instanceDefinition91)
	}

	if version.GE(v92) {
		queryDefinitions = append(queryDefinitions, instanceDefinition92)
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

	dataModels: &[]struct {
		ScheduledCheckpointsPerformed    *int `db:"scheduled_checkpoints_performed"      metric_name:"bgwriter.checkpointsScheduled"             source_type:"rate"`
		RequestedCheckpointsPerformed    *int `db:"requested_checkpoints_performed"      metric_name:"bgwriter.checkpointsRequested"             source_type:"rate"`
		BuffersWrittenDuringCheckpoint   *int `db:"buffers_written_during_checkpoint"    metric_name:"bgwriter.buffersWrittenForCheckpoints"     source_type:"rate"`
		BuffersWrittenByBackgroundWriter *int `db:"buffers_written_by_background_writer" metric_name:"bgwriter.buffersWrittenByBackgroundWriter" source_type:"rate"`
		BackgroundWriterStops            *int `db:"background_writer_stops"              metric_name:"bgwriter.backgroundWriterStops"            source_type:"rate"`
		BuffersWrittenByBackend          *int `db:"buffers_written_by_backend"           metric_name:"bgwriter.buffersWrittenByBackend"          source_type:"rate"`
		BuffersAllocated                 *int `db:"buffers_allocated"                    metric_name:"bgwriter.buffersAllocated"                 source_type:"rate"`
	}{},
}

var instanceDefinition91 = &QueryDefinition{
	query: `SELECT 
		BG.buffers_backend_fsync AS times_backend_executed_own_fsync
		FROM pg_stat_bgwriter BG;`,

	dataModels: &[]struct {
		BackendExecutedOwnFsync *int `db:"times_backend_executed_own_fsync" metric_name:"bgwriter.backendFsyncCalls" source_type:"rate"`
	}{},
}

var instanceDefinition92 = &QueryDefinition{
	query: `SELECT 
		cast(BG.checkpoint_write_time AS bigint) AS time_writing_checkpoint_files_to_disk,
		cast(BG.checkpoint_sync_time AS bigint) AS time_synchronizing_checkpoint_files_to_disk
		FROM pg_stat_bgwriter BG;`,

	dataModels: &[]struct {
		CheckpointWriteTime *int `db:"time_writing_checkpoint_files_to_disk"       metric_name:"bgwriter.checkpointWriteTimeInMilliseconds" source_type:"rate"`
		CheckpointSynTime   *int `db:"time_synchronizing_checkpoint_files_to_disk" metric_name:"bgwriter.checkpointSyncTimeInMilliseconds"  source_type:"rate"`
	}{},
}
