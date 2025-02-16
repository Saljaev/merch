package entity

import (
	"fmt"
	"github.com/sony/sonyflake"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"math"
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
	ID int64

	Coins int

	Password string
	UserName string

	Inventory []Inventory
}

type Inventory struct {
	Quantity int

	UserID int64

	Item string
}

type CoinHistory struct {
	Amount int

	FromUser string
	ToUser   string
	Type     string
}

func NewUser(username, password string) (User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}

	userID, err := sf.NextID()
	if err != nil {
		return User{}, fmt.Errorf("failed to generate userID: %w", err)
	}

	if userID > math.MaxInt64 {
		return User{}, err
	}

	return User{
		ID:        int64(userID),
		Coins:     DefaultCoins,
		UserName:  username,
		Password:  hashedPassword,
		Inventory: []Inventory{},
	}, nil
}

func HashPassword(password string) (string, error) {
	//hash := md5.Sum([]byte(password))
	//return hex.EncodeToString(hash[:])
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashPassword), nil
}

func (u *User) Identification(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
	//return HashPassword(password) == u.Password
}
