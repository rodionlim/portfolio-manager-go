package scheduler

import (
	"context"
	"sync"
	"time"

	"portfolio-manager/pkg/logging"

	"github.com/google/uuid"
)

// ScheduledTask contains information about a scheduled task
type ScheduledTask struct {
	ID       TaskID
	Task     Task
	Schedule Schedule
	NextRun  time.Time
}

// DefaultScheduler implements the Scheduler interface
type DefaultScheduler struct {
	tasks    map[TaskID]*ScheduledTask
	mu       sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup
	logger   *logging.Logger
}

// NewScheduler creates a new scheduler
func NewScheduler() *DefaultScheduler {
	return &DefaultScheduler{
		tasks:    make(map[TaskID]*ScheduledTask),
		stopChan: make(chan struct{}),
		logger:   logging.GetLogger(),
	}
}

// ScheduleTask implements Scheduler.ScheduleTask
func (s *DefaultScheduler) ScheduleTask(task Task, schedule Schedule) TaskID {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := TaskID(uuid.New().String())
	nextRun, hasNext := schedule.Next(time.Now())
	if !hasNext {
		s.logger.Warn("Task schedule has no next execution time")
		return ""
	}

	s.tasks[id] = &ScheduledTask{
		ID:       id,
		Task:     task,
		Schedule: schedule,
		NextRun:  nextRun,
	}

	s.logger.Infof("Task %s scheduled for %s", id, nextRun.Format(time.RFC3339))
	return id
}

// ScheduleTaskFunc implements Scheduler.ScheduleTaskFunc
func (s *DefaultScheduler) ScheduleTaskFunc(taskFunc TaskFunc, schedule Schedule) TaskID {
	return s.ScheduleTask(taskFunc, schedule)
}

// ScheduleOnce implements Scheduler.ScheduleOnce
func (s *DefaultScheduler) ScheduleOnce(task Task, delay time.Duration) TaskID {
	return s.ScheduleTask(task, NewDelaySchedule(delay))
}

// ScheduleOnceFunc implements Scheduler.ScheduleOnceFunc
func (s *DefaultScheduler) ScheduleOnceFunc(taskFunc TaskFunc, delay time.Duration) TaskID {
	return s.ScheduleTask(taskFunc, NewDelaySchedule(delay))
}

// Unschedule implements Scheduler.Unschedule
func (s *DefaultScheduler) Unschedule(taskID TaskID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[taskID]; exists {
		delete(s.tasks, taskID)
		s.logger.Infof("Task %s unscheduled", taskID)
		return true
	}
	return false
}

// Start implements Scheduler.Start
func (s *DefaultScheduler) Start(ctx context.Context) error {
	s.logger.Info("Starting scheduler")

	s.wg.Add(1)
	go s.run(ctx)

	return nil
}

// Stop implements Scheduler.Stop
func (s *DefaultScheduler) Stop() error {
	s.logger.Info("Stopping scheduler")
	close(s.stopChan)
	s.wg.Wait()
	return nil
}

// run is the main scheduler loop
func (s *DefaultScheduler) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAndRunTasks(ctx)
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		}
	}
}

// checkAndRunTasks checks for tasks that need to be executed and runs them
func (s *DefaultScheduler) checkAndRunTasks(ctx context.Context) {
	now := time.Now()
	var tasksToRun []*ScheduledTask

	// First, identify tasks that need to run under a read lock
	s.mu.RLock()
	for _, task := range s.tasks {
		if !task.NextRun.After(now) {
			tasksToRun = append(tasksToRun, task)
		}
	}
	s.mu.RUnlock()

	// Run tasks and update next run times
	for _, task := range tasksToRun {
		// Execute the task in a goroutine to avoid blocking the scheduler
		s.wg.Add(1)
		taskCopy := task // Create a copy to avoid closure issues
		go func() {
			defer s.wg.Done()
			s.executeTask(ctx, taskCopy)
		}()

		// Update the next run time
		s.mu.Lock()
		// Check if the task still exists (wasn't unscheduled while executing)
		if _, exists := s.tasks[task.ID]; exists {
			nextRun, hasNext := task.Schedule.Next(now)
			if hasNext {
				task.NextRun = nextRun
				s.logger.Infof("Next execution of task %s scheduled for %s", task.ID, nextRun.Format(time.RFC3339))
			} else {
				// Remove tasks with no more executions
				delete(s.tasks, task.ID)
				s.logger.Infof("Task %s completed its schedule and was removed", task.ID)
			}
		}
		s.mu.Unlock()
	}
}

// executeTask executes a task and handles any errors
func (s *DefaultScheduler) executeTask(ctx context.Context, task *ScheduledTask) {
	taskCtx, cancel := context.WithTimeout(ctx, 5*time.Minute) // Set a reasonable timeout
	defer cancel()

	s.logger.Infof("Executing task %s", task.ID)
	startTime := time.Now()

	err := task.Task.Run(taskCtx)
	duration := time.Since(startTime)

	if err != nil {
		s.logger.Errorf("Task %s failed after %v: %v", task.ID, duration, err)
	} else {
		s.logger.Infof("Task %s completed successfully in %v", task.ID, duration)
	}
}
