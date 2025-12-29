package scheduler

import (
	"testing"
	"time"
)

func TestCalculateReminderTimes(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		importance    int
		workStartHour int
		workEndHour   int
		wantCount     int
	}{
		{
			name:          "importance 1 - single reminder",
			importance:    1,
			workStartHour: 9,
			workEndHour:   18,
			wantCount:     1,
		},
		{
			name:          "importance 3 - three reminders",
			importance:    3,
			workStartHour: 9,
			workEndHour:   18,
			wantCount:     3,
		},
		{
			name:          "importance 5 - five reminders",
			importance:    5,
			workStartHour: 9,
			workEndHour:   18,
			wantCount:     5,
		},
		{
			name:          "importance 0 - no reminders",
			importance:    0,
			workStartHour: 9,
			workEndHour:   18,
			wantCount:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			times := CalculateReminderTimes(tt.importance, tt.workStartHour, tt.workEndHour, now)
			if len(times) != tt.wantCount {
				t.Errorf("CalculateReminderTimes() returned %d times, want %d", len(times), tt.wantCount)
			}
		})
	}
}

func TestCalculateReminderTimes_Distribution(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	t.Run("importance 1 should be at midday", func(t *testing.T) {
		times := CalculateReminderTimes(1, 9, 18, now)
		if len(times) != 1 {
			t.Fatalf("expected 1 time, got %d", len(times))
		}
		// Should be at 13:30 (middle of 9-18)
		expected := time.Date(2024, 1, 15, 13, 30, 0, 0, time.UTC)
		if !times[0].Equal(expected) {
			t.Errorf("time = %v, want %v", times[0], expected)
		}
	})

	t.Run("importance 5 should span work hours", func(t *testing.T) {
		times := CalculateReminderTimes(5, 9, 18, now)
		if len(times) != 5 {
			t.Fatalf("expected 5 times, got %d", len(times))
		}

		// First should be at work start
		expectedFirst := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
		if !times[0].Equal(expectedFirst) {
			t.Errorf("first time = %v, want %v", times[0], expectedFirst)
		}

		// Last should be at work end
		expectedLast := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
		if !times[4].Equal(expectedLast) {
			t.Errorf("last time = %v, want %v", times[4], expectedLast)
		}
	})

	t.Run("times should be in order", func(t *testing.T) {
		times := CalculateReminderTimes(5, 9, 18, now)
		for i := 1; i < len(times); i++ {
			if !times[i].After(times[i-1]) {
				t.Errorf("time[%d] (%v) should be after time[%d] (%v)", i, times[i], i-1, times[i-1])
			}
		}
	})
}

func TestShouldSendReminder(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	reminderTimes := []time.Time{
		baseTime,
		baseTime.Add(3 * time.Hour),  // 12:00
		baseTime.Add(6 * time.Hour),  // 15:00
	}

	tests := []struct {
		name               string
		remindersSentToday int
		now                time.Time
		want               bool
	}{
		{
			name:               "no reminders sent, at first time",
			remindersSentToday: 0,
			now:                baseTime,
			want:               true,
		},
		{
			name:               "no reminders sent, before first time",
			remindersSentToday: 0,
			now:                baseTime.Add(-1 * time.Hour),
			want:               false,
		},
		{
			name:               "1 reminder sent, at second time",
			remindersSentToday: 1,
			now:                baseTime.Add(3 * time.Hour),
			want:               true,
		},
		{
			name:               "1 reminder sent, before second time",
			remindersSentToday: 1,
			now:                baseTime.Add(2 * time.Hour),
			want:               false,
		},
		{
			name:               "all reminders sent",
			remindersSentToday: 3,
			now:                baseTime.Add(9 * time.Hour),
			want:               false,
		},
		{
			name:               "more reminders sent than scheduled",
			remindersSentToday: 5,
			now:                baseTime,
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSendReminder(reminderTimes, tt.remindersSentToday, tt.now)
			if got != tt.want {
				t.Errorf("ShouldSendReminder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsWithinWorkHours(t *testing.T) {
	tests := []struct {
		name          string
		workStartHour int
		workEndHour   int
		hour          int
		want          bool
	}{
		{
			name:          "at work start",
			workStartHour: 9,
			workEndHour:   18,
			hour:          9,
			want:          true,
		},
		{
			name:          "during work hours",
			workStartHour: 9,
			workEndHour:   18,
			hour:          12,
			want:          true,
		},
		{
			name:          "at work end",
			workStartHour: 9,
			workEndHour:   18,
			hour:          18,
			want:          false, // end hour is exclusive
		},
		{
			name:          "before work hours",
			workStartHour: 9,
			workEndHour:   18,
			hour:          8,
			want:          false,
		},
		{
			name:          "after work hours",
			workStartHour: 9,
			workEndHour:   18,
			hour:          19,
			want:          false,
		},
		{
			name:          "late night",
			workStartHour: 9,
			workEndHour:   18,
			hour:          2,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2024, 1, 15, tt.hour, 30, 0, 0, time.UTC)
			got := IsWithinWorkHours(tt.workStartHour, tt.workEndHour, now)
			if got != tt.want {
				t.Errorf("IsWithinWorkHours() = %v, want %v", got, tt.want)
			}
		})
	}
}
