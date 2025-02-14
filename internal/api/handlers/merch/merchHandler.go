package merch

import (
	"context"
	"merch/internal/cache"
	"merch/internal/entity"
	"merch/internal/usecase/shop"
	"merch/internal/usecase/storage"
)

type MerchHandlder struct {
	user  storage.UserRepo
	cache *cache.Cache[string]
	shop  shop.ShopRepo
}

func NewMerchHandler(u storage.UserRepo, c *cache.Cache[string], s shop.ShopRepo) *MerchHandlder {
	return &MerchHandlder{
		user:  u,
		cache: c,
		shop:  s,
	}
}

func (m *MerchHandlder) getUser(username string) *entity.User {
	user, ok := m.cache.Get(username)
	if !ok {
		getUser, err := m.user.GetUserByUsername(context.Background(), username)
		if err != nil {
			return nil
		}

		user = entity.User{
			ID:       getUser.ID,
			Coins:    getUser.Coins,
			UserName: username,
		}

		m.cache.Set(username, user)
	}

	u := user.(entity.User)

	return &u
}
