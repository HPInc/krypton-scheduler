-- Create a table to store information about scheduled runs for a task scheduled with the
-- the scheduler service.
CREATE TABLE scheduler.scheduled_runs(
  run_partition    TEXT, 
  next_run         TIMESTAMP,
  last_run         TIMESTAMP,
  task_id          TIMEUUID,
  device_id        UUID,
  PRIMARY KEY (run_partition, next_run)
)
WITH CLUSTERING ORDER BY (next_run ASC);
