package merch

import (
	"errors"
	"merch/internal/api/utilapi"
	"merch/internal/usecase/storage/repo/postgres"
	"net/http"
	"unicode/utf8"
)

var (
	ErrNotEnoughCoin = errors.New("not enough coin")
	ErrNotSuchUser   = errors.New("not souch user")
)

type SendCoinReq struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

func (s *SendCoinReq) IsValid() bool {
	return utf8.RuneCountInString(s.ToUser) >= 0 && s.Amount > 0
}

func (m *MerchHandlder) SendCoin(ctx *utilapi.APIContext) {
	var req SendCoinReq

	err := ctx.Decode(&req)
	if err != nil {
		ctx.Error("failed to decode req", err)
		ctx.WriteFailure(http.StatusBadRequest, "invalid request")
		return
	}

	username := ctx.GetValue("username").(string)
	fromUser := m.getUser(username)

	toUser := m.getUser(req.ToUser)

	if toUser.UserName == "" {
		ctx.Error("not such user", ErrNotSuchUser)
		ctx.WriteFailure(http.StatusBadRequest, "not such user")
		return
	}

	if fromUser.Coins-req.Amount < 0 {
		ctx.Error("not enough coin", ErrNotEnoughCoin)
		ctx.WriteFailure(http.StatusBadRequest, "not enough coin")
		return
	}

	err = m.user.Transfer(ctx, req.Amount, fromUser, toUser)
	if err != nil {
		ctx.Error("failed to transfer coin", err)
		if errors.Is(err, postgres.ErrNotEnoughCoins) {
			ctx.WriteFailure(http.StatusBadRequest, "not enough coin")
		} else {
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		}
		return
	}

	fromUser.Coins -= req.Amount
	m.cache.Set(fromUser.UserName, fromUser)

	toUser.Coins -= req.Amount
	m.cache.Set(toUser.UserName, toUser)

	ctx.Info("successful transfer coin to user", "to_user", req.ToUser)
	ctx.SuccessWithData("OK")
}
