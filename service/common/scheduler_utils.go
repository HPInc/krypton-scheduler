package common

import (
	pb "github.com/hpinc/krypton-scheduler/protos"
)

const (
	// Desired scheduling frequency
	SchedulingFrequencyEvery = "every"
	SchedulingFrequencyAt    = "at"
	SchedulingFrequencyCron  = "cron"
	SchedulingFrequencyNow   = "now"

	// Special device ID used to send out broadcast tasks/messages over
	// the MQTT channel.
	BroadcastDeviceID   = "@all"
	BroadcastDeviceUuid = "0a110a11-bca5-bca5-0a11-87dcb71a7f4d"

	SchedulerRequestSourceEvent = "event"
	SchedulerRequestSourceRest  = "rest"
)

// SchedulingUnit - defines the frequency with which tasks are scheduled.
type SchedulingUnit int

// IMPORTANT - DO NOT REORDER. This value is stored in the database.
const (
	// default unit is once
	Once SchedulingUnit = iota + 1
	Milliseconds
	Seconds
	Minutes
	Hours
	Days
	Weeks
	Months
	Duration
	Crontab
)

// InputEventHandlerFunc - function to process task requests received on the
// scheduler input queue.
type InputEventHandlerFunc func(request *pb.CreateScheduledTaskRequest,
	source string) (*pb.CreateScheduledTaskResponse, error)
