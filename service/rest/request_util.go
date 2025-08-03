package rest

const (
	// REST request headers and expected header values.
	headerContentType         = "Content-Type"
	headerRequestID           = "request_id"
	headerAuthorization       = "Authorization"
	bearerToken               = "Bearer "
	contentTypeFormUrlEncoded = "application/x-www-form-urlencoded"
	contentTypeProtobuf       = "application/x-protobuf"
	contentTypeJson           = "application/json"

	// REST request parameter names.
	paramTenantID      = "tenant_id"
	paramDeviceID      = "device_id"
	paramTaskID        = "task_id"
	paramConsignmentID = "consignment_id"
	paramTaskDetails   = "task_details"
	paramTaskSchedule  = "task_schedule"
)
