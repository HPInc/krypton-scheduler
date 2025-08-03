package scheduler

import (
	b64 "encoding/base64"
	"strings"

	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/dstsclient"
	"github.com/hpinc/krypton-scheduler/service/queuemgr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func handleTaskResponseMessage(mqttTopic string, message *pb.DeviceMessage) error {

	// Validate the device access token presented by the device in the message.
	claims, err := dstsclient.ValidateDeviceAccessToken(message.AccessToken)
	if err != nil {
		schedLogger.Error("Failed to validate the access token presented by the device!",
			zap.Error(err),
		)
		return err
	}

	// Retrieve information about the task referenced in the message.
	foundTask, err := db.GetTaskByID(message.TaskId, claims.Subject)
	if err != nil {
		schedLogger.Error("Failed to retrieve task",
			zap.String("Task ID", message.TaskId),
			zap.String("Device ID", claims.Subject),
			zap.Error(err),
		)
		return err
	}

	// Validate that metadata in the task response message matches the original
	// task sent down to the device.
	if foundTask.ServiceID != claims.ManagementService {
		schedLogger.Error("Mismatched service ID in task response message!",
			zap.String("Task ID", message.TaskId),
			zap.String("Service ID (msg)", claims.ManagementService),
			zap.String("Service ID (task)", foundTask.ServiceID),
		)
		return ErrInvalidServiceID
	}

	if foundTask.TenantID != claims.TenantID {
		schedLogger.Error("Mismatched tenant ID in task response message!",
			zap.String("Tenant ID (msg)", claims.TenantID),
			zap.String("Tenant ID (task)", foundTask.TenantID),
		)
		return ErrInvalidTenantID
	}

	// Update the status of the task in the database.
	switch strings.ToLower(message.TaskStatus) {
	case "complete", "success":
		err = db.MarkTaskComplete(foundTask)
	case "failed", "error":
		err = db.MarkTaskFailed(foundTask)
	default:
		schedLogger.Error("Device message (task response) specified an invalid status!",
			zap.String("Task ID", message.TaskId),
			zap.String("Device ID", claims.Subject),
		)
		return ErrInvalidRequest
	}
	if err != nil {
		schedLogger.Error("Failed to update task status to complete!",
			zap.String("Task ID: ", message.TaskId),
			zap.String("Device ID: ", claims.Subject),
			zap.Error(err),
		)
	}

	// Determine the appropriate registered service queue topic to which this
	// message should be dispatched.
	queueTopic := db.GetServiceQueueTopic(foundTask.ServiceID, mqttTopic)
	if queueTopic == "" {
		schedLogger.Error("Cannot determine a service queue topic to dispatch MQTT message!",
			zap.String("Service ID", claims.ManagementService),
			zap.String("MQTT topic", mqttTopic),
		)
		return ErrInvalidMessageType
	}

	payload, err := proto.Marshal(&pb.DeviceEvent{
		Version:       1,
		ServiceId:     claims.ManagementService,
		DeviceId:      claims.Subject,
		TaskId:        message.TaskId,
		ConsignmentId: foundTask.ConsignmentID,
		TenantId:      claims.TenantID,
		TaskStatus:    message.TaskStatus,
		MessageId:     message.MessageId,
		MessageType:   message.MessageType,
		Payload:       message.Payload,
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
			zap.Error(err),
		)
		return err
	}

	return nil
}
