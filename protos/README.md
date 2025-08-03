# Which proto file should I use?

## SQS (queue) messages

### Device management service to Krypton Scheduler
When the device management service requests the scheduler to create a new scheduled task, it uses the message format defined in the scheduled_task.proto file.
DMS -> Scheduler --> scheduled_task.proto

### Scheduler to the device management service
When the scheduler receives an MQTT message from the device, it converts it to a device_event proto message and posts it to the device management service's input queue.
Scheduler -> DMS --> device_event.proto


## MQTT messages

### MQTT message: from scheduler to the device
Messages sent from the scheduler to the device over the MQTT channel are defined in the mqtt_service_message.proto file. This includes tasks scheduled to be sent to the device by the device management service.
Scheduler -> Device (over MQTT) --> mqtt_service_message.proto

### MQTT message: from device to the scheduler
Device responses to the cloud (scheduler) are sent in the format specified in the mqtt_device_message.proto file.
Device -> Scheduler (over MQTT) --> mqtt_device_message.proto

