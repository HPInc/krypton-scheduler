package rest

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Route - used to route REST requests received by the service.
type Route struct {
	Name        string           // Name of the route
	Method      string           // REST method
	Path        string           // Resource path
	HandlerFunc http.HandlerFunc // Request handler function.
}

type routes []Route

// List of Routes and corresponding handler functions registered
// with the router.
var registeredRoutes = routes{
	// Health method.
	Route{
		Name:        "GetHealth",
		Method:      http.MethodGet,
		Path:        "/health",
		HandlerFunc: GetHealthHandler,
	},

	// Metrics method.
	Route{
		Name:        "GetMetrics",
		Method:      http.MethodGet,
		Path:        "/metrics",
		HandlerFunc: promhttp.Handler().(http.HandlerFunc),
	},

	// Task scheduling methods.
	Route{
		Name:        "CreateTask",
		Method:      http.MethodPost,
		Path:        "/api/v1/tasks",
		HandlerFunc: CreateTaskHandler,
	},
	Route{
		Name:        "GetTask",
		Method:      http.MethodGet,
		Path:        "/api/v1/tasks/{task_id}",
		HandlerFunc: GetTaskHandler,
	},
	Route{
		Name:        "ListTasks",
		Method:      http.MethodGet,
		Path:        "/api/v1/tasks",
		HandlerFunc: ListTasksHandler,
	},
	Route{
		Name:        "RemoveTask",
		Method:      http.MethodDelete,
		Path:        "/api/v1/tasks/{task_id}",
		HandlerFunc: RemoveTaskHandler,
	},
}
