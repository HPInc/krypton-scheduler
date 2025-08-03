-- Create a table to store information about tasks queued to the scheduler
-- service.
CREATE TABLE scheduler.tasks(
  task_id          TIMEUUID,
  device_id        UUID,
  tenant_id        TEXT,
  service_id       TEXT,
  consignment_id   TEXT,
  status           TEXT,
  retry_count      INT,
  create_time      TIMESTAMP,
  start_time       TIMESTAMP,
  end_time         TIMESTAMP,
  unit             INT,
  interval         INT,
  duration         BIGINT,
  run_at           SET<BIGINT>,
  week_days        SET<INT>,
  month_days       SET<INT>,
  start_at         TIMESTAMP,
  immediate        BOOLEAN,
  message_id       TEXT,
  message_type     TEXT,
  task_details     BLOB,
  PRIMARY KEY (device_id, task_id)
)
WITH CLUSTERING ORDER BY (task_id DESC);
