package scheduler

import (
	"time"

	"github.com/google/uuid"
	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/common"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/metrics"
	"go.uber.org/zap"
)

var ScheduleRequestHandlerFunc = handleSchedulerInputQueueRequest

// Process a request to schedule tasks - may be invoked either from the REST
// endpoint or while processing messages received on the scheduler input queue.
func handleSchedulerInputQueueRequest(request *pb.CreateScheduledTaskRequest,
	source string) (*pb.CreateScheduledTaskResponse, error) {

	switch source {
	case common.SchedulerRequestSourceEvent, common.SchedulerRequestSourceRest:
		break
	default:
		return nil, ErrInvalidRequest
	}

	// Reject requests with an un-registered service ID.
	if !db.IsValidServiceId(request.ServiceId) {
		schedLogger.Error("Unsupported service ID was specified in the request!",
			zap.String("Service ID", request.ServiceId),
		)
		return nil, ErrInvalidRequest
	}

	// Reject the request if the consignment ID was not specified.
	if request.ConsignmentId == "" {
		schedLogger.Error("No consignment ID was specified in the request!",
			zap.String("Consignment ID", request.ConsignmentId),
		)
		return nil, ErrInvalidRequest
	}

	// Validate that the required parameters were specified.
	if request.DeviceIds == nil {
		schedLogger.Error("No valid device ID was specified in the request!",
			zap.String("Consignment ID", request.ConsignmentId),
		)
		return nil, ErrInvalidRequest
	}

	// Check if the first device ID signifies this is a broadcast message/task
	// being scheduled - broadcast tasks must not specify any other device IDs.
	if request.DeviceIds[0] == common.BroadcastDeviceID {
		if len(request.DeviceIds) != 1 {
			schedLogger.Error("Broadcast message specified along with other device IDs.",
				zap.String("Consignment ID", request.ConsignmentId),
			)
			return nil, ErrInvalidRequest
		}

		// For broadcast task requests, the tenant ID is inconsequential.
		request.TenantId = common.BroadcastDeviceUuid
	} else {
		// For non-broadcast task requests, validate that the tenant ID was
		// specified.
		_, err := uuid.Parse(request.TenantId)
		if err != nil {
			schedLogger.Error("Invalid tenant ID was specified!",
				zap.String("Consignment ID", request.ConsignmentId),
				zap.Error(err),
			)
			return nil, ErrInvalidRequest
		}
	}

	if request.Payload == nil {
		schedLogger.Error("Invalid request payload specified!",
			zap.String("Consignment ID: ", request.ConsignmentId),
		)
		return nil, ErrInvalidRequest
	}

	response := &pb.CreateScheduledTaskResponse{
		Version:        request.Version,
		TaskCount:      0,
		ErrorCount:     0,
		ConsignmentId:  request.ConsignmentId,
		TenantId:       request.TenantId,
		TasksScheduled: []*pb.TaskInfo{},
	}

	for index, deviceID := range request.DeviceIds {
		if (index != 0) && (deviceID == common.BroadcastDeviceID) {
			// Broadcast tasks must not specify any other device IDs
			// in the request. Ignore any broadcast task request that
			// is not the sole (i.e. index = 0) device ID in the request.
			schedLogger.Error("Request mixes broadcast task request with requests for specific devices!")
			response.ErrorCount++
			continue
		}

		// Parse the provided task schedule and schedule the task.
		newTask, err := NewScheduledTask(time.UTC, deviceID, request).
			ParseSchedule(request.Schedule).
			Schedule()
		if err != nil {
			switch err {
			case db.ErrInvalidRequest, ErrInvalidSchedulingUnit,
				ErrInvalidScheduleType:
				metrics.MetricCreateTaskBadRequests.Inc()

			default:
				metrics.MetricCreateTaskInternalErrors.Inc()
			}

			schedLogger.Error("Failed to create a new scheduled task!",
				zap.String("Consignment ID", request.ConsignmentId),
				zap.String("Tenant ID", request.TenantId),
				zap.String("Device ID", deviceID),
				zap.Error(err),
			)
			response.ErrorCount++
		} else {
			response.TaskCount++
			if source == common.SchedulerRequestSourceRest {
				response.TasksScheduled = append(
					response.TasksScheduled, &pb.TaskInfo{
						TaskId:   newTask.TaskInfo.TaskID.String(),
						DeviceId: deviceID,
						Status:   newTask.TaskInfo.Status,
					})
			}
			schedLogger.Debug("Queued a scheduled task with the scheduler!",
				zap.String("Consignment ID", request.ConsignmentId),
				zap.String("Tenant ID", request.TenantId),
				zap.String("Device ID", deviceID),
				zap.String("Task ID", newTask.TaskInfo.TaskID.String()),
			)
		}
	}

	schedLogger.Info("Processed request to schedule tasks!",
		zap.String("Consignment ID", request.ConsignmentId),
		zap.String("Tenant ID", request.TenantId),
		zap.Int("Number of devices", len(response.TasksScheduled)),
		zap.Uint32("Failures", response.ErrorCount),
	)

	return response, nil
}
