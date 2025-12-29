package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"telegram-reminder-bot/internal/domain"
)

type TaskRepository struct {
	db *DB
}

func NewTaskRepository(db *DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (user_id, description, deadline, importance, frequency)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.db.Pool.QueryRow(ctx, query,
		task.UserID,
		task.Description,
		task.Deadline,
		task.Importance,
		task.Frequency,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	query := `
		SELECT id, user_id, description, deadline, importance, frequency, is_completed,
		       last_reminder_date, reminders_sent_today, created_at, updated_at
		FROM tasks
		WHERE id = $1`

	task := &domain.Task{}
	var freq string
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.Description,
		&task.Deadline,
		&task.Importance,
		&freq,
		&task.IsCompleted,
		&task.LastReminderDate,
		&task.RemindersSentToday,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	task.Frequency = domain.Frequency(freq)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *TaskRepository) GetActiveByUserID(ctx context.Context, userID int64) ([]*domain.Task, error) {
	query := `
		SELECT id, user_id, description, deadline, importance, frequency, is_completed,
		       last_reminder_date, reminders_sent_today, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND is_completed = false
		ORDER BY deadline ASC`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		var freq string
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Description,
			&task.Deadline,
			&task.Importance,
			&freq,
			&task.IsCompleted,
			&task.LastReminderDate,
			&task.RemindersSentToday,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		task.Frequency = domain.Frequency(freq)
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *TaskRepository) GetTasksForReminder(ctx context.Context) ([]*domain.Task, error) {
	query := `
		SELECT id, user_id, description, deadline, importance, frequency, is_completed,
		       last_reminder_date, reminders_sent_today, created_at, updated_at
		FROM tasks
		WHERE is_completed = false AND deadline >= CURRENT_DATE`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		var freq string
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Description,
			&task.Deadline,
			&task.Importance,
			&freq,
			&task.IsCompleted,
			&task.LastReminderDate,
			&task.RemindersSentToday,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		task.Frequency = domain.Frequency(freq)
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks
		SET description = $2, deadline = $3, importance = $4, frequency = $5,
		    is_completed = $6, last_reminder_date = $7, reminders_sent_today = $8, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query,
		task.ID,
		task.Description,
		task.Deadline,
		task.Importance,
		task.Frequency,
		task.IsCompleted,
		task.LastReminderDate,
		task.RemindersSentToday,
	)
	return err
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *TaskRepository) ResetDailyReminders(ctx context.Context) error {
	query := `UPDATE tasks SET reminders_sent_today = 0, last_reminder_date = NULL WHERE is_completed = false`
	_, err := r.db.Pool.Exec(ctx, query)
	return err
}
