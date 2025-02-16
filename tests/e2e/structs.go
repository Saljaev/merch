package e2e

import (
	"math/rand"
)

type authReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResp struct {
	Token string `json:"token"`
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
