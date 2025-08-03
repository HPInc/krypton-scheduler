package scheduler

import (
	"regexp"
	"sort"
	"time"
)

// regex patterns for supported time formats
var (
	timeWithSeconds    = regexp.MustCompile(`(?m)^\d{1,2}:\d\d:\d\d$`)
	timeWithoutSeconds = regexp.MustCompile(`(?m)^\d{1,2}:\d\d$`)
)

// Parse the provided timestring and determine the hour/min/second.
func parseTime(t string) (hour, min, sec int, err error) {
	var timeLayout string
	switch {
	case timeWithSeconds.Match([]byte(t)):
		timeLayout = "15:04:05"
	case timeWithoutSeconds.Match([]byte(t)):
		timeLayout = "15:04"
	default:
		return 0, 0, 0, ErrUnsupportedTimeFormat
	}

	parsedTime, err := time.Parse(timeLayout, t)
	if err != nil {
		return 0, 0, 0, ErrUnsupportedTimeFormat
	}
	return parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), nil
}

// Check if the specified weekday is one of the scheduled week days for the
// task to run.
func in(scheduleWeekdays []time.Weekday, weekday time.Weekday) bool {
	in := false

	for _, weekdayInSchedule := range scheduleWeekdays {
		if int(weekdayInSchedule) == int(weekday) {
			in = true
			break
		}
	}
	return in
}

func (s *ScheduledTask) addRunAtTime(t time.Duration) {
	if len(s.TaskInfo.RunAt) == 0 {
		s.TaskInfo.RunAt = append(s.TaskInfo.RunAt, t)
		return
	}
	exist := false
	index := sort.Search(len(s.TaskInfo.RunAt), func(i int) bool {
		atTime := s.TaskInfo.RunAt[i]
		b := atTime >= t
		if b {
			exist = atTime == t
		}
		return b
	})

	// ignore if present
	if exist {
		return
	}

	s.TaskInfo.RunAt = append(s.TaskInfo.RunAt, time.Duration(0))
	copy(s.TaskInfo.RunAt[index+1:], s.TaskInfo.RunAt[index:])
	s.TaskInfo.RunAt[index] = t
}
