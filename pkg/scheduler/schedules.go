package scheduler

import (
	"time"
)

// PeriodicSchedule represents a schedule that repeats at fixed intervals
type PeriodicSchedule struct {
	Interval time.Duration
}

// NewPeriodicSchedule creates a new schedule that repeats at fixed intervals
func NewPeriodicSchedule(interval time.Duration) *PeriodicSchedule {
	return &PeriodicSchedule{Interval: interval}
}

// Next implements the Schedule interface
func (s *PeriodicSchedule) Next(after time.Time) (time.Time, bool) {
	return after.Add(s.Interval), true
}

// DailySchedule represents a schedule that runs once per day at a specific time
type DailySchedule struct {
	Hour   int
	Minute int
	Second int
}

// NewDailySchedule creates a new schedule that runs once per day at a specific time
func NewDailySchedule(hour, minute, second int) *DailySchedule {
	return &DailySchedule{
		Hour:   hour,
		Minute: minute,
		Second: second,
	}
}

// Next implements the Schedule interface
func (s *DailySchedule) Next(after time.Time) (time.Time, bool) {
	next := time.Date(
		after.Year(), after.Month(), after.Day(),
		s.Hour, s.Minute, s.Second, 0,
		after.Location(),
	)

	if next.Before(after) || next.Equal(after) {
		// If the time for today has already passed, schedule for tomorrow
		next = next.AddDate(0, 0, 1)
	}

	return next, true
}

// OneTimeSchedule represents a schedule that runs only once at a specific time
type OneTimeSchedule struct {
	Time time.Time
}

// NewOneTimeSchedule creates a new schedule that runs only once at a specific time
func NewOneTimeSchedule(t time.Time) *OneTimeSchedule {
	return &OneTimeSchedule{Time: t}
}

// Next implements the Schedule interface
func (s *OneTimeSchedule) Next(after time.Time) (time.Time, bool) {
	if after.Before(s.Time) {
		return s.Time, true
	}
	return time.Time{}, false // No more executions
}

// DelaySchedule represents a schedule that runs once after a delay
type DelaySchedule struct {
	Delay time.Duration
	start time.Time
	used  bool
}

// NewDelaySchedule creates a new schedule that runs once after a delay
func NewDelaySchedule(delay time.Duration) *DelaySchedule {
	return &DelaySchedule{
		Delay: delay,
		start: time.Now(),
		used:  false,
	}
}

// Next implements the Schedule interface
func (s *DelaySchedule) Next(after time.Time) (time.Time, bool) {
	if s.used {
		return time.Time{}, false
	}

	s.used = true
	return s.start.Add(s.Delay), true
}
