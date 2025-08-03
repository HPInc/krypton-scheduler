package mqtt

import (
	"fmt"

	"github.com/eclipse/paho.golang/paho"
	"github.com/hpinc/krypton-scheduler/service/common"
)

const (
	// Format strings for topics listened to by the MQTT client/agent on devices
	// for messages sent from the cloud.
	deviceTasksTopicFormat    = "v1/%s/tasks"    // v1/{DEVICE_ID}/tasks
	broadcastTasksTopicFormat = "v1/@devices/%s" // v1/@devices/{SERVICE_ID}

	// Format strings for topics listened to by the scheduler for messages from
	// managed devices.
	taskResponsesTopicSubscription  = "$share/krypton/v1/@cloud/task_responses"
	serviceMessageTopicSubscription = "$share/krypton/v1/@cloud"
)

// A map of topics to which the scheduler subscribes to.
var mqttSubscriptionsMap = []paho.SubscribeOptions{
	{Topic: taskResponsesTopicSubscription, QoS: 0},
	{Topic: serviceMessageTopicSubscription, QoS: 0},
}

func GetMqttTopicForDeviceTask(deviceID string, serviceID string) string {
	// If this is a broadcast task, ensure it is routed to the appropriate
	// broadcast topic.
	if deviceID == common.BroadcastDeviceUuid {
		return fmt.Sprintf(broadcastTasksTopicFormat, serviceID)
	}

	return fmt.Sprintf(deviceTasksTopicFormat, deviceID)
}
