package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

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

// CronSchedule represents a schedule based on a cron expression
// Uses robfig/cron/v3 for parsing and next time calculation
// Only standard 5-field cron expressions are supported
// (minute, hour, day of month, month, day of week)
type CronSchedule struct {
	expr  string
	sched cron.Schedule
}

// NewCronSchedule creates a new schedule from a cron expression (5 fields)
func NewCronSchedule(expr string) (*CronSchedule, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}
	return &CronSchedule{
		expr:  expr,
		sched: sched,
	}, nil
}

// Next implements the Schedule interface for CronSchedule
func (s *CronSchedule) Next(after time.Time) (time.Time, bool) {
	next := s.sched.Next(after)
	if next.IsZero() {
		return time.Time{}, false
	}
	return next, true
}
