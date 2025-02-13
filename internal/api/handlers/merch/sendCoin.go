package merch

import (
	"context"
	"errors"
	"merch/internal/api/utilapi"
	"merch/internal/entity"
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

	var toUserID int

	user, ok := m.cache.Get(req.ToUser)
	if !ok {
		getUser, err := m.user.GetUser(context.Background(), req.ToUser)
		if err != nil {
			ctx.Error("failed to search user", err)
			ctx.WriteFailure(http.StatusBadRequest, "invalid request")
			return
		}
		m.cache.Set(req.ToUser, entity.User{
			ID:       getUser.ID,
			Coins:    0,
			UserName: req.ToUser,
		})
		toUserID = int(getUser.ID)
	} else {
		toUserID = int(user.(entity.User).ID)
	}

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
