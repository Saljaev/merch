package shoprepo

import (
	_ "embed"
	"errors"
	"log"
	"merch/internal/usecase/shop"
	"strconv"
	"strings"
)

type MapStore struct {
	store map[string]int
}

var _ shop.ShopRepo = (*MapStore)(nil)

func (m *MapStore) Buy(userCoin int, name string) (int, error) {
	v, ok := m.store[name]
	if !ok {
		return 0, ErrNoItem
	}
	if userCoin < v {
		return 0, ErrNotEnoughCoin
	}

	return v, nil
}

var (
	ErrNoItem        = errors.New("no such item")
	ErrNotEnoughCoin = errors.New("no enough coin to buy item")
)

//go:embed store.txt
var itemList string

func NewMapStore() *MapStore {
	items := strings.Split(itemList, "\n")

	result := make(map[string]int, len(items))

	for _, item := range items {
		fields := strings.Split(item, " ")
		name, value := fields[0], fields[1]
		cost, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("failed to parse item: %s value: %v", name, value)
		}

		result[name] = cost
	}

	return &MapStore{store: result}
}
