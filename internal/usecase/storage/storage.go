package storage

import (
	"context"
	"merch/internal/entity"
)

type UserRepo interface {
	AddUser(ctx context.Context, user entity.User) (int, error)
	Purchase(ctx context.Context, user entity.User, value int, item string) error
	Transfer(ctx context.Context, amount int, fromUser, toUser entity.User) error
	GetInfo(ctx context.Context, user entity.User) (entity.User, []entity.CoinHistory, error)
	GetUserByUsername(ctx context.Context, username string) (entity.User, error)
}

type Storage struct {
	repo UserRepo
}

func (s *Storage) AddUser(ctx context.Context, user entity.User) (int, error) {
	return s.repo.AddUser(ctx, user)
}

func (s *Storage) Purchase(ctx context.Context, user entity.User, value int, item string) error {
	return s.repo.Purchase(ctx, user, value, item)
}

func (s *Storage) GetInfo(ctx context.Context, user entity.User) (entity.User, []entity.CoinHistory, error) {
	return s.repo.GetInfo(ctx, user)
}

func (s *Storage) Transfer(ctx context.Context, amount int, fromUserID, toUserID entity.User) error {
	return s.repo.Transfer(ctx, amount, fromUserID, toUserID)
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (entity.User, error) {
	return s.repo.GetUserByUsername(ctx, username)
}

func NewStorage(repo UserRepo) *Storage {
	return &Storage{
		repo: repo,
	}
}
