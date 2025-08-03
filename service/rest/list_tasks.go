package rest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/metrics"
	"go.uber.org/zap"
)

// ListTasksResponse - JSON encoded response to the ListTasks REST request.
type ListTasksResponse struct {
	Count        int              `json:"count"`
	Tasks        []db.Consignment `json:"tasks,omitempty"`
	NextPage     string           `json:"next_page,omitempty"`
	ResponseTime time.Time        `json:"response_time"`
}

func ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	var (
		foundConsignments []*db.Consignment
		nextPage          []byte
		err               error
	)

	// Extract the request ID.
	requestID := r.Header.Get(headerRequestID)

	// Check if the request provided a valid app access token.
	if isValidAppAccessToken(r) != nil {
		sendUnauthorizedErrorResponse(w, requestID, reasonInvalidAppToken)
		return
	}

	// Extract the consignment ID from the request query string. If not specified,
	// reject the request as bad.
	consignmentID := r.URL.Query().Get(paramConsignmentID)
	if consignmentID == "" {
		schedLogger.Error("Received an invalid request with no consignment ID!",
			zap.String("Request ID: ", requestID),
		)
		sendBadRequestErrorResponse(w, requestID, reasonMissingConsignmentId)
		metrics.MetricListTasksBadRequests.Inc()
		return
	}

	// Extract the tenant ID from the query parameter. If not specified,
	// reject the request as bad.
	tenantID := r.URL.Query().Get(paramTenantID)
	if tenantID == "" {
		schedLogger.Error("Received an invalid request with no tenant ID!",
			zap.String("Request ID: ", requestID),
		)
		sendBadRequestErrorResponse(w, requestID, reasonMissingTenantId)
		metrics.MetricListTasksBadRequests.Inc()
		return
	}

	foundConsignments, nextPage, err = db.GetTasksForConsignment(tenantID,
		consignmentID, nextPage, 0)
	if err != nil {
		schedLogger.Error("Failed to query tasks for specified consignment from the scheduler database!",
			zap.Error(err),
		)
		sendInternalServerErrorResponse(w)
		metrics.MetricListTasksInternalErrors.Inc()
		return
	}

	w.Header().Set(headerContentType, contentTypeJson)
	resp := ListTasksResponse{
		Count:        len(foundConsignments),
		Tasks:        nil,
		NextPage:     string(nextPage),
		ResponseTime: time.Now(),
	}
	for _, item := range foundConsignments {
		resp.Tasks = append(resp.Tasks, *item)
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		sendInternalServerErrorResponse(w)
		metrics.MetricListTasksInternalErrors.Inc()
		return
	}

	w.WriteHeader(http.StatusOK)
	metrics.MetricListTasksReponses.Inc()
}
