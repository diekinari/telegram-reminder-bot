package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"

	"telegram-reminder-bot/internal/domain"
	"telegram-reminder-bot/internal/repository"
	"telegram-reminder-bot/internal/service"
)

type ReminderSender interface {
	SendReminder(ctx context.Context, telegramID int64, message string, taskID int64) error
}

type Scheduler struct {
	scheduler   gocron.Scheduler
	taskService *service.TaskService
	userRepo    repository.UserRepository
	sender      ReminderSender
}

func New(taskService *service.TaskService, userRepo repository.UserRepository, sender ReminderSender) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		scheduler:   s,
		taskService: taskService,
		userRepo:    userRepo,
		sender:      sender,
	}, nil
}

func (s *Scheduler) Start(ctx context.Context) error {
	_, err := s.scheduler.NewJob(
		gocron.DurationJob(5*time.Minute),
		gocron.NewTask(func() {
			s.checkReminders(ctx)
		}),
	)
	if err != nil {
		return err
	}

	_, err = s.scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(0, 0, 0))),
		gocron.NewTask(func() {
			s.resetDailyReminders(ctx)
		}),
	)
	if err != nil {
		return err
	}

	s.scheduler.Start()
	log.Info().Msg("scheduler started")

	return nil
}

func (s *Scheduler) Stop() error {
	return s.scheduler.Shutdown()
}

func (s *Scheduler) checkReminders(ctx context.Context) {
	tasks, err := s.taskService.GetTasksForReminder(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get tasks for reminder")
		return
	}

	for _, task := range tasks {
		if !task.CanSendReminder() {
			continue
		}

		user, err := s.userRepo.GetByID(ctx, task.UserID)
		if err != nil || user == nil {
			log.Error().Err(err).Int64("task_id", task.ID).Msg("failed to get user for task")
			continue
		}

		now := time.Now().In(user.Location())

		if !IsWithinWorkHours(user.WorkStartHour, user.WorkEndHour, now) {
			continue
		}

		reminderTimes := CalculateReminderTimes(task.Importance, user.WorkStartHour, user.WorkEndHour, now)

		if !ShouldSendReminder(reminderTimes, task.RemindersSentToday, now) {
			continue
		}

		message := formatReminderMessage(task, user)
		if err := s.sender.SendReminder(ctx, user.TelegramID, message, task.ID); err != nil {
			log.Error().Err(err).Int64("task_id", task.ID).Msg("failed to send reminder")
			continue
		}

		if err := s.taskService.IncrementReminderCount(ctx, task); err != nil {
			log.Error().Err(err).Int64("task_id", task.ID).Msg("failed to increment reminder count")
		}

		log.Info().
			Int64("task_id", task.ID).
			Int64("user_id", user.TelegramID).
			Int("reminder_number", task.RemindersSentToday+1).
			Msg("reminder sent")
	}
}

func (s *Scheduler) resetDailyReminders(ctx context.Context) {
	if err := s.taskService.ResetDailyReminders(ctx); err != nil {
		log.Error().Err(err).Msg("failed to reset daily reminders")
		return
	}
	log.Info().Msg("daily reminders reset")
}

func formatReminderMessage(task *domain.Task, user *domain.User) string {
	days := task.DaysUntilDeadline()
	hours := task.WorkHoursRemaining(user.WorkHoursPerDay)

	daysText := "дней"
	if days == 1 {
		daysText = "день"
	} else if days >= 2 && days <= 4 {
		daysText = "дня"
	}

	hoursText := "часов"
	if hours == 1 {
		hoursText = "час"
	} else if hours >= 2 && hours <= 4 {
		hoursText = "часа"
	}

	reminderNum := task.RemindersSentToday + 1

	return fmt.Sprintf(`🔔 <b>Напоминание</b> (%d/%d за сегодня)

📋 %s

⏰ До дедлайна: <b>%d %s</b>
⏱ Рабочих часов: <b>%d %s</b>
⚡ Важность: %s`,
		reminderNum, task.Importance,
		escapeHTML(task.Description),
		days, daysText,
		hours, hoursText,
		task.ImportanceStars(),
	)
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
