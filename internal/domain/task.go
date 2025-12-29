package domain

import (
	"time"
)

type Task struct {
	ID                 int64
	UserID             int64
	Description        string
	Deadline           time.Time
	Importance         int
	Frequency          Frequency
	IsCompleted        bool
	LastReminderDate   *time.Time
	RemindersSentToday int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewTask(userID int64, description string, deadline time.Time, importance int, frequency Frequency) *Task {
	return &Task{
		UserID:      userID,
		Description: description,
		Deadline:    deadline,
		Importance:  importance,
		Frequency:   frequency,
		IsCompleted: false,
	}
}

func (t *Task) DaysUntilDeadline() int {
	now := time.Now().Truncate(24 * time.Hour)
	deadline := t.Deadline.Truncate(24 * time.Hour)
	return int(deadline.Sub(now).Hours() / 24)
}

func (t *Task) WorkHoursRemaining(workHoursPerDay int) int {
	days := t.DaysUntilDeadline()
	if days < 0 {
		return 0
	}
	return days * workHoursPerDay
}

func (t *Task) ShouldRemindToday() bool {
	if t.IsCompleted {
		return false
	}

	today := time.Now().Truncate(24 * time.Hour)
	deadline := t.Deadline.Truncate(24 * time.Hour)

	if today.After(deadline) {
		return false
	}

	switch t.Frequency {
	case FrequencyDaily:
		return true
	case FrequencyEveryOtherDay:
		daysSinceCreation := int(today.Sub(t.CreatedAt.Truncate(24*time.Hour)).Hours() / 24)
		return daysSinceCreation%2 == 0
	case FrequencyWeekly:
		return today.Weekday() == t.Deadline.Weekday()
	default:
		return true
	}
}

func (t *Task) CanSendReminder() bool {
	return t.ShouldRemindToday() && t.RemindersSentToday < t.Importance
}

func (t *Task) ImportanceStars() string {
	stars := ""
	for i := 0; i < 5; i++ {
		if i < t.Importance {
			stars += "★"
		} else {
			stars += "☆"
		}
	}
	return stars
}
