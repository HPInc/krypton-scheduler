package scheduler

import (
	"fmt"
	"strings"
	"time"

	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/common"
	"github.com/hpinc/krypton-scheduler/service/db"
	"github.com/hpinc/krypton-scheduler/service/queuemgr"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Represents a scheduled task that is submitted for execution by the Krypton
// scheduler.
type ScheduledTask struct {
	location     *time.Location   // location of time used to schedule the task.
	TaskInfo     *db.Task         // Task information stored in the scheduler database.
	ScheduleInfo *db.ScheduledRun // Information about the schedule for this task.
	error        error            // error related to task
	cronSchedule cron.Schedule    // stores the schedule when a task uses cron
}

func NewScheduledTask(loc *time.Location, deviceID string,
	request *pb.CreateScheduledTaskRequest) *ScheduledTask {
	var err error
	newTask := &ScheduledTask{
		location:     loc,
		TaskInfo:     nil,
		ScheduleInfo: nil,
	}

	// Initialize the task information.
	newTask.TaskInfo, err = db.NewTask(&request.TenantId,
		&deviceID, &request.ConsignmentId, &request.Payload)
	if err != nil {
		newTask.error = wrapOrError(newTask.error, err)
		return newTask
	}

	newTask.TaskInfo.ServiceID = request.ServiceId
	newTask.TaskInfo.MessageType = request.MessageType
	newTask.TaskInfo.MessageId = request.MessageId

	// Initialize the schedule information.
	newTask.ScheduleInfo = db.NewScheduledRun(newTask.TaskInfo)
	return newTask
}

// At - schedules the Task to run at a specific time of day in the form
// "HH:MM:SS" or "HH:MM" or time.Time (note that only the hours, minutes,
// seconds and nanos are used).
func (s *ScheduledTask) At(i interface{}) *ScheduledTask {

	switch t := i.(type) {
	case string:
		for _, tt := range strings.Split(t, ";") {
			hour, min, sec, err := parseTime(tt)
			if err != nil {
				schedLogger.Error("Failed to parse the specified time string!",
					zap.Error(err),
				)
				s.error = wrapOrError(s.error, err)
				return s
			}
			// save atTime start as duration from midnight
			s.addRunAtTime(time.Duration(hour)*time.Hour +
				time.Duration(min)*time.Minute +
				time.Duration(sec)*time.Second)
		}

	case time.Time:
		s.addRunAtTime(time.Duration(t.Hour())*time.Hour +
			time.Duration(t.Minute())*time.Minute +
			time.Duration(t.Second())*time.Second +
			time.Duration(t.Nanosecond())*time.Nanosecond)

	default:
		schedLogger.Error("Unsupported time format specified!")
		s.error = wrapOrError(s.error, ErrUnsupportedTimeFormat)
	}
	return s
}

// Now schedules the task to be run 'now' - i.e. there is no schedule for the
// task and it is a one-time task.
func (s *ScheduledTask) Now() *ScheduledTask {
	s.setSchedulingUnit(common.Once)
	return s.At(time.Now())
}

// StartAt schedules the next run of the task. If this time is in the past,
// the configured interval will be used to calculate the next future time
func (s *ScheduledTask) StartAt(t time.Time) *ScheduledTask {
	s.TaskInfo.StartAt = t
	s.TaskInfo.StartImmediately = false
	return s
}

// Every schedules a new periodic task with an interval.
// Interval can be an int, time.Duration or a string that parses with
// time.ParseDuration().
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func (s *ScheduledTask) Every(interval interface{}) *ScheduledTask {
	switch interval := interval.(type) {
	case int:
		s.TaskInfo.Interval = interval
		if interval <= 0 {
			schedLogger.Error("Invalid time interval specified!",
				zap.Int("Specified interval", interval),
			)
			s.error = wrapOrError(s.error, ErrInvalidInterval)
		}

	case time.Duration:
		s.TaskInfo.Interval = 0
		s.TaskInfo.Duration = interval
		s.setSchedulingUnit(common.Duration)

	case string:
		d, err := time.ParseDuration(interval)
		if err != nil {
			schedLogger.Error("Failed to parse the duration specified from the interval string!",
				zap.String("Specified interval", interval),
				zap.Error(err),
			)
			s.error = wrapOrError(s.error, err)
		}
		s.TaskInfo.Duration = d
		s.setSchedulingUnit(common.Duration)

	default:
		schedLogger.Error("Invalid interval type specified!")
		s.error = wrapOrError(s.error, ErrInvalidIntervalType)
	}

	return s
}

// Cron - specifies a cron formatted schedule for the task.
func (s *ScheduledTask) Cron(cronExpression string,
	withSeconds bool) *ScheduledTask {
	var (
		withLocation string
		cronSchedule cron.Schedule
		err          error
	)

	// Parse the specified cron expression.
	if strings.HasPrefix(cronExpression, "TZ=") ||
		strings.HasPrefix(cronExpression, "CRON_TZ=") {
		withLocation = cronExpression
	} else {
		// Specified cron expression doesn't have location information. Use the
		// default location information for the task.
		withLocation = fmt.Sprintf("CRON_TZ=%s %s", s.location.String(),
			cronExpression)
	}

	if withSeconds {
		p := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom |
			cron.Month | cron.Dow | cron.Descriptor)
		cronSchedule, err = p.Parse(withLocation)
	} else {
		cronSchedule, err = cron.ParseStandard(withLocation)
	}

	if err != nil {
		s.error = wrapOrError(err, ErrCronParseFailure)
	}
	s.cronSchedule = cronSchedule
	s.TaskInfo.Unit = common.Crontab
	s.TaskInfo.StartImmediately = false

	return s
}

// Schedule - requests the scheduler to schedule the task.
func (s *ScheduledTask) Schedule() (*ScheduledTask, error) {
	var err error

	// Perform some validation checks on the type of scheduled task being
	// submitted to the scheduler for execution.
	if (len(s.TaskInfo.ScheduledWeekdays) != 0) &&
		(s.TaskInfo.Unit != common.Weeks) {
		s.error = wrapOrError(s.error, ErrWeekdayNotSupported)
	}

	if s.TaskInfo.Unit != common.Crontab && s.TaskInfo.Interval == 0 {
		if (s.TaskInfo.Unit != common.Duration) &&
			(s.TaskInfo.Unit != common.Once) {
			s.error = wrapOrError(s.error, ErrInvalidInterval)
		}
	}

	// If the prior steps while building the scheduled task resulted in errors,
	// fail scheduling.
	if s.error != nil {
		schedLogger.Error("Failed to schedule the task!",
			zap.Error(s.error),
		)
		return nil, s.error
	}

	// Create a task and store it in the scheduler database.
	err = s.TaskInfo.CreateTask()
	if err != nil {
		schedLogger.Error("Failed to store a task in the scheduler database!",
			zap.String("Tenant ID", s.TaskInfo.TenantID),
			zap.String("Device ID", s.TaskInfo.DeviceID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// If this is a one time task execution request, send it to the dispatch
	// queue for delivery to the device right away.
	if s.TaskInfo.Unit == common.Once {
		// Encode the task information to prepare for posting to the
		// dispatch queue.
		payload, err := s.TaskInfo.MarshalServiceMessage()
		if err != nil {
			schedLogger.Error("Failed to encode the message for delivery to the device!",
				zap.String("Task ID", s.TaskInfo.TaskID.String()),
				zap.Error(err),
			)
			return nil, err
		}

		// Send the task to the dispatch queue. Tasks on the dispatch
		// queue are sent to the MQTT broker for delivery to the device.
		err = queuemgr.Provider.SendDispatchQueueMessage(payload)
		if err != nil {
			schedLogger.Error("Failed to dispatch the scheduled task to the device!",
				zap.String("Task ID", s.TaskInfo.TaskID.String()),
				zap.String("Device ID", s.TaskInfo.DeviceID.String()),
				zap.Error(err),
			)
			return nil, err
		}
	} else {
		// Create a scheduled run for the task and store it in the database.
		s.ScheduleInfo.TaskID = s.TaskInfo.TaskID
		s.ScheduleInfo.NextRun = time.Now()

		err = s.ScheduleInfo.CreateScheduledRun()
		if err != nil {
			schedLogger.Error("Failed to store a scheduled run in the scheduler database!",
				zap.String("Task ID", s.TaskInfo.TaskID.String()),
				zap.String("Tenant ID", s.TaskInfo.TenantID),
				zap.String("Device ID", s.TaskInfo.DeviceID.String()),
				zap.Error(err),
			)
			return nil, err
		}
	}

	return s, nil
}
