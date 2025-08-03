package mqtt

import (
	pb "github.com/hpinc/krypton-scheduler/protos"

	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MqttSubscriberRoute struct {
	Topic          string              // Name of the MQTT topic.
	MessageHandler paho.MessageHandler // Message handler function.
}
type routes []MqttSubscriberRoute

var registeredRoutes = routes{
	// Route to handle responses from devices to task requests issued by the
	// scheduler.
	MqttSubscriberRoute{
		Topic:          taskResponsesTopicSubscription,
		MessageHandler: TaskResponseMessageHandler,
	},

	// Route to handle messages from devices to registered services, received
	// from the broker.
	MqttSubscriberRoute{
		Topic:          serviceMessageTopicSubscription,
		MessageHandler: DeviceToServiceMessageHandler,
	},
}

// Initialize routes to handle messages for all MQTT topics subscribed to by
// the scheduler.
func initMqttSubscriberRouter() *paho.StandardRouter {
	router := paho.NewStandardRouter()
	for _, route := range registeredRoutes {
		router.RegisterHandler(route.Topic, route.MessageHandler)
	}
	return router
}

// Validate the fields in the envelope of the device message.
func isValidDeviceMessageEnvelope(msg *pb.DeviceMessage) bool {
	if msg.AccessToken == "" {
		schedLogger.Error("Invalid device access token in the device message")
		return false
	}

	// Task ID must be a valid UUID, if specified.
	if msg.TaskId != "" {
		_, err := uuid.Parse(msg.TaskId)
		if err != nil {
			schedLogger.Error("Invalid task ID in the device message",
				zap.Error(err),
			)
			return false
		}
	}

	// Message type must be specified.
	if msg.MessageType == "" {
		schedLogger.Error("Invalid message type in the device message")
		return false
	}

	return true
}
