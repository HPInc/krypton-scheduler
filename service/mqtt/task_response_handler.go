package mqtt

import (
	"github.com/eclipse/paho.golang/paho"
	"github.com/hpinc/krypton-scheduler/service/metrics"
	"go.uber.org/zap"
)

func TaskResponseMessageHandler(msg *paho.Publish) {
	metrics.MetricMqttTaskResponseMessagesReceived.Inc()

	// Unmarshal the protobuf encoded message and perform some basic
	// validation checks on it.
	decodedPacket, err := decodeAndValidateDeviceMessage(msg)
	if err != nil {
		metrics.MetricMqttInvalidTaskResponseMessages.Inc()
		return
	}

	schedLogger.Info("Received a message on the task responses topic!",
		zap.String("Topic name", msg.Topic),
		zap.Uint16("Packet ID", msg.PacketID),
		zap.String("QOS", string(msg.QoS)),
		zap.Any("Decoded message", decodedPacket),
	)

	// Invoke the scheduler to process the task response message received from
	// the MQTT broker.
	err = messageHandlers.OnTaskResponse(msg.Topic, decodedPacket)
	if err != nil {
		metrics.MetricMqttTaskResponseProcessingErrors.Inc()
		schedLogger.Error("Failed to process task response message!",
			zap.String("Task ID", decodedPacket.TaskId),
			zap.Any("Decoded message", decodedPacket),
			zap.Error(err),
		)
		return
	}
	metrics.MetricMqttTaskResponseMessagesProcessed.Inc()
}
