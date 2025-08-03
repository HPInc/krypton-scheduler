package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/hpinc/krypton-scheduler/service/config"
	"go.uber.org/zap"
)

var (
	schedLogger          *zap.Logger
	debugLogRestRequests bool
	appTokenAuthnEnabled bool
)

const (
	// HTTP server timeouts for the REST endpoint.
	readTimeout        = (time.Second * 5)
	writeTimeout       = (time.Second * 5)
	defaultIdleTimeout = (time.Second * 65)
)

type schedRestService struct {
	// Signal handling to support SIGTERM and SIGINT for the service.
	errChannel  chan error
	stopChannel chan os.Signal

	router *mux.Router
	port   int
}

func newSchedRestService() *schedRestService {
	s := &schedRestService{}

	// Initial signal handling.
	s.errChannel = make(chan error)
	s.stopChannel = make(chan os.Signal, 1)
	signal.Notify(s.stopChannel, syscall.SIGINT, syscall.SIGTERM)

	s.router = initRequestRouter()
	return s
}

func (s *schedRestService) startServing() {
	// Start the HTTP REST server. http.ListenAndServe() always returns
	// a non-nil error
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", s.port),
		Handler:        s.router,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    defaultIdleTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	err := server.ListenAndServe()
	schedLogger.Error("Received a fatal error from http.ListenAndServe",
		zap.Error(err),
	)

	// Signal the error channel so we can shutdown the service.
	s.errChannel <- err
}

func (s *schedRestService) awaitTermination() {
	select {
	case err := <-s.errChannel:
		schedLogger.Error("Shutting down due to a fatal error.",
			zap.Error(err),
		)
	case sig := <-s.stopChannel:
		schedLogger.Info("Received an OS signal to shut down!",
			zap.String("Signal received: ", sig.String()),
		)
	}
}

func Init(logger *zap.Logger, cfgMgr *config.ConfigMgr) {
	schedLogger = logger
	debugLogRestRequests = cfgMgr.IsDebugLoggingRestRequestsEnabled()

	s := newSchedRestService()
	s.port = cfgMgr.GetServerConfig().RestPort
	appTokenAuthnEnabled = cfgMgr.GetServerConfig().AuthenticateRestApiRequests

	// Initialize the REST server and listen for REST requests on a separate
	// goroutine. Report fatal errors via the error channel.
	go s.startServing()
	schedLogger.Info("Started the Scheduler REST service!",
		zap.Int("Port: ", s.port),
	)

	s.awaitTermination()
}

func InitTestServer(logger *zap.Logger, cfgMgr *config.ConfigMgr) {
	schedLogger = logger
	debugLogRestRequests = cfgMgr.IsDebugLoggingRestRequestsEnabled()
}

func ExecuteTestRequest(r *http.Request,
	handler http.HandlerFunc) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	router := initRequestRouter()
	router.ServeHTTP(rec, r)
	return rec
}
