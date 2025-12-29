package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"telegram-reminder-bot/internal/bot"
	"telegram-reminder-bot/internal/config"
	"telegram-reminder-bot/internal/repository/postgres"
	"telegram-reminder-bot/internal/scheduler"
	"telegram-reminder-bot/internal/service"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	userRepo := postgres.NewUserRepository(db)
	taskRepo := postgres.NewTaskRepository(db)

	userService := service.NewUserService(userRepo)
	taskService := service.NewTaskService(taskRepo)

	telegramBot, err := bot.New(cfg.TelegramBotToken, userService, taskService)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create telegram bot")
	}

	reminderScheduler, err := scheduler.New(taskService, userRepo, telegramBot)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create scheduler")
	}

	if err := reminderScheduler.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to start scheduler")
	}

	go telegramBot.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Info().Msg("shutting down...")

	if err := reminderScheduler.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop scheduler")
	}

	cancel()
	log.Info().Msg("shutdown complete")
}
