package merch

import (
	"context"
	"merch/internal/api/handlers/auth"
	"merch/internal/cache"
	"merch/internal/entity"
	"merch/internal/usecase/shop"
	"merch/internal/usecase/storage"
)

type MerchHandlder struct {
	user  storage.UserRepo
	cache *cache.Cache[string]
	shop  shop.ShopRepo
	jwt   *auth.JWTManager
}

func NewMerchHandler(u storage.UserRepo, c *cache.Cache[string], s shop.ShopRepo, j *auth.JWTManager) *MerchHandlder {
	return &MerchHandlder{
		user:  u,
		cache: c,
		shop:  s,
		jwt:   j,
	}
}

func (m *MerchHandlder) getUser(username string) entity.User {
	user, ok := m.cache.Get(username)
	if !ok {
		getUser, err := m.user.GetUserByUsername(context.Background(), username)
		if err != nil {
			return entity.User{}
		}

		m.cache.Set(username, getUser)
		return getUser
	}

	//TODO: valid from nil
	u := user.(entity.User)

	return u
}
