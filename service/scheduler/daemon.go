package scheduler

import (
	"time"

	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/queuemgr"
	"go.uber.org/zap"
)

const (
	schedulerExecutionQuantum = 1 * time.Minute
)

func runSchedulerDaemon() {
	var (
		foundRuns    []*db.ScheduledRun
		nextPage     []byte
		err          error
		runPartition string
	)

	schedLogger.Info("Starting the scheduler daemon!")

	for {
		// Check if the scheduler daemon goroutine needs to stop execution, if
		// the service is shutting down.
		if schedCtx.Err() != nil {
			schedLogger.Info("Stopping the scheduler daemon! Context has been canceled")
			return
		}

		runPartition = db.GetRunPartition(time.Now())

		for {
			// Retrieve a page worth of scheduled runs that are ready for execution
			// from the scheduled runs table.
			foundRuns, nextPage, err = db.GetScheduledRuns(runPartition, nextPage)
			if err != nil {
				schedLogger.Error("Failed to query next run tasks from the scheduler database!",
					zap.Error(err),
				)
				break
			}

			for _, item := range foundRuns {
				// Check if the scheduler daemon goroutine needs to stop execution, if
				// the service is shutting down.
				if schedCtx.Err() != nil {
					schedLogger.Info("Stopping the scheduler daemon! Context has been canceled")
					return
				}

				schedLogger.Debug("Retrieved a candidate task for execution from database!",
					zap.String("Task ID", item.TaskID.String()),
					zap.Time("Next Run", item.NextRun),
				)

				// Retrieve information about the task such as its payload
				// from the database.
				task, err := db.GetTaskByID(item.TaskID.String(),
					item.DeviceID.String())
				if err != nil {
					schedLogger.Error("Failed to retrieve task information!",
						zap.String("Task ID", item.TaskID.String()),
						zap.Error(err),
					)
					continue
				}

				// Encode the task information to prepare for posting to the
				// dispatch queue.
				payload, err := task.MarshalServiceMessage()
				if err != nil {
					schedLogger.Error("Failed to encode the message for delivery to the device!",
						zap.String("Task ID", task.TaskID.String()),
						zap.Error(err),
					)
					continue
				}

				// Send the task to the dispatch queue. Tasks on the dispatch
				// queue are sent to the MQTT broker for delivery to the device.
				err = queuemgr.Provider.SendDispatchQueueMessage(payload)
				if err != nil {
					schedLogger.Error("Failed to dispatch the scheduled task to the device!",
						zap.String("Task ID", item.TaskID.String()),
						zap.String("Device ID", item.DeviceID.String()),
						zap.Error(err),
					)
					continue
				}

			}

			if len(nextPage) == 0 {
				break
			}
		}

		select {
		case <-time.After(schedulerExecutionQuantum):

		case <-schedCtx.Done():
			schedLogger.Info("Received signal to stop the scheduler daemon!")
			return
		}
	}
}
