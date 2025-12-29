package service

import (
	"context"

	"telegram-reminder-bot/internal/domain"
	"telegram-reminder-bot/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetOrCreate(ctx context.Context, telegramID int64, username string) (*domain.User, error) {
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	user = domain.NewUser(telegramID, username)
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateSettings(ctx context.Context, user *domain.User) error {
	return s.userRepo.Update(ctx, user)
}
