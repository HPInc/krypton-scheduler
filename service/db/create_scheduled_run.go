package db

import (
	"time"

	"go.uber.org/zap"
)

const (
	scheduledRunPartitionKeyDateFormat = "2006-Jan-02"
)

func GetRunPartition(instant time.Time) string {
	return instant.Format(scheduledRunPartitionKeyDateFormat)
}

func (s *ScheduledRun) CreateScheduledRun() error {
	s.RunPartition = GetRunPartition(s.NextRun)

	// Create a new scheduled run for the task in the scheduler database.
	gSessionMutex.RLock()
	err := gSession.Query(scheduledRunsStatements.insert.statement,
		scheduledRunsStatements.insert.names).
		BindStruct(s).
		ExecRelease()
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to add scheduled run to the scheduler database!",
			zap.String("Task ID:", s.TaskID.String()),
			zap.Error(err),
		)
		return err
	}

	schedLogger.Info("Added a new scheduled run into the scheduler database!",
		zap.String("Task ID:", s.TaskID.String()),
	)
	return nil
}
