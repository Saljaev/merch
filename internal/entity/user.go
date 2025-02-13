package entity

import (
	"github.com/sony/sonyflake"

	"log/slog"
)

var sf *sonyflake.Sonyflake
var log = slog.Default()

func InitSonyflake() {
	sf = sonyflake.NewSonyflake(sonyflake.Settings{})
	if sf == nil {
		log.Error("failed to init sonyflake")
	}
}

type User struct {
	ID        int64
	Coins     int
	UserName  string
	Inventory Inventory
}

type Inventory struct {
	UserID   int64
	Item     string
	quantity int
}

type CoinHistory struct {
	UserID   int64
	FromUser string
	ToUser   string
	Amount   int
	Type     string
}

//TODO: user session with token

func NewUser(coins int, username string) User {
	userID, err := sf.NextID()
	if err != nil {
		log.Error("failed to generate userID", slog.Any("error", err))
	}

	return User{
		ID:        int64(userID),
		Coins:     coins,
		UserName:  username,
		Inventory: Inventory{},
	}
}
