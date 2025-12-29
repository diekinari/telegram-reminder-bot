package scheduler

import (
	"time"
)

func CalculateReminderTimes(importance int, workStartHour, workEndHour int, now time.Time) []time.Time {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	workStart := today.Add(time.Duration(workStartHour) * time.Hour)
	workEnd := today.Add(time.Duration(workEndHour) * time.Hour)

	if importance <= 0 {
		return nil
	}

	if importance == 1 {
		midDay := workStart.Add(workEnd.Sub(workStart) / 2)
		return []time.Time{midDay}
	}

	times := make([]time.Time, importance)
	duration := workEnd.Sub(workStart)
	interval := duration / time.Duration(importance-1)

	for i := 0; i < importance; i++ {
		times[i] = workStart.Add(interval * time.Duration(i))
	}

	return times
}

func ShouldSendReminder(reminderTimes []time.Time, remindersSentToday int, now time.Time) bool {
	if remindersSentToday >= len(reminderTimes) {
		return false
	}

	nextReminderTime := reminderTimes[remindersSentToday]

	return now.After(nextReminderTime) || now.Equal(nextReminderTime)
}

func IsWithinWorkHours(workStartHour, workEndHour int, now time.Time) bool {
	hour := now.Hour()
	return hour >= workStartHour && hour < workEndHour
}
