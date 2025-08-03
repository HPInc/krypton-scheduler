package rest

import (
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/common"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/metrics"
	"github.com/hpinc/krypton-scheduler/service/scheduler"
)

// CreateTaskResponse - JSON encoded response to the CreateTask REST request.
type CreateTaskResponse struct {
	TaskID     string    `json:"task_id"`
	CreateTime time.Time `json:"create_time"`
}

// CreateTask REST request handler - adds the specified task request to the task
// scheduler database.
// Parameters:
//   - tenant_id - The unique ID of the tenant to which the device belongs.
//   - device_id - The unique device ID of the device to which the task needs to
//     be dispatched.
//   - task_details - The payload specifying what needs to be done as part of the
//     task. This payload is not interpreted by the scheduler and is passed to the
//     device as is.
func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the contents of the POST were provided using protobuf content type.
	if r.Header.Get(headerContentType) != contentTypeProtobuf {
		sendUnsupportedMediaTypeResponse(w)
		metrics.MetricCreateTaskBadRequests.Inc()
		return
	}

	// Extract the request ID.
	requestID := r.Header.Get(headerRequestID)

	// Check if the request provided a valid app access token.
	if isValidAppAccessToken(r) != nil {
		sendUnauthorizedErrorResponse(w, requestID, reasonInvalidAppToken)
		return
	}

	// Retrieve the request parameters from the form.
	reqBytes, err := io.ReadAll(r.Body)
	if err != nil {
		schedLogger.Error("Failed to retrieve the request body!",
			zap.Error(err),
		)
		sendBadRequestErrorResponse(w, requestID, reasonRequestParsingFailed)
		metrics.MetricCreateTaskBadRequests.Inc()
		return
	}

	// Unmarshal the request.
	var request pb.CreateScheduledTaskRequest
	err = proto.Unmarshal(reqBytes, &request)
	if err != nil {
		schedLogger.Error("Failed to unmarshal request received at scheduler REST endpoint!",
			zap.Error(err),
		)
		sendBadRequestErrorResponse(w, requestID, reasonProtobufUnmarshalFailed)
		metrics.MetricCreateTaskBadRequests.Inc()
		return
	}

	schedLogger.Info("Parsed create_task request!",
		zap.String("Service ID", request.ServiceId),
		zap.Strings("Device ID", request.DeviceIds),
		zap.String("Consignment ID", request.ConsignmentId),
		zap.String("Tenant ID", request.TenantId),
		zap.String("Schedule", request.Schedule),
		zap.ByteString("Request payload", request.Payload),
	)

	if request.Payload == nil {
		schedLogger.Error("Invalid scheduled task request received at scheduler REST endpoint!",
			zap.Error(err),
		)
		sendBadRequestErrorResponse(w, requestID, reasonMissingTaskPayload)
		metrics.MetricCreateTaskBadRequests.Inc()
		return
	}

	// Parse the provided task schedule and schedule the task with the scheduler.
	response, err := scheduler.ScheduleRequestHandlerFunc(&request,
		common.SchedulerRequestSourceRest)
	if err != nil {
		schedLogger.Error("Failed to create a new scheduled task!",
			zap.String("Request ID: ", requestID),
			zap.Error(err),
		)
		switch err {
		case db.ErrInvalidRequest, scheduler.ErrInvalidSchedulingUnit,
			scheduler.ErrInvalidRequest:
			sendBadRequestErrorResponse(w, requestID, reasonInvalidSchedulingUnit)
			metrics.MetricCreateTaskBadRequests.Inc()
			return

		default:
			sendInternalServerErrorResponse(w)
			metrics.MetricCreateTaskInternalErrors.Inc()
			return
		}
	}

	// Return the generated task ID to the caller.
	err = sendJsonResponse(w, http.StatusCreated, response)
	if err != nil {
		schedLogger.Error("Failed to encode JSON response!",
			zap.String("Request ID: ", requestID),
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricCreateTaskInternalErrors.Inc()
		return
	}

	metrics.MetricCreateTaskResponses.Inc()
}
