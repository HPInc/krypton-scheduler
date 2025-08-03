package scheduler

import (
	"errors"
	"fmt"
)

// Error declarations for task schedule related errors
var (
	ErrNotScheduledWeekday              = errors.New("task not scheduled weekly on a weekday")
	ErrUnsupportedTimeFormat            = errors.New("the given time format is not supported")
	ErrInvalidScheduleType              = errors.New("schedule type must be either Every or At")
	ErrInvalidSchedulingUnit            = errors.New("invalid scheduling unit requested")
	ErrInvalidInterval                  = errors.New(".Every() interval must be greater than 0")
	ErrInvalidIntervalType              = errors.New(".Every() interval must be of type int, time.Duration, or string")
	ErrInvalidIntervalUnitsSelection    = errors.New(".Every(time.Duration) and .Cron() cannot be used with units (e.g. .Seconds())")
	ErrAtTimeNotSupported               = errors.New("the At() method is not supported for this time unit")
	ErrWeekdayNotSupported              = errors.New("weekday is not supported for time unit")
	ErrInvalidDayOfMonthEntry           = errors.New("only days 1 through 28 are allowed for monthly schedules")
	ErrCronParseFailure                 = errors.New("specified cron expression could not be parsed")
	ErrInvalidDaysOfMonthDuplicateValue = errors.New("duplicate days of month is not allowed in Month() and Months() methods")
	ErrInvalidTenantID                  = errors.New("the specified tenant ID is invalid")
	ErrInvalidServiceID                 = errors.New("the specified service ID is invalid")
	ErrInvalidMessageType               = errors.New("the specified message type is invalid")
	ErrInvalidRequest                   = errors.New("invalid request")
)

// Wrap the existing error or set to the specified error.
func wrapOrError(toWrap error, err error) error {
	if toWrap != nil && !errors.Is(err, toWrap) {
		return fmt.Errorf("%s: %w", err, toWrap)
	} else {
		return err
	}
}
