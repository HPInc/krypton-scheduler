package rest

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/metrics"
)

// RemoveTask REST request handler - removes the specified task request from the
// task scheduler database.
// Parameters:
//   - task_id - The unique ID of the task being removed is specified in the URL
//     eg. api/v1/tasks/{task_id}
//   - device_id - The unique device ID of the device to which the task needs to
//     be dispatched.
func RemoveTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the request ID.
	requestID := r.Header.Get(headerRequestID)

	// Check if the request provided a valid app access token.
	if isValidAppAccessToken(r) != nil {
		sendUnauthorizedErrorResponse(w, requestID, reasonInvalidAppToken)
		return
	}

	// Extract the task ID from the request path. If not specified,
	// reject the request as bad.
	params := mux.Vars(r)
	taskID := params[paramTaskID]
	if taskID == "" {
		schedLogger.Error("Received an invalid request with no task ID!",
			zap.String("Request ID: ", requestID),
		)
		sendBadRequestErrorResponse(w, requestID, reasonMissingTaskId)
		metrics.MetricRemoveTaskBadRequests.Inc()
		return
	}

	// Extract the device ID from the query parameter. If not specified,
	// reject the request as bad.
	deviceID := r.URL.Query().Get(paramDeviceID)
	_, err := uuid.Parse(deviceID)
	if err != nil {
		schedLogger.Error("Request contains an invalid device ID!",
			zap.String("Request ID: ", requestID),
		)
		sendBadRequestErrorResponse(w, requestID, reasonMissingDeviceId)
		metrics.MetricRemoveTaskBadRequests.Inc()
		return
	}

	// Remove the task from the scheduler database.
	removeTask := db.Task{}
	err = removeTask.RemoveTask(taskID, deviceID)
	if err != nil {
		schedLogger.Error("Failed to remove the specified task!",
			zap.String("Request ID: ", requestID),
			zap.String("Task ID: ", taskID),
			zap.String("Device ID: ", deviceID),
			zap.Error(err),
		)
		if err == db.ErrInvalidRequest {
			sendBadRequestErrorResponse(w, requestID, reasonFailedRequestDbError)
			metrics.MetricRemoveTaskBadRequests.Inc()
			return
		}

		sendInternalServerErrorResponse(w)
		metrics.MetricRemoveTaskInternalErrors.Inc()
		return
	}

	w.WriteHeader(http.StatusOK)
	metrics.MetricRemoveTaskResponses.Inc()
}
