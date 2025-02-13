package storage

type Shop struct {
	repo ShopRepo
}

type ShopRepo interface {
	Buy(userCoin int, name string) (int, error)
}

func (m *Shop) Buy(userCoin int, name string) (int, error) {
	return m.repo.Buy(userCoin, name)
}

func NewShop(repo ShopRepo) *Shop {
	return &Shop{repo: repo}
}
