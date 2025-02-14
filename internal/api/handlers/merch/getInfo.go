package merch

import (
	"merch/internal/api/utilapi"
	"merch/internal/entity"
	"net/http"
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
		Coins       int         `json:"coins"`
		Inventory   []Item      `json:"inventory"`
		CoinHistory CoinHistory `json:"coinHistory"`
	}
)

func (m *MerchHandlder) Info(ctx *utilapi.APIContext) {
	//TODO: get username from token
	username := ctx.GetFromHeader("username")
	user := m.getUser(username)

	userInfo, history, err := m.user.GetInfo(ctx, int(user.ID))
	if err != nil {
		ctx.Error("failed to get info user", err)
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}

	resp := makeInfoResp(userInfo, history)

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
