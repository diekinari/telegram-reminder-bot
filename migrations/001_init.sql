-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    timezone VARCHAR(50) DEFAULT 'Europe/Moscow',
    work_hours_per_day INT DEFAULT 8,
    work_start_hour INT DEFAULT 9,
    work_end_hour INT DEFAULT 18,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    deadline DATE NOT NULL,
    importance INT NOT NULL CHECK (importance >= 1 AND importance <= 5),
    frequency VARCHAR(20) NOT NULL CHECK (frequency IN ('daily', 'every_other_day', 'weekly')),
    is_completed BOOLEAN DEFAULT FALSE,
    last_reminder_date DATE,
    reminders_sent_today INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_deadline ON tasks(deadline);
CREATE INDEX IF NOT EXISTS idx_tasks_active ON tasks(user_id, is_completed, deadline) WHERE is_completed = FALSE;
