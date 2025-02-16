package e2e

import (
	"crypto/rand"
	"math/big"
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
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil {
			return ""
		}
		b[i] = letterBytes[num.Int64()]
	}
	return string(b)
}
