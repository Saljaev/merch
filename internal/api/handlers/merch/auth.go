package merch

import (
	"errors"
	"fmt"
	"merch/internal/api/utilapi"
	"merch/internal/usecase"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	ErrNoAuthorization = errors.New("not Authorization header")
)

type AuthReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *AuthReq) IsValid() bool {
	return utf8.RuneCountInString(a.Username) > 0 && utf8.RuneCountInString(a.Password) > 0
}

func (m *MerchHandler) Login(ctx *utilapi.APIContext) {
	var req AuthReq

	err := ctx.Decode(&req)
	if err != nil {
		ctx.Error("failed to encode request", err)
		ctx.WriteFailure(http.StatusBadRequest, "invalid request")
		return
	}

	id, err := m.user.AddUser(ctx, req.Username, req.Password)
	if err != nil {
		ctx.Error("failed to add user", err)
		if errors.Is(errors.Unwrap(err), usecase.ErrNotAuthorization) {
			ctx.WriteFailure(http.StatusUnauthorized, "incorrect username or password")
			return
		} else {
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
			return
		}
	}

	customClaims := map[string]string{
		"id": strconv.Itoa(id),
	}

	token, err := m.jwt.Generate(req.Username, customClaims)
	if err != nil {
		ctx.Error("failed to generate JWT token", err)
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}

	ctx.Info("successful generate jwt token for", "user", req.Username)
	ctx.SuccessWithData(map[string]string{
		"token": token,
	})
}

func (m *MerchHandler) TokenMiddleware(ctx *utilapi.APIContext) {
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
