package historical

import (
	"fmt"
	"portfolio-manager/internal/config"
	"portfolio-manager/internal/mocks"
	"portfolio-manager/internal/mocks/testify"
	"portfolio-manager/pkg/scheduler"
	"portfolio-manager/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateMetricsJob(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	mockScheduler := scheduler.NewScheduler()
	mockMetrics := new(testify.MockMetricsService)

	service := &Service{
		metricsService:  mockMetrics,
		db:              mockDB,
		scheduler:       mockScheduler,
		collectionTasks: []scheduler.TaskID{},
	}

	t.Run("CreateMetricsJob with valid parameters", func(t *testing.T) {
		bookFilter := "book1"
		cronExpr := "0 10 * * *" // Daily at 10 AM

		expectedKey := fmt.Sprintf("%s:%s:%s", types.ScheduledJobKeyPrefix, types.CustomMetricsJobKeyPrefix, bookFilter)

		// Mock database expectations
		mockDB.On("Get", expectedKey, mock.AnythingOfType("*historical.MetricsJob")).Return(assert.AnError).Once() // Job doesn't exist
		mockDB.On("Put", expectedKey, mock.AnythingOfType("historical.MetricsJob")).Return(nil).Once()

		// Execute
		job, err := service.CreateMetricsJob(cronExpr, bookFilter)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, job)
		assert.Len(t, service.collectionTasks, 1)

		// Cleanup
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateMetricsJob with empty book filter should fail", func(t *testing.T) {
		cronExpr := "0 10 * * *"
		bookFilter := ""

		// Execute
		cancelFunc, err := service.CreateMetricsJob(cronExpr, bookFilter)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cancelFunc)
		assert.Contains(t, err.Error(), "book_filter cannot be empty")
	})
}

func TestDeleteMetricsJob(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	mockScheduler := scheduler.NewScheduler()
	mockMetrics := new(testify.MockMetricsService)

	service := &Service{
		metricsService:  mockMetrics,
		db:              mockDB,
		scheduler:       mockScheduler,
		collectionTasks: []scheduler.TaskID{"task1"},
	}

	t.Run("DeleteMetricsJob with empty book filter should fail", func(t *testing.T) {
		bookFilter := ""

		// Execute
		err := service.DeleteMetricsJob(bookFilter)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "book_filter cannot be empty")
	})

	t.Run("DeleteMetricsJob with non-existent job should fail", func(t *testing.T) {
		bookFilter := "nonexistent"
		expectedKey := fmt.Sprintf("%s:%s:%s", types.ScheduledJobKeyPrefix, types.CustomMetricsJobKeyPrefix, bookFilter)

		// Mock database expectations
		mockDB.On("Get", expectedKey, mock.AnythingOfType("*historical.MetricsJob")).Return(assert.AnError).Once()

		// Execute
		err := service.DeleteMetricsJob(bookFilter)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metrics job not found")

		mockDB.AssertExpectations(t)
	})
}

func TestListMetricsJobs(t *testing.T) {
	mockDB := new(mocks.MockDatabase)
	mockScheduler := scheduler.NewScheduler()
	mockMetrics := new(testify.MockMetricsService)

	service := &Service{
		metricsService:  mockMetrics,
		db:              mockDB,
		scheduler:       mockScheduler,
		collectionTasks: []scheduler.TaskID{},
	}

	t.Run("ListMetricsJobs returns custom jobs only", func(t *testing.T) {
		prefix := fmt.Sprintf("SCHEDULED_JOB:%s", types.CustomMetricsJobKeyPrefix)
		keys := []string{
			prefix + ":book1",
			prefix + ":book2",
			prefix + ":portfolio", // This should be excluded
		}

		job1 := MetricsJob{BookFilter: "book1", CronExpr: "0 10 * * *", TaskId: "task1"}
		job2 := MetricsJob{BookFilter: "book2", CronExpr: "0 11 * * *", TaskId: "task2"}
		portfolioJob := MetricsJob{BookFilter: "portfolio", CronExpr: "0 12 * * *", TaskId: "task3"}

		// Mock expectations
		mockDB.On("GetAllKeysWithPrefix", prefix).Return(keys, nil).Once()
		mockDB.On("Get", keys[0], mock.AnythingOfType("*historical.MetricsJob")).Run(func(args mock.Arguments) {
			arg := args.Get(1).(*MetricsJob)
			*arg = job1
		}).Return(nil).Once()
		mockDB.On("Get", keys[1], mock.AnythingOfType("*historical.MetricsJob")).Run(func(args mock.Arguments) {
			arg := args.Get(1).(*MetricsJob)
			*arg = job2
		}).Return(nil).Once()
		mockDB.On("Get", keys[2], mock.AnythingOfType("*historical.MetricsJob")).Run(func(args mock.Arguments) {
			arg := args.Get(1).(*MetricsJob)
			*arg = portfolioJob
		}).Return(nil).Once()

		// Execute
		jobs, err := service.ListMetricsJobs()

		// Assert
		assert.NoError(t, err)
		assert.Len(t, jobs, 2) // Should exclude the portfolio job
		assert.Equal(t, "book1", jobs[0].BookFilter)
		assert.Equal(t, "book2", jobs[1].BookFilter)

		// Verify mocks
		mockDB.AssertExpectations(t)
	})
}

func TestDefaultMetricsScheduleVariable(t *testing.T) {
	t.Run("DefaultMetricsSchedule works with custom schedule in CreateMetricsJob", func(t *testing.T) {
		// Set up a test config with a custom schedule
		config.DefaultMetricsSchedule = "0 9 * * 1-5"

		mockDB := new(mocks.MockDatabase)
		mockScheduler := scheduler.NewScheduler()
		mockMetrics := new(testify.MockMetricsService)

		service := &Service{
			metricsService:  mockMetrics,
			db:              mockDB,
			scheduler:       mockScheduler,
			collectionTasks: []scheduler.TaskID{},
		}

		bookFilter := "book1"
		cronExpr := "" // Empty cron expression should use DefaultMetricsSchedule

		// Mock database expectations
		mockDB.On("Get", mock.AnythingOfType("string"), mock.AnythingOfType("*historical.MetricsJob")).Return(assert.AnError).Once() // Job doesn't exist
		mockDB.On("Put", mock.AnythingOfType("string"), mock.AnythingOfType("historical.MetricsJob")).Return(nil).Once()

		// Execute - this should use DefaultMetricsSchedule since cronExpr is empty
		job, err := service.CreateMetricsJob(cronExpr, bookFilter)

		// Assert - If it did not use DefaultMetricsSchedule, it would have returned an error
		assert.NoError(t, err)
		assert.NotNil(t, job)

		// Cleanup
		mockDB.AssertExpectations(t)
	})

	t.Run("DefaultMetricsSchedule fails when both cronExpr and default are empty", func(t *testing.T) {
		// Set DefaultMetricsSchedule to empty
		config.DefaultMetricsSchedule = ""

		mockDB := new(mocks.MockDatabase)
		mockScheduler := scheduler.NewScheduler()
		mockMetrics := new(testify.MockMetricsService)

		service := &Service{
			metricsService:  mockMetrics,
			db:              mockDB,
			scheduler:       mockScheduler,
			collectionTasks: []scheduler.TaskID{},
		}

		bookFilter := "book1"
		cronExpr := "" // Empty cron expression

		// Execute - this should fail since both cronExpr and DefaultMetricsSchedule are empty
		cancelFunc, err := service.CreateMetricsJob(cronExpr, bookFilter)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cancelFunc)
		assert.Contains(t, err.Error(), "no cron expression provided and no default schedule configured")
	})
}
