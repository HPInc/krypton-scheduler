package queuemgr

import (
	"github.com/hpinc/krypton-scheduler/service/common"
	"github.com/hpinc/krypton-scheduler/service/config"
	"github.com/hpinc/krypton-scheduler/service/queuemgr/sqs_provider"

	"go.uber.org/zap"
)

var (
	// Structured logging using Uber Zap.
	schedLogger *zap.Logger

	// Queue provider used by the scheduler service.
	Provider QueueProvider
)

func Init(logger *zap.Logger, cfgMgr *config.ConfigMgr,
	inputEventHandler common.InputEventHandlerFunc) error {
	schedLogger = logger

	// Initialize the AWS SQS queue provider. For now, we only have a single
	// provider. When more providers are added, update this logic to select
	// the right queue provider based on configuration.
	Provider = sqs_provider.NewSqsProvider()

	// Initialize the selected queue provider
	err := Provider.Init(schedLogger, cfgMgr)
	if err != nil {
		schedLogger.Error("Failed to initialize the selected queue provider.",
			zap.Error(err),
		)
		return err
	}

	// Watch the scheduler input queue for new scheduling requests.
	go Provider.WatchInputQueue(inputEventHandler)

	// Watch the scheduler dispatch queue for tasks to be dispatched to the
	// MQTT broker for delivery to devices.
	go Provider.WatchDispatchQueue()

	return nil
}

func Shutdown() {
	schedLogger.Info("HP Scheduler: signalling shutdown to queue subscriber")
	Provider.Shutdown()
}
