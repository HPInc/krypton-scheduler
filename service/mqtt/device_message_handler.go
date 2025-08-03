package mqtt

import (
	"github.com/eclipse/paho.golang/paho"
	"github.com/hpinc/krypton-scheduler/service/metrics"
	"go.uber.org/zap"
)

func DeviceToServiceMessageHandler(msg *paho.Publish) {
	metrics.MetricMqttServiceMessagesReceived.Inc()

	decodedPacket, err := decodeAndValidateDeviceMessage(msg)
	if err != nil {
		metrics.MetricMqttInvalidServiceMessages.Inc()
		return
	}

	schedLogger.Info("Received a message on the service messages topic!",
		zap.String("Topic name", msg.Topic),
		zap.Uint16("Packet ID", msg.PacketID),
		zap.String("QOS", string(msg.QoS)),
		zap.Any("Decoded message", decodedPacket),
	)

	// Invoke the scheduler to process the device to service message received from
	// the MQTT broker.
	err = messageHandlers.OnDeviceMessage(msg.Topic, decodedPacket)
	if err != nil {
		metrics.MetricMqttServiceMessageProcessingErrors.Inc()
		schedLogger.Error("Failed to process device to service message!",
			zap.String("Access Token", decodedPacket.AccessToken),
			zap.Error(err),
		)
		return
	}
	metrics.MetricMqttServiceMessagesProcessed.Inc()
}
