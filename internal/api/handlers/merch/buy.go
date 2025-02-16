package merch

import (
	"errors"
	"merch/internal/api/utilapi"
	"merch/internal/usecase"
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

	username, ok := ctx.GetValue("username").(string)
	if !ok {
		ctx.Error("failed to get username string from ctx", errors.New("invalid type"))
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}
	idURL, ok := ctx.GetValue("id").(string)
	if !ok {
		ctx.Error("failed to get id from ctx", errors.New("invalid type"))
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}
	id, err := strconv.Atoi(idURL)
	if err != nil {
		ctx.Error("failed to convert id", err)
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}

	err = m.user.Purchase(ctx, id, username, item)
	if err != nil {
		ctx.Error("failed to buy item", err)
		if errors.Is(errors.Unwrap(err), usecase.ErrNotEnoughCoin) {
			ctx.WriteFailure(http.StatusBadRequest, "not enough coin")
		} else if errors.Is(errors.Unwrap(err), usecase.ErrNoItem) {
			ctx.WriteFailure(http.StatusBadRequest, "no such item")
		} else {
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		}
		return
	}

	ctx.Info("successful buy item", "item", item)
	ctx.SuccessWithData("OK")
}
