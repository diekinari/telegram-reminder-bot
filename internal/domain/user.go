package domain

import "time"

type User struct {
	ID              int64
	TelegramID      int64
	Username        string
	Timezone        string
	WorkHoursPerDay int
	WorkStartHour   int
	WorkEndHour     int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewUser(telegramID int64, username string) *User {
	return &User{
		TelegramID:      telegramID,
		Username:        username,
		Timezone:        "Europe/Moscow",
		WorkHoursPerDay: 8,
		WorkStartHour:   9,
		WorkEndHour:     18,
	}
}

func (u *User) Location() *time.Location {
	loc, err := time.LoadLocation(u.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}
