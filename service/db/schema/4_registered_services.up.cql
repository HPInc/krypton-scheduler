-- Create a table to store information about tasks sent to the MQTT broker for
-- delivery to the device for execution.
CREATE TABLE scheduler.registered_services(
  service_id          TEXT,
  name                TEXT,
  owner_aws_account   TEXT,
  topics              MAP<TEXT, TEXT>,
  PRIMARY KEY (service_id)
);
