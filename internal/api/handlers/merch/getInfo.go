package merch

import (
	"errors"
	"merch/internal/api/utilapi"
	"merch/internal/entity"
	"net/http"
	"strconv"
)

type (
	Transaction struct {
		FromUser string `json:"fromUser,omitempty"`
		ToUser   string `json:"toUser,omitempty"`
		Amount   int    `json:"amount"`
	}

	Item struct {
		Type     string `json:"type"`
		Quantity int    `json:"quantity"`
	}

	CoinHistory struct {
		Received []Transaction `json:"received"`
		Sent     []Transaction `json:"sent"`
	}

	UserInfoResp struct {
		Coins int `json:"coins"`

		Inventory []Item `json:"inventory"`

		CoinHistory CoinHistory `json:"coinHistory"`
	}
)

func (m *MerchHandler) Info(ctx *utilapi.APIContext) {
	username, ok := ctx.GetValue("username").(string)
	if !ok {
		ctx.Error("failed to get username from ctx", errors.New("invalid type"))
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
		ctx.Error("failed to convert string", err)
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}

	var userInfo entity.User
	var history []entity.CoinHistory

	userInfo, history, err = m.user.GetInfo(ctx, id, username)
	if err != nil {
		ctx.Error("failed to get info user", err)
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}

	resp := makeInfoResp(userInfo, history)

	ctx.Info("successful get info", "user", username)
	ctx.SuccessWithData(resp)
}

func makeInfoResp(user entity.User, history []entity.CoinHistory) UserInfoResp {
	resp := UserInfoResp{
		Coins:     user.Coins,
		Inventory: make([]Item, len(user.Inventory)),
	}

	items := []Item{}

	for _, v := range user.Inventory {
		item := Item{
			Type:     v.Item,
			Quantity: v.Quantity,
		}

		items = append(items, item)
	}

	received := []Transaction{}
	sent := []Transaction{}

	for _, v := range history {
		tran := Transaction{Amount: v.Amount}

		if v.Type == "sent" {
			tran.ToUser = v.ToUser
			sent = append(sent, tran)
		} else {
			tran.FromUser = v.FromUser
			received = append(received, tran)
		}

	}

	resp.CoinHistory.Sent = sent
	resp.CoinHistory.Received = received
	resp.Inventory = items

	return resp
}
