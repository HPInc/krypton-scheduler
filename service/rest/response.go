package rest

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type FailedRequestError struct {
	HttpCode uint   `json:"code"`
	Reason   string `json:"reason"`
}

const (
	// #nosec spurious G101 (CWE-798): Potential hardcoded credentials
	reasonInvalidAppToken         = "invalid app access token specified"
	reasonRequestParsingFailed    = "error parsing request parameters"
	reasonProtobufUnmarshalFailed = "failed to protobuf unmarshal scheduled task request"
	reasonMissingTaskPayload      = "task payload was not specified"
	reasonInvalidSchedulingUnit   = "invalid schedule task request or invalid scheduling unit specified"
	reasonFailedRequestDbError    = "failed to get task information, invalid request"
	reasonMissingTaskId           = "task_id parameter was not specified"
	reasonMissingDeviceId         = "device_id parameter was not specified"
	reasonMissingTenantId         = "tenant_id parameter was not specified"
	reasonMissingConsignmentId    = "consignment_id parameter was not specified"
)

func sendInternalServerErrorResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

func sendBadRequestErrorResponse(w http.ResponseWriter, requestID string,
	reason string) {
	err := sendJsonResponse(w, http.StatusBadRequest, FailedRequestError{
		HttpCode: http.StatusBadRequest,
		Reason:   reason,
	})
	if err != nil {
		schedLogger.Error("Failed to encode JSON response!",
			zap.String("Request ID: ", requestID),
			zap.Error(err),
		)
	}
}

func sendUnsupportedMediaTypeResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
		http.StatusUnsupportedMediaType)
}

func sendNotFoundErrorResponse(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusNotFound),
		http.StatusNotFound)
}

func sendUnauthorizedErrorResponse(w http.ResponseWriter, requestID string,
	reason string) {
	err := sendJsonResponse(w, http.StatusUnauthorized, FailedRequestError{
		HttpCode: http.StatusUnauthorized,
		Reason:   reason,
	})
	if err != nil {
		schedLogger.Error("Failed to encode JSON response!",
			zap.String("Request ID: ", requestID),
			zap.Error(err),
		)
	}
}

// JSON encode and send the specified payload & the specified HTTP status code.
func sendJsonResponse(w http.ResponseWriter, statusCode int,
	payload interface{}) error {
	w.Header().Set(headerContentType, contentTypeJson)
	w.WriteHeader(statusCode)

	if payload != nil {
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(payload)
		if err != nil {
			schedLogger.Error("Failed to encode JSON response!",
				zap.Error(err),
			)
			sendInternalServerErrorResponse(w)
			return err
		}
	}

	return nil
}
