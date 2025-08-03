package scheduler

import (
	"context"

	"github.com/hpinc/krypton-scheduler/service/config"
	"github.com/hpinc/krypton-scheduler/service/mqtt"
	"go.uber.org/zap"
)

var (
	schedLogger *zap.Logger
	schedCtx    context.Context
	cancelFunc  context.CancelFunc
)

// Initialize the scheduler.
func Init(logger *zap.Logger, cfgMgr *config.ConfigMgr) error {
	schedLogger = logger

	schedCtx, cancelFunc = context.WithCancel(context.Background())

	// Initialize the MQTT publisher & subscriber.
	err := mqtt.Init(schedLogger, cfgMgr, &mqtt.MessageHandlers{
		OnDeviceMessage: handleDeviceToServiceMessage,
		OnTaskResponse:  handleTaskResponseMessage,
	})
	if err != nil {
		schedLogger.Error("Failed to initialize the queue manager!",
			zap.Error(err),
		)
		cancelFunc()
		return err
	}

	// Start off a goroutine that schedules tasks.
	go runSchedulerDaemon()

	schedLogger.Info("Starting the scheduler engine!")
	return nil
}

// Shutdown the scheduler.
func Shutdown() {
	cancelFunc()
	_ = mqtt.Shutdown()
}
