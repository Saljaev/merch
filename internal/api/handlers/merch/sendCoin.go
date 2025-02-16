package merch

import (
	"errors"
	"merch/internal/api/utilapi"
	"merch/internal/usecase"
	"net/http"
	"strconv"
	"unicode/utf8"
)

var (
	ErrTransferToYourself = errors.New("transfer coin to yourself")
)

type SendCoinReq struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

func (s *SendCoinReq) IsValid() bool {
	return utf8.RuneCountInString(s.ToUser) >= 0 && s.Amount > 0
}

func (m *MerchHandler) SendCoin(ctx *utilapi.APIContext) {
	var req SendCoinReq

	err := ctx.Decode(&req)
	if err != nil {
		ctx.Error("failed to decode req", err)
		ctx.WriteFailure(http.StatusBadRequest, "invalid request")
		return
	}

	fromUsername := ctx.GetValue("username").(string)
	idURL := ctx.GetValue("id").(string)
	fromID, _ := strconv.Atoi(idURL)

	if fromUsername == req.ToUser {
		ctx.Error("failed to transfer", ErrTransferToYourself)
		ctx.WriteFailure(http.StatusBadRequest, "can't transfer coin to yourself")
		return
	}

	err = m.user.Transfer(ctx, req.Amount, fromID, fromUsername, req.ToUser)
	if err != nil {
		ctx.Error("failed to transfer coin", err)
		if errors.Is(errors.Unwrap(err), usecase.ErrNotEnoughCoin) {
			ctx.WriteFailure(http.StatusBadRequest, "not enough coin")
		} else if errors.Is(errors.Unwrap(err), usecase.ErrUserNotFound) {
			ctx.WriteFailure(http.StatusBadRequest, "not such user")
		} else {
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		}
		return
	}

	ctx.Info("successful transfer coin to user", "to_user", req.ToUser)
	ctx.SuccessWithData("OK")
}
