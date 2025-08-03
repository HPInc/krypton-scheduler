package rest

import (
	"net/http"
)

func GetHealthHandler(w http.ResponseWriter, r *http.Request) {
	_ = sendJsonResponse(w, http.StatusOK, nil)
}
