package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"

	"telegram-reminder-bot/internal/domain"
	"telegram-reminder-bot/internal/service"
)

type Handler struct {
	userService *service.UserService
	taskService *service.TaskService
	stateManager *StateManager
}

func NewHandler(userService *service.UserService, taskService *service.TaskService) *Handler {
	return &Handler{
		userService:  userService,
		taskService:  taskService,
		stateManager: NewStateManager(),
	}
}

func (h *Handler) HandleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	telegramID := update.Message.From.ID
	username := update.Message.From.Username

	_, err := h.userService.GetOrCreate(ctx, telegramID, username)
	if err != nil {
		log.Error().Err(err).Msg("failed to create user")
		return
	}

	text := `Привет! Я бот для напоминаний о задачах.

Я помогу тебе не забыть о важных делах. Вот что я умею:

📌 Добавить задачу - создать новое напоминание
📋 Мои задачи - посмотреть активные задачи
⚙️ Настройки - изменить рабочие часы

Используй кнопки меню или команды:
/add - добавить задачу
/list - список задач
/settings - настройки`

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        text,
		ReplyMarkup: mainMenuKeyboard(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to send start message")
	}
}

func (h *Handler) HandleAdd(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	h.stateManager.Set(userID, &UserState{Step: StateWaitingDescription})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Введи описание задачи:",
		ReplyMarkup: cancelKeyboard(),
	})
}

func (h *Handler) HandleList(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	telegramID := update.Message.From.ID

	user, err := h.userService.GetOrCreate(ctx, telegramID, update.Message.From.Username)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		return
	}

	tasks, err := h.taskService.GetActiveByUserID(ctx, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get tasks")
		return
	}

	if len(tasks) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "У тебя нет активных задач. Добавь новую с помощью /add",
		})
		return
	}

	for _, task := range tasks {
		text := formatTaskMessage(task, user)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: taskActionsKeyboard(task.ID),
		})
	}
}

func (h *Handler) HandleSettings(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	telegramID := update.Message.From.ID

	user, err := h.userService.GetOrCreate(ctx, telegramID, update.Message.From.Username)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		return
	}

	text := fmt.Sprintf(`⚙️ <b>Настройки</b>

Рабочие часы в день: <b>%d</b>
Часовой пояс: <b>%s</b>

Выбери что изменить:`, user.WorkHoursPerDay, user.Timezone)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: settingsKeyboard(),
	})
}

func (h *Handler) HandleMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	text := update.Message.Text
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	switch text {
	case "Добавить задачу":
		h.HandleAdd(ctx, b, update)
		return
	case "Мои задачи":
		h.HandleList(ctx, b, update)
		return
	case "Настройки":
		h.HandleSettings(ctx, b, update)
		return
	}

	state := h.stateManager.Get(userID)
	if state == nil {
		return
	}

	switch state.Step {
	case StateWaitingDescription:
		state.Description = text
		state.Step = StateWaitingDeadline
		h.stateManager.Set(userID, state)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Введи дедлайн в формате ДД.ММ.ГГГГ (например, 15.01.2025):",
			ReplyMarkup: cancelKeyboard(),
		})

	case StateWaitingDeadline:
		deadline, err := time.Parse("02.01.2006", text)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Неверный формат даты. Введи в формате ДД.ММ.ГГГГ:",
			})
			return
		}

		if deadline.Before(time.Now().Truncate(24 * time.Hour)) {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "Дедлайн не может быть в прошлом. Введи корректную дату:",
			})
			return
		}

		state.Deadline = deadline
		state.Step = StateWaitingImportance
		h.stateManager.Set(userID, state)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Выбери важность задачи (влияет на количество напоминаний в день):",
			ReplyMarkup: importanceKeyboard(),
		})
	}
}

func (h *Handler) HandleCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	callback := update.CallbackQuery
	chatID := callback.Message.Message.Chat.ID
	userID := callback.From.ID
	data := callback.Data

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
	})

	if data == "cancel" {
		h.stateManager.Delete(userID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Действие отменено.",
			ReplyMarkup: mainMenuKeyboard(),
		})
		return
	}

	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		return
	}

	action, value := parts[0], parts[1]

	switch action {
	case "importance":
		h.handleImportanceCallback(ctx, b, chatID, userID, value)
	case "frequency":
		h.handleFrequencyCallback(ctx, b, chatID, userID, value)
	case "done":
		h.handleDoneCallback(ctx, b, chatID, callback.Message.Message.ID, value)
	case "delete":
		h.handleDeleteCallback(ctx, b, chatID, callback.Message.Message.ID, value)
	case "settings":
		h.handleSettingsCallback(ctx, b, chatID, userID, value)
	case "work_hours":
		h.handleWorkHoursCallback(ctx, b, chatID, userID, value)
	}
}

func (h *Handler) handleImportanceCallback(ctx context.Context, b *bot.Bot, chatID int64, userID int64, value string) {
	importance, err := strconv.Atoi(value)
	if err != nil {
		return
	}

	state := h.stateManager.Get(userID)
	if state == nil || state.Step != StateWaitingImportance {
		return
	}

	state.Importance = importance
	state.Step = StateWaitingFrequency
	h.stateManager.Set(userID, state)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        "Выбери частоту напоминаний:",
		ReplyMarkup: frequencyKeyboard(),
	})
}

func (h *Handler) handleFrequencyCallback(ctx context.Context, b *bot.Bot, chatID int64, userID int64, value string) {
	frequency, ok := domain.ParseFrequency(value)
	if !ok {
		return
	}

	state := h.stateManager.Get(userID)
	if state == nil || state.Step != StateWaitingFrequency {
		return
	}

	user, err := h.userService.GetOrCreate(ctx, userID, "")
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		return
	}

	task, err := h.taskService.Create(ctx, user.ID, state.Description, state.Deadline, state.Importance, frequency)
	if err != nil {
		log.Error().Err(err).Msg("failed to create task")
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Ошибка при создании задачи. Попробуй ещё раз.",
		})
		return
	}

	h.stateManager.Delete(userID)

	text := fmt.Sprintf("✅ Задача создана!\n\n%s", formatTaskMessage(task, user))
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: mainMenuKeyboard(),
	})
}

func (h *Handler) handleDoneCallback(ctx context.Context, b *bot.Bot, chatID int64, messageID int, value string) {
	taskID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return
	}

	if err := h.taskService.Complete(ctx, taskID); err != nil {
		log.Error().Err(err).Msg("failed to complete task")
		return
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      "✅ Задача выполнена!",
	})
}

func (h *Handler) handleDeleteCallback(ctx context.Context, b *bot.Bot, chatID int64, messageID int, value string) {
	taskID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return
	}

	if err := h.taskService.Delete(ctx, taskID); err != nil {
		log.Error().Err(err).Msg("failed to delete task")
		return
	}

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      "🗑 Задача удалена.",
	})
}

func (h *Handler) handleSettingsCallback(ctx context.Context, b *bot.Bot, chatID int64, _ int64, value string) {
	switch value {
	case "work_hours":
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        "Выбери количество рабочих часов в день:",
			ReplyMarkup: workHoursKeyboard(),
		})
	case "timezone":
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Отправь свой часовой пояс (например, Europe/Moscow, Asia/Yekaterinburg):",
		})
	}
}

func (h *Handler) handleWorkHoursCallback(ctx context.Context, b *bot.Bot, chatID int64, userID int64, value string) {
	hours, err := strconv.Atoi(value)
	if err != nil {
		return
	}

	user, err := h.userService.GetOrCreate(ctx, userID, "")
	if err != nil {
		log.Error().Err(err).Msg("failed to get user")
		return
	}

	user.WorkHoursPerDay = hours
	if err := h.userService.UpdateSettings(ctx, user); err != nil {
		log.Error().Err(err).Msg("failed to update user settings")
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        fmt.Sprintf("✅ Рабочие часы обновлены: %d часов в день", hours),
		ReplyMarkup: mainMenuKeyboard(),
	})
}
