package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hpinc/krypton-scheduler/service/config"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/queuemgr"
	"github.com/hpinc/krypton-scheduler/service/rest"
	"github.com/hpinc/krypton-scheduler/service/scheduler"
	"go.uber.org/zap"
)

var (
	// --version: displays versioning information.
	versionFlag = flag.Bool("version", false,
		"Print the version of the service and exit!")

	// --log_level: specify the logging level to use.
	logLevelFlag = flag.String("log_level", "",
		"Specify the logging level.")

	gitCommitHash string
	builtAt       string
	builtBy       string
	builtOn       string

	// Service configuration settings.
	cfgMgr *config.ConfigMgr
)

func printVersionInformation() {
	fmt.Printf("%s: version information\n", config.ServiceName)
	fmt.Printf("- Git commit hash: %s\n - Built at: %s\n - Built by: %s\n - Built on: %s\n",
		gitCommitHash, builtAt, builtBy, builtOn)
}

func main() {
	// Parse the command line flags.
	flag.Parse()
	if *versionFlag {
		printVersionInformation()
		return
	}

	// Initialize structured logging.
	initLogger(*logLevelFlag)

	// Read and parse the configuration file.
	cfgMgr = config.NewConfigMgr(schedLogger, config.ServiceName)
	if !cfgMgr.Load(false) {
		schedLogger.Error("Failed to load configuration. Exiting!")
		shutdownLogger()
		os.Exit(2)
	}

	// Initialize the connection to the database.
	err := db.Init(schedLogger, cfgMgr)
	if err != nil {
		schedLogger.Error("Failed to initialize the database. Exiting!",
			zap.Error(err),
		)
		shutdownLogger()
		os.Exit(2)
	}

	// Initialize the scheduler engine.
	err = scheduler.Init(schedLogger, cfgMgr)
	if err != nil {
		schedLogger.Error("Failed to initialize the scheduler engine!",
			zap.Error(err),
		)
		db.Shutdown()
		shutdownLogger()
		os.Exit(2)
	}

	// Initialize the queue manager.
	err = queuemgr.Init(schedLogger, cfgMgr, scheduler.ScheduleRequestHandlerFunc)
	if err != nil {
		schedLogger.Error("Failed to initialize the queue manager!",
			zap.Error(err),
		)
		scheduler.Shutdown()
		db.Shutdown()
		shutdownLogger()
		os.Exit(2)
	}

	// Initialize the REST server and listen for REST requests on a separate
	// goroutine. Report fatal errors via the error channel.
	rest.Init(schedLogger, cfgMgr)

	// Cleanup various subsystems and exit
	scheduler.Shutdown()
	db.Shutdown()
	shutdownLogger()
	fmt.Printf("%s: Goodbye!", config.ServiceName)
}
