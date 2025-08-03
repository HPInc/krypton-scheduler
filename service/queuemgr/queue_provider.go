package queuemgr

import (
	"github.com/hpinc/krypton-scheduler/service/common"
	"github.com/hpinc/krypton-scheduler/service/config"
	"go.uber.org/zap"
)

// QueueProvider represents the interface implemented by queue providers
// registered with the Scheduler service.
type QueueProvider interface {
	// Initialize the queue provider.
	Init(logger *zap.Logger, cfgMgr *config.ConfigMgr) error

	// Watch the scheduler input queue for new scheduling requests.
	WatchInputQueue(handler common.InputEventHandlerFunc)

	// Watch the scheduler dispatch queue for tasks to be dispatched to the
	// MQTT broker.
	WatchDispatchQueue()

	// Send a message to the specified queue.
	SendMessage(serviceId string, queueTopic string, msg *string) error

	// Send a message to the specified queue.
	SendDispatchQueueMessage(msg *string) error

	// Send a message to the DCM input queue.
	SendDcmInputQueueMessage(msg *string) error

	// Close the provider and cleanup resources.
	Shutdown()
}
