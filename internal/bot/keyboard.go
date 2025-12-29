package bot

import (
	"fmt"
	"strings"

	"github.com/go-telegram/bot/models"

	"telegram-reminder-bot/internal/domain"
)

func mainMenuKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: "Добавить задачу"}, {Text: "Мои задачи"}},
			{{Text: "Настройки"}},
		},
		ResizeKeyboard: true,
	}
}

func importanceKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "1 ★", CallbackData: "importance:1"},
				{Text: "2 ★★", CallbackData: "importance:2"},
				{Text: "3 ★★★", CallbackData: "importance:3"},
			},
			{
				{Text: "4 ★★★★", CallbackData: "importance:4"},
				{Text: "5 ★★★★★", CallbackData: "importance:5"},
			},
		},
	}
}

func frequencyKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Ежедневно", CallbackData: "frequency:daily"}},
			{{Text: "Через день", CallbackData: "frequency:every_other_day"}},
			{{Text: "Раз в неделю", CallbackData: "frequency:weekly"}},
		},
	}
}

func taskActionsKeyboard(taskID int64) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Выполнено", CallbackData: fmt.Sprintf("done:%d", taskID)},
				{Text: "Удалить", CallbackData: fmt.Sprintf("delete:%d", taskID)},
			},
		},
	}
}

func reminderKeyboard(taskID int64) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Выполнено", CallbackData: fmt.Sprintf("done:%d", taskID)},
			},
		},
	}
}

func settingsKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Рабочие часы в день", CallbackData: "settings:work_hours"}},
			{{Text: "Часовой пояс", CallbackData: "settings:timezone"}},
		},
	}
}

func workHoursKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "6 часов", CallbackData: "work_hours:6"},
				{Text: "8 часов", CallbackData: "work_hours:8"},
				{Text: "10 часов", CallbackData: "work_hours:10"},
			},
			{
				{Text: "12 часов", CallbackData: "work_hours:12"},
			},
		},
	}
}

func cancelKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "Отмена", CallbackData: "cancel"}},
		},
	}
}

func formatTaskMessage(task *domain.Task, user *domain.User) string {
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

	return fmt.Sprintf(`📋 <b>%s</b>

⏰ До дедлайна: <b>%d %s</b>
⏱ Рабочих часов осталось: <b>%d %s</b>
⚡ Важность: %s (%d/5)
🔄 Частота: %s`,
		escapeHTML(task.Description),
		days, daysText,
		hours, hoursText,
		task.ImportanceStars(), task.Importance,
		task.Frequency.DisplayName(),
	)
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func escapeMarkdown(s string) string {
	replacer := map[rune]string{
		'_': "\\_",
		'*': "\\*",
		'[': "\\[",
		']': "\\]",
		'(': "\\(",
		')': "\\)",
		'~': "\\~",
		'`': "\\`",
		'>': "\\>",
		'#': "\\#",
		'+': "\\+",
		'-': "\\-",
		'=': "\\=",
		'|': "\\|",
		'{': "\\{",
		'}': "\\}",
		'.': "\\.",
		'!': "\\!",
	}

	result := ""
	for _, r := range s {
		if escaped, ok := replacer[r]; ok {
			result += escaped
		} else {
			result += string(r)
		}
	}
	return result
}
