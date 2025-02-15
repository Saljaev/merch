package shop

import (
	_ "embed"
	"errors"
	"log"
	"merch/internal/usecase/usecase"
	"strconv"
	"strings"
)

type MapStore struct {
	store map[string]int
}

var _ usecase.ShopRepo = (*MapStore)(nil)

func (m *MapStore) GetCost(name string) (int, error) {
	v, ok := m.store[name]
	if !ok {
		return 0, errors.Join(usecase.ErrNoItem)
	}

	return v, nil
}

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
