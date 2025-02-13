package merch

import (
	"merch/internal/cache"
	"merch/internal/usecase/storage"
)

type MerchHandlder struct {
	user  storage.UserRepo
	cache *cache.Cache[string]
}

func NewMerchHandler(u storage.UserRepo, c *cache.Cache[string]) *MerchHandlder {
	return &MerchHandlder{
		user:  u,
		cache: c,
	}
}
