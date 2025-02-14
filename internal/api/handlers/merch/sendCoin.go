package merch

import (
	"errors"
	"merch/internal/api/utilapi"
	"merch/internal/usecase/storage/repo/postgres"
	"net/http"
	"strconv"
	"unicode/utf8"
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

	//TODO: fromUserID from token JWT
	id := ctx.GetFromHeader("id")
	fromUserID, _ := strconv.Atoi(id)

	user := m.getUser(req.ToUser)

	toUserID := int(user.ID)

	err = m.user.Transfer(ctx, fromUserID, toUserID, req.Amount)
	if err != nil {
		ctx.Error("failed to transfer coin", err)
		if errors.Is(err, postgres.ErrNotEnoughCoins) {
			ctx.WriteFailure(http.StatusBadRequest, "not enough coin")
		} else {
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		}
		return
	}

	ctx.Info("successful transfer coin to user", "user", req.ToUser)
	ctx.SuccessWithData("OK")
}
