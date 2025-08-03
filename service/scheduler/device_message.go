package scheduler

import (
	b64 "encoding/base64"
	"encoding/json"

	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/dstsclient"
	"github.com/hpinc/krypton-scheduler/service/queuemgr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	// Event containing configuration information sent from the device to the cloud.
	configMessageType = "CFG.E"
)

// Handler to process task response messages received by the scheduler on the
// MQTT topic from the broker.
func handleDeviceToServiceMessage(mqttTopic string, message *pb.DeviceMessage) error {

	// Validate the device access token presented by the device in the message.
	claims, err := dstsclient.ValidateDeviceAccessToken(message.AccessToken)
	if err != nil {
		schedLogger.Error("Failed to validate the access token presented by the device!",
			zap.Error(err),
		)
		return err
	}

	// Determine the appropriate registered service queue topic to which this
	// message should be dispatched.
	queueTopic := db.GetServiceQueueTopic(claims.ManagementService, mqttTopic)
	if queueTopic == "" {
		schedLogger.Error("Cannot determine a service queue topic to dispatch MQTT message!",
			zap.String("Service ID", claims.ManagementService),
			zap.String("MQTT topic", mqttTopic),
		)
		return ErrInvalidMessageType
	}

	payload, err := proto.Marshal(&pb.DeviceEvent{
		Version:     1,
		ServiceId:   claims.ManagementService,
		DeviceId:    claims.Subject,
		TaskId:      message.TaskId,
		TenantId:    claims.TenantID,
		MessageId:   message.MessageId,
		MessageType: message.MessageType,
		Payload:     message.Payload,
	})
	if err != nil {
		schedLogger.Error("Failed to marshal device message for sending to the service!",
			zap.Error(err),
		)
		return err
	}

	// Base 64 encode the protobuf encoded byte stream for transmission over
	// SQS.
	b64Payload := b64.StdEncoding.EncodeToString(payload)

	// Send the task response message to the corresponding queue for the target
	// service.
	err = queuemgr.Provider.SendMessage(claims.ManagementService, queueTopic,
		&b64Payload)
	if err != nil {
		schedLogger.Error("Failed to dispatch the task response message to the service!",
			zap.String("Task ID", message.TaskId),
			zap.String("Device ID", claims.Subject),
			zap.String("MQTT topic", mqttTopic),
			zap.String("Service queue topic", queueTopic),
			zap.Error(err),
		)
		return err
	}

	// Dispatch configuration events to the DCM.
	if message.MessageType == configMessageType {
		err = sendConfigEventMessageToDcm(claims.TenantID, claims.Subject,
			claims.ManagementService, message.Payload)
		if err != nil {
			schedLogger.Error("Failed to dispatch the task response message to the service!",
				zap.String("Task ID", message.TaskId),
				zap.String("Device ID", claims.Subject),
				zap.String("MQTT topic", mqttTopic),
				zap.String("Service queue topic", queueTopic),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}

type DcmEvent struct {
	TenantID    string `json:"tenant_id"`
	DeviceID    string `json:"device_id"`
	MgmtService string `json:"ms"`
	Payload     string `json:"payload"`
}

// JSON encode a device configuration event and send it to the DCM service.
func sendConfigEventMessageToDcm(tenantID string, deviceID string, ms string,
	payload []byte) error {
	jsonPayload, err := json.Marshal(DcmEvent{
		TenantID:    tenantID,
		DeviceID:    deviceID,
		MgmtService: ms,
		Payload:     string(payload),
	})
	if err != nil {
		schedLogger.Error("Failed to encode config event message!",
			zap.String("Tenant ID", tenantID),
			zap.String("Device ID", deviceID),
			zap.Error(err),
		)
		return err
	}

	b64Payload := b64.StdEncoding.EncodeToString(jsonPayload)

	err = queuemgr.Provider.SendDcmInputQueueMessage(&b64Payload)
	if err != nil {
		schedLogger.Error("Failed to dispatch the config event message to the DCM service!",
			zap.String("Tenant ID", tenantID),
			zap.String("Device ID", deviceID),
			zap.Error(err),
		)
		return err
	}
	return nil
}
