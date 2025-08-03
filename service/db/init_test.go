package db

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hpinc/krypton-scheduler/service/config"
	"go.uber.org/zap"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	t.Errorf("Received: %v (type %v), Expected: %v (type %v)", a,
		reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func shutdownLogger() {
	_ = schedLogger.Sync()
}

func TestMain(m *testing.M) {
	// Initialize logging for the test run.
	logger, err := zap.NewProduction(zap.AddCaller())
	if err != nil {
		fmt.Println("Failed to intialize structured logging for the database tests!")
		os.Exit(2)
	}
	schedLogger = logger

	// Load configuration and rationalize environment overrides for the
	// config file location, and database password.
	// Read and parse the configuration file.
	cfgMgr := config.NewConfigMgr(schedLogger, config.ServiceName)
	if !cfgMgr.Load(false) {
		schedLogger.Error("Failed to load configuration. Exiting!")
		shutdownLogger()
		os.Exit(2)
	}
	cfgMgr.SetDatabasePassword(os.Getenv("SCHEDULER_DB_PASSWORD"))

	// Initialize the connection to the database.
	err = Init(schedLogger, cfgMgr)
	if err != nil {
		schedLogger.Error("Failed to initialize the database. Exiting!",
			zap.Error(err),
		)
		shutdownLogger()
		os.Exit(2)
	}

	retCode := m.Run()
	Shutdown()
	shutdownLogger()
	fmt.Println("Finished running the scheduler database unit tests!")
	os.Exit(retCode)
}
