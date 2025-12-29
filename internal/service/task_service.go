package service

import (
	"context"
	"fmt"
	"time"

	"telegram-reminder-bot/internal/domain"
	"telegram-reminder-bot/internal/repository"
)

type TaskService struct {
	taskRepo repository.TaskRepository
}

func NewTaskService(taskRepo repository.TaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo}
}

func (s *TaskService) Create(ctx context.Context, userID int64, description string, deadline time.Time, importance int, frequency domain.Frequency) (*domain.Task, error) {
	if importance < 1 || importance > 5 {
		return nil, fmt.Errorf("importance must be between 1 and 5")
	}

	task := domain.NewTask(userID, description, deadline, importance, frequency)
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	return s.taskRepo.GetByID(ctx, id)
}

func (s *TaskService) GetActiveByUserID(ctx context.Context, userID int64) ([]*domain.Task, error) {
	return s.taskRepo.GetActiveByUserID(ctx, userID)
}

func (s *TaskService) Complete(ctx context.Context, id int64) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}

	task.IsCompleted = true
	return s.taskRepo.Update(ctx, task)
}

func (s *TaskService) Delete(ctx context.Context, id int64) error {
	return s.taskRepo.Delete(ctx, id)
}

func (s *TaskService) GetTasksForReminder(ctx context.Context) ([]*domain.Task, error) {
	return s.taskRepo.GetTasksForReminder(ctx)
}

func (s *TaskService) IncrementReminderCount(ctx context.Context, task *domain.Task) error {
	now := time.Now()
	task.RemindersSentToday++
	task.LastReminderDate = &now
	return s.taskRepo.Update(ctx, task)
}

func (s *TaskService) ResetDailyReminders(ctx context.Context) error {
	return s.taskRepo.ResetDailyReminders(ctx)
}
