package scheduler

import (
	"strconv"
	"strings"
	"time"

	"github.com/hpinc/krypton-scheduler/service/common"
	"go.uber.org/zap"
)

var (
	// A lookup table used to find the appropriate task scheduling function for the
	// requested scheduling unit string.
	schedulingUnitLookupTable = map[string]func(*ScheduledTask) *ScheduledTask{
		"millisecond":  (*ScheduledTask).Millisecond,
		"milliseconds": (*ScheduledTask).Milliseconds,
		"second":       (*ScheduledTask).Second,
		"seconds":      (*ScheduledTask).Seconds,
		"minute":       (*ScheduledTask).Minute,
		"minutes":      (*ScheduledTask).Minutes,
		"hour":         (*ScheduledTask).Hour,
		"hours":        (*ScheduledTask).Hours,
		"day":          (*ScheduledTask).Day,
		"days":         (*ScheduledTask).Days,
		"week":         (*ScheduledTask).Week,
		"weeks":        (*ScheduledTask).Weeks,
		"monday":       (*ScheduledTask).Monday,
		"tuesday":      (*ScheduledTask).Tuesday,
		"wednesday":    (*ScheduledTask).Wednesday,
		"thursday":     (*ScheduledTask).Thursday,
		"friday":       (*ScheduledTask).Friday,
		"saturday":     (*ScheduledTask).Saturday,
		"sunday":       (*ScheduledTask).Sunday,
		"midday":       (*ScheduledTask).Midday,
	}
)

// Sample schedule strings:
// - Every 2h
// - Every 1 day
// - Every 1 week
func (s *ScheduledTask) ParseSchedule(scheduleStr string) *ScheduledTask {

	// If the schedule is not specified, default to scheduling the task
	// immediately.
	if (scheduleStr == "") ||
		(strings.ToLower(scheduleStr) == common.SchedulingFrequencyNow) {
		s.TaskInfo.Unit = common.Once
		s.TaskInfo.StartAt = time.Now()
		s.TaskInfo.StartImmediately = true
		return s.Now()
	}

	// Parse the requested task schedule so we can determine how to send it to
	// the scheduler.
	parsedSchedule := strings.Split(scheduleStr, " ")
	if len(parsedSchedule) < 2 {
		schedLogger.Error("Invalid task schedule specified",
			zap.String("Schedule string: ", scheduleStr),
		)
		s.error = wrapOrError(s.error, ErrInvalidScheduleType)
		return s
	}

	switch strings.ToLower(parsedSchedule[0]) {

	// Tasks performed periodically at the specified interval.
	case common.SchedulingFrequencyEvery:
		// First try to parse as an integer to figure out if the interval is
		// being set. If this fails, fall back to treating it as a duration
		// string.
		duration, err := strconv.Atoi(parsedSchedule[1])
		if err == nil {
			s = s.Every(duration)
		} else {
			s = s.Every(parsedSchedule[1])
		}

		// Determine if the caller specified a scheduling unit along with the
		// interval.
		if len(parsedSchedule) > 2 {
			unit := strings.ToLower(parsedSchedule[2])
			switch unit {
			case "monthdays":
				// Parse the Days of the month specified.
				if len(parsedSchedule) < 4 {
					schedLogger.Error("Invalid task schedule specified for days of the month",
						zap.String("Schedule string", scheduleStr),
					)
					s.error = wrapOrError(s.error, ErrInvalidScheduleType)
					return s
				}
				s = s.Month(s.ParseDaysOfMonth(parsedSchedule[3])...)

			default:
				schedulingUnitFn, ok := schedulingUnitLookupTable[unit]
				if !ok {
					schedLogger.Error("Invalid scheduling unit requested!",
						zap.String("Requested unit", parsedSchedule[2]),
					)
					s.error = wrapOrError(s.error, ErrInvalidSchedulingUnit)
					return s
				}
				s = schedulingUnitFn(s)
			}
		}

	// Tasks performed at a specified time of day - specified in "HH:MM:SS" or
	// "HH:MM" or time.Time format.
	case common.SchedulingFrequencyAt:
		s = s.At(parsedSchedule[1])

	// Tasks performed at a requested cron schedule time.
	case common.SchedulingFrequencyCron:
		s = s.Cron(parsedSchedule[1], true)

	default:
		s.error = wrapOrError(s.error, ErrInvalidScheduleType)
	}
	return s
}

// setSchedulingUnit sets the type of scheduling unit for the task.
func (s *ScheduledTask) setSchedulingUnit(unit common.SchedulingUnit) {
	currentUnit := s.TaskInfo.Unit
	if currentUnit == common.Duration || currentUnit == common.Crontab {
		s.error = wrapOrError(s.error, ErrInvalidIntervalUnitsSelection)
		return
	}
	s.TaskInfo.Unit = unit
}

// Millisecond sets the scheduling unit to milliseconds
func (s *ScheduledTask) Millisecond() *ScheduledTask {
	return s.Milliseconds()
}

// Milliseconds sets the scheduling unit to milliseconds
func (s *ScheduledTask) Milliseconds() *ScheduledTask {
	s.setSchedulingUnit(common.Milliseconds)
	return s
}

// Second sets the scheduling unit to seconds
func (s *ScheduledTask) Second() *ScheduledTask {
	return s.Seconds()
}

// Seconds sets the scheduling unit to seconds
func (s *ScheduledTask) Seconds() *ScheduledTask {
	s.setSchedulingUnit(common.Seconds)
	return s
}

// Minute sets the scheduling unit to minutes
func (s *ScheduledTask) Minute() *ScheduledTask {
	return s.Minutes()
}

// Minutes sets the scheduling unit to minutes
func (s *ScheduledTask) Minutes() *ScheduledTask {
	s.setSchedulingUnit(common.Minutes)
	return s
}

// Hour sets the scheduling unit to hours
func (s *ScheduledTask) Hour() *ScheduledTask {
	return s.Hours()
}

// Hours sets the scheduling unit to hours
func (s *ScheduledTask) Hours() *ScheduledTask {
	s.setSchedulingUnit(common.Hours)
	return s
}

// Day sets the scheduling unit to days
func (s *ScheduledTask) Day() *ScheduledTask {
	s.setSchedulingUnit(common.Days)
	return s
}

// Days sets the scheduling unit to days
func (s *ScheduledTask) Days() *ScheduledTask {
	s.setSchedulingUnit(common.Days)
	return s
}

// Week sets the scheduling unit to weeks
func (s *ScheduledTask) Week() *ScheduledTask {
	s.setSchedulingUnit(common.Weeks)
	return s
}

// Weeks sets the scheduling unit to weeks
func (s *ScheduledTask) Weeks() *ScheduledTask {
	s.setSchedulingUnit(common.Weeks)
	return s
}

// Month sets the scheduling unit to months
func (s *ScheduledTask) Month(daysOfMonth ...int) *ScheduledTask {
	return s.Months(daysOfMonth...)
}

// MonthLastDay sets the unit with months at every last day of the month
func (s *ScheduledTask) MonthLastDay() *ScheduledTask {
	return s.Months(-1)
}

// Parse the days of the month from the specified comma separated string.
func (s *ScheduledTask) ParseDaysOfMonth(days string) []int {
	var err error
	strs := strings.Split(days, ",")
	res := make([]int, len(strs))
	for i := range res {
		res[i], err = strconv.Atoi(strs[i])
		if err != nil {
			s.error = wrapOrError(s.error, err)
			return nil
		}
	}
	return res
}

// Months sets the scheduling unit to months
// Note: Only days 1 through 28 are allowed for monthly schedules
// Note: Multiple add same days of month cannot be allowed
// Note: -1 is a special value and can only occur as single argument
func (s *ScheduledTask) Months(daysOfTheMonth ...int) *ScheduledTask {

	if len(daysOfTheMonth) == 0 {
		s.error = wrapOrError(s.error, ErrInvalidDayOfMonthEntry)
	} else if len(daysOfTheMonth) == 1 {
		dayOfMonth := daysOfTheMonth[0]
		if dayOfMonth != -1 && (dayOfMonth < 1 || dayOfMonth > 28) {
			s.error = wrapOrError(s.error, ErrInvalidDayOfMonthEntry)
		}
	} else {

		repeatMap := make(map[int]int)
		for _, dayOfMonth := range daysOfTheMonth {

			if dayOfMonth < 1 || dayOfMonth > 28 {
				s.error = wrapOrError(s.error, ErrInvalidDayOfMonthEntry)
				break
			}

			for _, dayOfMonthInTask := range s.TaskInfo.ScheduledDaysOfTheMonth {
				if dayOfMonthInTask == dayOfMonth {
					s.error = wrapOrError(s.error,
						ErrInvalidDaysOfMonthDuplicateValue)
					break
				}
			}

			if _, ok := repeatMap[dayOfMonth]; ok {
				s.error = wrapOrError(s.error, ErrInvalidDaysOfMonthDuplicateValue)
				break
			} else {
				repeatMap[dayOfMonth]++
			}
		}
	}
	if s.TaskInfo.ScheduledDaysOfTheMonth == nil {
		s.TaskInfo.ScheduledDaysOfTheMonth = make([]int, 0)
	}
	s.TaskInfo.ScheduledDaysOfTheMonth = append(s.TaskInfo.ScheduledDaysOfTheMonth,
		daysOfTheMonth...)
	s.TaskInfo.StartImmediately = false
	s.setSchedulingUnit(common.Months)
	return s
}

// NOTE: If the dayOfTheMonth for the above two functions is
// more than the number of days in that month, the extra day(s)
// spill over to the next month. Similarly, if it's less than 0,
// it will go back to the month before

// Adds a specified weekday to the list of scheduledWeekdays.
func (s *ScheduledTask) Weekday(weekDay time.Weekday) *ScheduledTask {
	isScheduledDay := in(s.TaskInfo.ScheduledWeekdays, weekDay)
	if !isScheduledDay {
		s.TaskInfo.ScheduledWeekdays = append(s.TaskInfo.ScheduledWeekdays,
			weekDay)
	}

	s.TaskInfo.StartImmediately = false
	s.setSchedulingUnit(common.Weeks)
	return s
}

func (s *ScheduledTask) Midday() *ScheduledTask {
	return s.At("12:00")
}

// Monday sets the start day as Monday
func (s *ScheduledTask) Monday() *ScheduledTask {
	return s.Weekday(time.Monday)
}

// Tuesday sets the start day as Tuesday
func (s *ScheduledTask) Tuesday() *ScheduledTask {
	return s.Weekday(time.Tuesday)
}

// Wednesday sets the start day as Wednesday
func (s *ScheduledTask) Wednesday() *ScheduledTask {
	return s.Weekday(time.Wednesday)
}

// Thursday sets the start day as Thursday
func (s *ScheduledTask) Thursday() *ScheduledTask {
	return s.Weekday(time.Thursday)
}

// Friday sets the start day as Friday
func (s *ScheduledTask) Friday() *ScheduledTask {
	return s.Weekday(time.Friday)
}

// Saturday sets the start day as Saturday
func (s *ScheduledTask) Saturday() *ScheduledTask {
	return s.Weekday(time.Saturday)
}

// Sunday sets the start day as Sunday
func (s *ScheduledTask) Sunday() *ScheduledTask {
	return s.Weekday(time.Sunday)
}
