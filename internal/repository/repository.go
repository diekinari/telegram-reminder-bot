package repository

import (
	"context"

	"telegram-reminder-bot/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id int64) (*domain.Task, error)
	GetActiveByUserID(ctx context.Context, userID int64) ([]*domain.Task, error)
	GetTasksForReminder(ctx context.Context) ([]*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id int64) error
	ResetDailyReminders(ctx context.Context) error
}
