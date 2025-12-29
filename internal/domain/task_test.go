package domain

import (
	"testing"
	"time"
)

func TestTask_DaysUntilDeadline(t *testing.T) {
	tests := []struct {
		name     string
		deadline time.Time
		want     int
	}{
		{
			name:     "deadline today",
			deadline: time.Now().Truncate(24 * time.Hour),
			want:     0,
		},
		{
			name:     "deadline tomorrow",
			deadline: time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour),
			want:     1,
		},
		{
			name:     "deadline in 5 days",
			deadline: time.Now().Truncate(24 * time.Hour).Add(5 * 24 * time.Hour),
			want:     5,
		},
		{
			name:     "deadline yesterday",
			deadline: time.Now().Truncate(24 * time.Hour).Add(-24 * time.Hour),
			want:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{Deadline: tt.deadline}
			got := task.DaysUntilDeadline()
			if got != tt.want {
				t.Errorf("DaysUntilDeadline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_WorkHoursRemaining(t *testing.T) {
	tests := []struct {
		name            string
		daysUntil       int
		workHoursPerDay int
		want            int
	}{
		{
			name:            "3 days with 8 hours",
			daysUntil:       3,
			workHoursPerDay: 8,
			want:            24,
		},
		{
			name:            "5 days with 6 hours",
			daysUntil:       5,
			workHoursPerDay: 6,
			want:            30,
		},
		{
			name:            "0 days",
			daysUntil:       0,
			workHoursPerDay: 8,
			want:            0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deadline := time.Now().Truncate(24 * time.Hour).Add(time.Duration(tt.daysUntil) * 24 * time.Hour)
			task := &Task{Deadline: deadline}
			got := task.WorkHoursRemaining(tt.workHoursPerDay)
			if got != tt.want {
				t.Errorf("WorkHoursRemaining() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_ShouldRemindToday(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	tests := []struct {
		name      string
		task      *Task
		want      bool
	}{
		{
			name: "completed task",
			task: &Task{
				IsCompleted: true,
				Deadline:    tomorrow,
				Frequency:   FrequencyDaily,
			},
			want: false,
		},
		{
			name: "deadline passed",
			task: &Task{
				IsCompleted: false,
				Deadline:    yesterday,
				Frequency:   FrequencyDaily,
			},
			want: false,
		},
		{
			name: "daily frequency - should remind",
			task: &Task{
				IsCompleted: false,
				Deadline:    tomorrow,
				Frequency:   FrequencyDaily,
			},
			want: true,
		},
		{
			name: "weekly frequency - same weekday as deadline",
			task: &Task{
				IsCompleted: false,
				Deadline:    today.Add(7 * 24 * time.Hour), // next week same day
				Frequency:   FrequencyWeekly,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.ShouldRemindToday()
			if got != tt.want {
				t.Errorf("ShouldRemindToday() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_CanSendReminder(t *testing.T) {
	tomorrow := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)

	tests := []struct {
		name string
		task *Task
		want bool
	}{
		{
			name: "can send - no reminders sent yet",
			task: &Task{
				IsCompleted:        false,
				Deadline:           tomorrow,
				Frequency:          FrequencyDaily,
				Importance:         3,
				RemindersSentToday: 0,
			},
			want: true,
		},
		{
			name: "can send - some reminders sent",
			task: &Task{
				IsCompleted:        false,
				Deadline:           tomorrow,
				Frequency:          FrequencyDaily,
				Importance:         3,
				RemindersSentToday: 2,
			},
			want: true,
		},
		{
			name: "cannot send - all reminders sent",
			task: &Task{
				IsCompleted:        false,
				Deadline:           tomorrow,
				Frequency:          FrequencyDaily,
				Importance:         3,
				RemindersSentToday: 3,
			},
			want: false,
		},
		{
			name: "cannot send - more than importance",
			task: &Task{
				IsCompleted:        false,
				Deadline:           tomorrow,
				Frequency:          FrequencyDaily,
				Importance:         2,
				RemindersSentToday: 5,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.CanSendReminder()
			if got != tt.want {
				t.Errorf("CanSendReminder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_ImportanceStars(t *testing.T) {
	tests := []struct {
		importance int
		want       string
	}{
		{1, "★☆☆☆☆"},
		{2, "★★☆☆☆"},
		{3, "★★★☆☆"},
		{4, "★★★★☆"},
		{5, "★★★★★"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			task := &Task{Importance: tt.importance}
			got := task.ImportanceStars()
			if got != tt.want {
				t.Errorf("ImportanceStars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTask(t *testing.T) {
	deadline := time.Now().Add(24 * time.Hour)
	task := NewTask(123, "Test task", deadline, 3, FrequencyDaily)

	if task.UserID != 123 {
		t.Errorf("UserID = %v, want 123", task.UserID)
	}
	if task.Description != "Test task" {
		t.Errorf("Description = %v, want 'Test task'", task.Description)
	}
	if task.Importance != 3 {
		t.Errorf("Importance = %v, want 3", task.Importance)
	}
	if task.Frequency != FrequencyDaily {
		t.Errorf("Frequency = %v, want %v", task.Frequency, FrequencyDaily)
	}
	if task.IsCompleted {
		t.Error("IsCompleted should be false")
	}
}
