/a# Telegram Reminder Bot

Telegram-bot for task reminders with deadlines, importance levels, and reminder frequency.

## Features

- Create tasks with deadline, importance (1-5), and reminder frequency
- Importance determines how many times per day to remind (1-5 times)
- Frequency determines how often to remind (daily, every other day, weekly)
- Shows remaining time in days and work hours
- Per-user settings for work hours and timezone
- PostgreSQL storage

## Bot Commands

- `/start` - start the bot
- `/add` - add a new task
- `/list` - list active tasks
- `/settings` - settings (work hours, timezone)

## Running

### With Docker (recommended)

1. Create `.env` file:
```bash
cp .env.example .env
```

2. Add your bot token to `.env`:
```
TELEGRAM_BOT_TOKEN=your_bot_token_here
```

3. Run:
```bash
docker-compose up -d
```

### Locally

1. Install PostgreSQL and create a database

2. Create `.env`:
```bash
cp .env.example .env
```

3. Fill in `.env`:
```
TELEGRAM_BOT_TOKEN=your_bot_token_here
DATABASE_URL=postgres://user:pass@localhost:5432/reminder_bot?sslmode=disable
```

4. Run:
```bash
make run
```

## Project Structure

```
├── cmd/bot/main.go          # Entry point
├── internal/
│   ├── bot/                 # Telegram bot
│   ├── config/              # Configuration
│   ├── domain/              # Domain models
│   ├── repository/          # Repositories (PostgreSQL)
│   ├── scheduler/           # Reminder scheduler
│   └── service/             # Business logic
├── migrations/              # SQL migrations
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Testing

Run all tests:
```bash
make test
```

Tests cover:
- `internal/domain` - Task and Frequency models (DaysUntilDeadline, WorkHoursRemaining, ShouldRemindToday, etc.)
- `internal/scheduler` - Reminder time calculations (CalculateReminderTimes, ShouldSendReminder, IsWithinWorkHours)

## Makefile Commands

- `make build` - build binary
- `make run` - build and run
- `make test` - run tests
- `make docker-up` - run with Docker
- `make docker-down` - stop Docker
- `make docker-logs` - view logs

## Troubleshooting

### Bot not responding after code changes
Docker may use cached layers. Rebuild without cache:
```bash
docker-compose down && docker-compose build --no-cache && docker-compose up -d
```

### Connection timeout to Telegram API
Add DNS servers to `docker-compose.yml` under bot service:
```yaml
dns:
  - 8.8.8.8
  - 1.1.1.1
```
