package usecase

import (
	"context"
	"merch/internal/entity"
)

type (
	UserRepo interface {
		AddUser(ctx context.Context, user entity.User) error
		Purchase(ctx context.Context, id, coins, value int, item string) error
		Transfer(ctx context.Context, amount int, fromUser, toUser entity.User) error
		GetInfo(ctx context.Context, ID int) (entity.User, []entity.CoinHistory, error)
		GetTransaction(ctx context.Context, ID int) ([]entity.CoinHistory, error)
		GetUserByID(ctx context.Context, ID int) (entity.User, error)
		GetUserByUsername(ctx context.Context, username string) (entity.User, error)
	}

	UserUseCase interface {
		AddUser(ctx context.Context, username, id string) (int, error)
		Purchase(ctx context.Context, id int, username, item string) error
		Transfer(ctx context.Context, amount, fromUserID int, fromUser, toUser string) error
		GetInfo(ctx context.Context, id int, username string) (entity.User, []entity.CoinHistory, error)
	}

	ShopRepo interface {
		GetCost(name string) (int, error)
	}

	Cache[T comparable] interface {
		Get(key T) (any, bool)
		Set(key T, data any)
	}
)
