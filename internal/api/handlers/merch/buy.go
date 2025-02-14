package merch

import (
	"errors"
	"fmt"
	"merch/internal/api/utilapi"
	shoprepo "merch/internal/usecase/shop/repo"
	"net/http"
)

var ErrInvalidItem = errors.New("no such item")

func (m *MerchHandlder) Buy(ctx *utilapi.APIContext) {
	item := ctx.GetFromQuery("item")
	fmt.Println(item)

	if item == "" {
		ctx.Error("invalid item", ErrInvalidItem)
		ctx.WriteFailure(http.StatusBadRequest, "invalid item")
		return
	}

	//TODO: username from token
	username := ctx.GetFromHeader("username")
	user := m.getUser(username)
	cost, err := m.shop.Buy(user.Coins, item)
	if err != nil {
		if errors.Is(shoprepo.ErrNoItem, err) {
			ctx.Error("invalid item", err)
			ctx.WriteFailure(http.StatusBadRequest, "invalid item")
		} else {
			ctx.Error("not enough coins", err)
			ctx.WriteFailure(http.StatusBadRequest, "not enough coins")
		}
		return
	}

	err = m.user.Purchase(ctx, int(user.ID), cost, item)
	if err != nil {
		ctx.Error("failed to buy item", err)
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}
	user.Coins -= cost
	m.cache.Set("username", user)

	ctx.SuccessWithData("OK")
}
