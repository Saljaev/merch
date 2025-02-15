package merch

import (
	"errors"
	"merch/internal/api/utilapi"
	"merch/internal/usecase/usecase"
	"net/http"
	"strconv"
)

var (
	ErrInvalidItem = errors.New("no such item")
)

func (m *MerchHandler) Buy(ctx *utilapi.APIContext) {
	item := ctx.GetFromQuery("item")

	if item == "" {
		ctx.Error("invalid item", ErrInvalidItem)
		ctx.WriteFailure(http.StatusBadRequest, "no such item")
		return
	}

	username := ctx.GetValue("username").(string)
	idURL := ctx.GetValue("id").(string)
	id, _ := strconv.Atoi(idURL)

	err := m.user.Purchase(ctx, id, username, item)
	if err != nil {
		ctx.Error("failed to buy item", err)
		if errors.Is(err, usecase.ErrNotEnoughCoin) {
			ctx.WriteFailure(http.StatusBadRequest, "not enough coin")
		} else {
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		}
		return
	}
	
	ctx.Info("successful buy item", "item", item)
	ctx.SuccessWithData("OK")
}
