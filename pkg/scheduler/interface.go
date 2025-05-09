// Package scheduler provides functionality for scheduling tasks to run
// on a recurring or one-time basis.
package scheduler

import (
	"context"
	"time"
)

// TaskID is a unique identifier for a scheduled task
type TaskID string

// Task represents a runnable task in the scheduler
type Task interface {
	// Run executes the task with the provided context
	Run(ctx context.Context) error
}

// TaskFunc is a function type that implements the Task interface
type TaskFunc func(ctx context.Context) error

// Run implements the Task interface for TaskFunc
func (f TaskFunc) Run(ctx context.Context) error {
	return f(ctx)
}

// Schedule defines when and how a task should be executed
type Schedule interface {
	// Next returns the next time the task should be executed
	// after the given time. If no more executions are scheduled,
	// the boolean return value will be false.
	Next(time.Time) (time.Time, bool)
}

// Scheduler manages scheduled tasks
type Scheduler interface {
	// ScheduleTask schedules a task to run according to a given schedule
	// Returns a task ID that can be used to unschedule the task
	ScheduleTask(task Task, schedule Schedule) TaskID

	// ScheduleTaskFunc is a convenience method for scheduling a function
	ScheduleTaskFunc(taskFunc TaskFunc, schedule Schedule) TaskID

	// ScheduleOnce schedules a task to run once after the given delay
	ScheduleOnce(task Task, delay time.Duration) TaskID

	// ScheduleOnceFunc is a convenience method for scheduling a function once
	ScheduleOnceFunc(taskFunc TaskFunc, delay time.Duration) TaskID

	// Unschedule removes a scheduled task by its ID
	Unschedule(taskID TaskID) bool

	// Start starts the scheduler
	Start(ctx context.Context) error

	// Stop gracefully stops the scheduler
	Stop() error
}
