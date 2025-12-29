package bot

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"

	"telegram-reminder-bot/internal/service"
)

type Bot struct {
	bot     *bot.Bot
	handler *Handler
}

func New(token string, userService *service.UserService, taskService *service.TaskService) (*Bot, error) {
	handler := NewHandler(userService, taskService)

	opts := []bot.Option{
		bot.WithDefaultHandler(handler.defaultHandler),
		bot.WithDebug(),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, err
	}

	// Delete webhook to ensure long polling works
	b.DeleteWebhook(context.Background(), &bot.DeleteWebhookParams{})

	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, handler.HandleStart)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/add", bot.MatchTypeExact, handler.HandleAdd)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/list", bot.MatchTypeExact, handler.HandleList)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/settings", bot.MatchTypeExact, handler.HandleSettings)
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, handler.HandleCallback)

	return &Bot{
		bot:     b,
		handler: handler,
	}, nil
}

func (b *Bot) Start(ctx context.Context) {
	log.Info().Msg("starting telegram bot")
	b.bot.Start(ctx)
}

func (b *Bot) SendReminder(ctx context.Context, telegramID int64, message string, taskID int64) error {
	_, err := b.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      telegramID,
		Text:        message,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: reminderKeyboard(taskID),
	})
	return err
}

func (h *Handler) defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		h.HandleMessage(ctx, b, update)
	}
}
