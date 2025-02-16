package entity

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/sony/sonyflake"
	"log/slog"
)

const DefaultCoins = 1000

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
	Password  string
	Coins     int
	UserName  string
	Inventory []Inventory
}

type Inventory struct {
	UserID   int64
	Item     string
	Quantity int
}

type CoinHistory struct {
	FromUser string
	ToUser   string
	Amount   int
	Type     string
}

func NewUser(username, password string) (User, error) {
	hashedPassword := HashPassword(password)
	
	userID, err := sf.NextID()
	if err != nil {
		return User{}, fmt.Errorf("failed to generate userID: %w", err)
	}

	return User{
		ID:        int64(userID),
		Coins:     DefaultCoins,
		UserName:  username,
		Password:  hashedPassword,
		Inventory: []Inventory{},
	}, nil
}

func HashPassword(password string) string {
	hash := md5.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (u *User) Identification(password string) bool {
	return HashPassword(password) == u.Password
}
