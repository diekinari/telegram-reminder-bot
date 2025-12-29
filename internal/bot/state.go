package bot

import (
	"sync"
	"time"

	"telegram-reminder-bot/internal/domain"
)

type UserState struct {
	Step        string
	Description string
	Deadline    time.Time
	Importance  int
	Frequency   domain.Frequency
}

type StateManager struct {
	mu     sync.RWMutex
	states map[int64]*UserState
}

func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[int64]*UserState),
	}
}

func (sm *StateManager) Get(userID int64) *UserState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.states[userID]
}

func (sm *StateManager) Set(userID int64, state *UserState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.states[userID] = state
}

func (sm *StateManager) Delete(userID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.states, userID)
}

const (
	StateWaitingDescription = "waiting_description"
	StateWaitingDeadline    = "waiting_deadline"
	StateWaitingImportance  = "waiting_importance"
	StateWaitingFrequency   = "waiting_frequency"
)
