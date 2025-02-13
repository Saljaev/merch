package storage

import (
	"context"
	"merch/internal/entity"
)

type UserRepo interface {
	AddUser(ctx context.Context, user entity.User) (int, error)
	Purchase(ctx context.Context, userID, value int, item string) error
	GetInfo(ctx context.Context, userID int) (entity.User, []entity.CoinHistory, error)
	Transfer(ctx context.Context, fromUserID, toUserID, amount int) error
	GetUser(ctx context.Context, username string) (entity.User, error)
}

type Storage struct {
	repo UserRepo
}

func (s *Storage) AddUser(ctx context.Context, user entity.User) (int, error) {
	return s.repo.AddUser(ctx, user)
}

func (s *Storage) Purchase(ctx context.Context, userID, value int, item string) error {
	return s.repo.Purchase(ctx, userID, value, item)
}

func (s *Storage) GetInfo(ctx context.Context, userID int) (entity.User, []entity.CoinHistory, error) {
	return s.repo.GetInfo(ctx, userID)
}

func (s *Storage) Transfer(ctx context.Context, fromUserID, toUserID, amount int) error {
	return s.repo.Transfer(ctx, fromUserID, toUserID, amount)
}

func (s *Storage) GetUser(ctx context.Context, username string) (entity.User, error) {
	return s.repo.GetUser(ctx, username)
}

func NewStorage(repo UserRepo) *Storage {
	return &Storage{
		repo: repo,
	}
}
