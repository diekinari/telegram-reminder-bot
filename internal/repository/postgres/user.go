package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"telegram-reminder-bot/internal/domain"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (telegram_id, username, timezone, work_hours_per_day, work_start_hour, work_end_hour)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	return r.db.Pool.QueryRow(ctx, query,
		user.TelegramID,
		user.Username,
		user.Timezone,
		user.WorkHoursPerDay,
		user.WorkStartHour,
		user.WorkEndHour,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, telegram_id, username, timezone, work_hours_per_day, work_start_hour, work_end_hour, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.Timezone,
		&user.WorkHoursPerDay,
		&user.WorkStartHour,
		&user.WorkEndHour,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	query := `
		SELECT id, telegram_id, username, timezone, work_hours_per_day, work_start_hour, work_end_hour, created_at, updated_at
		FROM users
		WHERE telegram_id = $1`

	user := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.Timezone,
		&user.WorkHoursPerDay,
		&user.WorkStartHour,
		&user.WorkEndHour,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = $2, timezone = $3, work_hours_per_day = $4, work_start_hour = $5, work_end_hour = $6, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Timezone,
		user.WorkHoursPerDay,
		user.WorkStartHour,
		user.WorkEndHour,
	)
	return err
}
