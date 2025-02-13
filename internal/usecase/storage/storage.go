package storage

import (
	"context"
	"merch/internal/entity"
)

const BucketCount = 4

type UserRepo interface {
	AddUser(ctx context.Context, user entity.User) (int, error)
	Purchase(ctx context.Context, userID, value int, item string) error
	GetInfo(ctx context.Context, userID int) (*entity.User, []*entity.CoinHistory, error)
	Transfer(ctx context.Context, fromUserID, toUserID, amount int) error
}

type Storage struct {
	repo UserRepo
}

func NewStorage(repo UserRepo) *Storage {
	return &Storage{
		repo: repo,
	}
}
