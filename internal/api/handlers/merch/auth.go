package merch

import (
	"errors"
	"fmt"
	"merch/internal/api/utilapi"
	"merch/internal/entity"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrIdentification   = errors.New("incorrect username or password")
	ErrNoAuthorization  = errors.New("not Authorization header")
	ErrInvalidCacheData = errors.New("invalid cache data")
)

type AuthReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *AuthReq) IsValid() bool {
	return utf8.RuneCountInString(a.Username) > 0 && utf8.RuneCountInString(a.Password) > 0
}

func (m *MerchHandlder) Login(ctx *utilapi.APIContext) {
	var req AuthReq

	err := ctx.Decode(&req)
	if err != nil {
		ctx.Error("failed to encode request", err)
		ctx.WriteFailure(http.StatusBadRequest, "invalid request")
		return
	}

	var user entity.User
	cachedUser, ok := m.cache.Get(req.Username)
	if !ok {

		user, err = entity.NewUser(req.Username, req.Password)
		if err != nil {
			ctx.Error("failed to create user", err)
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
			return
		}

		_, err = m.user.AddUser(ctx, user)
		if err != nil {
			ctx.Error("failed to add user to DB", err)
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
			return
		}
		m.cache.Set(req.Username, user)

	} else {
		user, ok = cachedUser.(entity.User)
		if !ok {
			ctx.Error("invalid user in cache", ErrInvalidCacheData)
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
			return
		}
	}

	auth := user.Identification(req.Password)
	if !auth {
		ctx.Error("failed to user authorization", ErrIdentification)
		ctx.WriteFailure(http.StatusUnauthorized, "incorrect username/password")
		return
	}

	customClaims := map[string]string{
		"id": strconv.Itoa(int(user.ID)),
	}

	token, err := m.jwt.Generate(user, customClaims)
	if err != nil {
		ctx.Error("failed to generate JWT token", err)
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}

	ctx.SuccessWithData(map[string]string{
		"token": token,
	})
}

func (m *MerchHandlder) TokenMiddleware(ctx *utilapi.APIContext) {
	authHeader := ctx.GetFromHeader("Authorization")
	if authHeader == "" {
		ctx.Error("unauthorized request", ErrNoAuthorization)
		ctx.WriteFailure(http.StatusUnauthorized, "login to your account")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	sub, claims := m.jwt.Parse(tokenString, true)
	if sub == "" {
		ctx.Error("failed to parse token", fmt.Errorf("error %v", claims["error"]))
		ctx.WriteFailure(http.StatusUnauthorized, "login to your account")
		return
	}

	ctx.SetValue("username", claims["sub"])
	ctx.SetValue("id", claims["id"])
}
