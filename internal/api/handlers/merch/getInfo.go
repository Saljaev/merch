package merch

import "merch/internal/api/utilapi"

func (m *MerchHandlder) Info(api *utilapi.APIContext) {
	//id := r.Header.Get("id")
	//
	//	userID, _ := strconv.Atoi(id)
	//
	//	fmt.Println("USERID", id)
	//
	//	user, history, err := repo.GetInfo(context.Background(), userID)
	//	if err != nil {
	//		log.Error("failed to get user info", slog.Any("error", err))
	//	}
	//
	//	type Transaction struct {
	//		FromUser string `json:"fromUser,omitempty"` // Только для received
	//		ToUser   string `json:"toUser,omitempty"`   // Только для sent
	//		Amount   int    `json:"amount"`
	//	}
	//
	//	type Item struct {
	//		Type     string `json:"type"`
	//		Quantity int    `json:"quantity"`
	//	}
	//
	//	type CoinHistory struct {
	//		Received []Transaction `json:"received"`
	//		Sent     []Transaction `json:"sent"`
	//	}
	//
	//	type UserResponse struct {
	//		Coins       int         `json:"coins"`
	//		Inventory   []Item      `json:"inventory"`
	//		CoinHistory CoinHistory `json:"coinHistory"`
	//	}
	//
	//	resp := UserResponse{
	//		Coins:     user.Coins,
	//		Inventory: make([]Item, len(user.Inventory)),
	//	}
	//
	//	items := []Item{}
	//
	//	for _, v := range user.Inventory {
	//		item := Item{
	//			Type:     v.Item,
	//			Quantity: v.Quantity,
	//		}
	//
	//		items = append(items, item)
	//	}
	//
	//	received := []Transaction{}
	//	sent := []Transaction{}
	//
	//	for _, v := range history {
	//		tran := Transaction{Amount: v.Amount}
	//
	//		if v.Type == "sent" {
	//			tran.ToUser = v.ToUser
	//			sent = append(sent, tran)
	//		} else {
	//			tran.FromUser = v.FromUser
	//			received = append(received, tran)
	//		}
	//
	//	}
	//
	//	resp.CoinHistory.Sent = sent
	//	resp.CoinHistory.Received = received
	//	resp.Inventory = items
	//
	//	data, err := json.Marshal(resp)
	//
	//	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//	w.Write(data)
	//	w.WriteHeader(http.StatusOK)
	//	return
}
