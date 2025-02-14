package entity

import (
	"fmt"
	"github.com/sony/sonyflake"
	"golang.org/x/crypto/bcrypt"
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := sf.NextID()
	if err != nil {
		return User{}, fmt.Errorf("failed to generate userID: %w", err)
	}

	return User{
		ID:        int64(userID),
		Coins:     DefaultCoins,
		UserName:  username,
		Password:  string(hashedPassword),
		Inventory: []Inventory{},
	}, nil
}

func (u *User) Identification(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
