package usecase

import (
	"context"
	"errors"
	"fmt"
	"merch/internal/entity"
)

var (
	ErrNoItem           = errors.New("no such item")
	ErrNotAuthorization = errors.New("invalid password")
	ErrNotEnoughCoin    = errors.New("not enough coin")
	ErrUserNotFound     = errors.New("no such user")
)

type UseCase struct {
	repo UserRepo
	//cache *cache.Cache[string]
	cache Cache[string]
	shop  ShopRepo
}

func (u *UseCase) AddUser(ctx context.Context, username, password string) (int, error) {
	const op = "UseCase - AddUser"

	var user entity.User
	var err error

	cacheUser, ok := u.cache.Get(username)
	if !ok {
		user, err = entity.NewUser(username, password)
		if err != nil {
			return 0, fmt.Errorf("%s - failed to create new user: %w", op, err)
		}

		err = u.repo.AddUser(ctx, user)
		if err != nil {
			return 0, fmt.Errorf("%s - failed to add user to db: %w", op, err)
		}
		u.cache.Set(username, user)
	} else {
		user = cacheUser.(entity.User)
		if !user.Identification(password) {
			return 0, fmt.Errorf("%s - failed to identification user: %w", op, ErrNotAuthorization)
		}
	}

	return int(user.ID), nil
}

func (u *UseCase) Purchase(ctx context.Context, id int, username, item string) error {
	const op = "UseCase - Purchase"

	var user entity.User
	var err error

	cost, err := u.shop.GetCost(item)
	if err != nil {
		if errors.Is(err, ErrNoItem) {
			return ErrNoItem
		}
		return fmt.Errorf("%s - failed to get cost item: %w", op, err)
	}

	cacheUser, ok := u.cache.Get(username)
	if !ok {
		user, err = u.repo.GetUserByID(ctx, id)
		if err != nil {
			return fmt.Errorf("%s - failed to get user by id: %w", op, err)
		}
		u.cache.Set(username, user)
	} else {
		user = cacheUser.(entity.User)
	}

	if user.Coins < cost {
		return ErrNotEnoughCoin
	}

	err = u.repo.Purchase(ctx, id, user.Coins, cost, item)
	if err != nil {
		return fmt.Errorf("%s - failed to buy item: %w", op, err)
	}

	user.Coins -= cost
	exists := false
	for k, v := range user.Inventory {
		if v.Item == item {
			user.Inventory[k].Quantity++
			exists = true
			break
		}
	}

	if !exists {
		user.Inventory = append(user.Inventory, entity.Inventory{
			UserID:   user.ID,
			Item:     item,
			Quantity: 1,
		})
	}

	u.cache.Set(username, user)

	return nil
}

func (u *UseCase) GetInfo(ctx context.Context, id int, username string) (entity.User, []entity.CoinHistory, error) {
	const op = "UseCase - GetInfo"

	var user entity.User
	var history []entity.CoinHistory
	var err error

	cacheUser, ok := u.cache.Get(username)
	if !ok {
		user, history, err = u.repo.GetInfo(ctx, id)
		if err != nil {
			return entity.User{}, nil, fmt.Errorf("%s - failed to get user info: %w", op, err)
		}
		u.cache.Set(username, user)
	} else {
		user = cacheUser.(entity.User)
		history, err = u.repo.GetTransaction(ctx, id)
		if err != nil {
			return entity.User{}, nil, fmt.Errorf("%s - failed to get user transactions: %w", op, err)
		}
	}

	return user, history, nil
}

func (u *UseCase) Transfer(ctx context.Context, amount, fromID int, fromUsername, toUsername string) error {
	const op = "UseCase - Transfer"

	var toUser entity.User
	var err error

	cacheToUser, ok := u.cache.Get(toUsername)
	if !ok {
		toUser, err = u.repo.GetUserByUsername(ctx, toUsername)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				return ErrUserNotFound
			} else {
				return fmt.Errorf("%s - failed to get user by username: %w", op, err)
			}
		}
		u.cache.Set(toUsername, toUser)
	} else {
		toUser = cacheToUser.(entity.User)
	}

	var fromUser entity.User

	cacheFromUser, ok := u.cache.Get(fromUsername)
	if !ok {
		fromUser, err = u.repo.GetUserByID(ctx, fromID)
		if err != nil {
			return fmt.Errorf("%s - failed to get user by id: %w", op, err)
		}
		u.cache.Set(fromUsername, fromUser)
	} else {
		fromUser = cacheFromUser.(entity.User)
		if fromUser.Coins < amount {
			return ErrNotEnoughCoin
		}
	}

	err = u.repo.Transfer(ctx, amount, fromUser, toUser)
	if err != nil {
		if errors.Is(err, ErrNotEnoughCoin) {
			return ErrNotEnoughCoin
		} else {
			return fmt.Errorf("%s - failed to transfer coin: %w", op, err)
		}
	}

	fromUser.Coins -= amount
	toUser.Coins += amount

	u.cache.Set(toUsername, toUser)
	u.cache.Set(fromUsername, fromUser)

	return nil
}

func (u *UseCase) GetUserByUsername(ctx context.Context, username string) (entity.User, error) {
	return u.repo.GetUserByUsername(ctx, username)
}

func NewStorage(repo UserRepo, c Cache[string], shop ShopRepo) *UseCase {
	return &UseCase{
		repo:  repo,
		cache: c,
		shop:  shop,
	}
}
