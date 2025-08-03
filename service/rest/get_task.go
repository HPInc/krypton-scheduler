package rest

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/metrics"
	"go.uber.org/zap"
)

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
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
		metrics.MetricGetTaskBadRequests.Inc()
		return
	}

	// Extract the device ID from the query parameter. If not specified,
	// reject the request as bad.
	deviceID := r.URL.Query().Get(paramDeviceID)
	_, err := uuid.Parse(deviceID)
	if err != nil {
		schedLogger.Error("Received a request with an invalid device ID!",
			zap.String("Request ID: ", requestID),
		)
		sendBadRequestErrorResponse(w, requestID, reasonMissingDeviceId)
		metrics.MetricGetTaskBadRequests.Inc()
		return
	}

	// Get the task from the scheduler database.
	foundTask, err := db.GetTaskByID(taskID, deviceID)
	if err != nil {
		schedLogger.Error("Failed to get task information!",
			zap.String("Request ID: ", requestID),
			zap.String("Task ID: ", taskID),
			zap.String("Device ID: ", deviceID),
			zap.Error(err),
		)
		if err == db.ErrInvalidRequest {
			sendBadRequestErrorResponse(w, requestID, reasonFailedRequestDbError)
			metrics.MetricGetTaskBadRequests.Inc()
			return
		} else if err == db.ErrNotFound {
			sendNotFoundErrorResponse(w)
			metrics.MetricGetTaskNotFoundErrors.Inc()
			return
		}

		sendInternalServerErrorResponse(w)
		metrics.MetricGetTaskInternalErrors.Inc()
		return
	}

	// Return the task information to the caller.
	err = sendJsonResponse(w, http.StatusOK, foundTask)
	if err != nil {
		schedLogger.Error("Failed to encode JSON response!",
			zap.String("Request ID: ", requestID),
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricGetTaskInternalErrors.Inc()
		return
	}

	metrics.MetricGetTaskResponses.Inc()
}
