-- Create a table to store the mapping between consignment IDs and the 
-- corresponding tasks created by the scheduler.
CREATE TABLE scheduler.consignments(
  device_id      UUID,
  tenant_id      TEXT,
  consignment_id TEXT,
  task_id        TIMEUUID,
  create_time    TIMESTAMP,
  status         TEXT,
  PRIMARY KEY ((tenant_id, consignment_id), task_id)
)
WITH CLUSTERING ORDER BY (task_id DESC);
