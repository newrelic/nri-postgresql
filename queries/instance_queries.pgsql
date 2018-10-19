Select 
BG.checkpoints_timed AS scheduled_checkpoints_performed,
BG.checkpoints_req AS requested_checkpoints_performed,
BG.buffers_checkpoint AS buffers_written_during_checkpoint,
BG.buffers_clean AS buffers_written_by_background_writer,
BG.maxwritten_clean AS imes_background_writer_stopped_due_to_too_many_buffers,
BG.buffers_backend AS buffers_written_by_backend,
BG.buffers_alloc AS buffers_allocated
FROM pg_stat_bgwriter BG;

-- Requires 9.1
Select 
BG.buffers_backend_fsync AS times_backend_executed_own_fsync
FROM pg_stat_bgwriter BG;

-- Requires 9.2
Select 
cast(BG.checkpoint_write_time AS bigint) AS time_writing_checkpoint_files_to_disk,
cast(BG.checkpoint_sync_time AS bigint) AS time_synchronizing_checkpoint_files_to_disk
FROM pg_stat_bgwriter BG;
