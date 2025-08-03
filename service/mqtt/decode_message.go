package mqtt

import (
	"errors"

	"github.com/eclipse/paho.golang/paho"
	pb "github.com/hpinc/krypton-scheduler/protos"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidDeviceMessage = errors.New("invalid device message")
)

func decodeAndValidateDeviceMessage(msg *paho.Publish) (*pb.DeviceMessage, error) {
	var decodedPacket pb.DeviceMessage

	// Unmarshal the MQTT request received.
	err := proto.Unmarshal(msg.Payload, &decodedPacket)
	if err != nil {
		schedLogger.Error("Failed to unmarshal message received on MQTT task responses topic!",
			zap.String("Topic name:", msg.Topic),
			zap.Error(err),
		)
		return nil, err
	}

	if decodedPacket.Payload == nil {
		schedLogger.Error("Received invalid message on MQTT task responses topic!",
			zap.String("Topic name:", msg.Topic),
		)
		return nil, ErrInvalidDeviceMessage
	}

	// Validate the fields in the MQTT message envelope.
	if !isValidDeviceMessageEnvelope(&decodedPacket) {
		schedLogger.Error("Failed to process MQTT message with invalid envelope!",
			zap.String("Topic name:", msg.Topic),
		)
		return nil, ErrInvalidDeviceMessage
	}

	return &decodedPacket, nil
}
