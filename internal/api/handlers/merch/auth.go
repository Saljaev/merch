package merch

import (
	"errors"
	"merch/internal/api/utilapi"
	"merch/internal/entity"
	"net/http"
	"unicode/utf8"
)

var ErrIdentification = errors.New("incorrect username or password")

type AuthReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *AuthReq) IsValid() bool {
	return utf8.RuneCountInString(a.Username) > 0 && utf8.RuneCountInString(a.Password) > 0
}

func (m *MerchHandlder) Register(ctx *utilapi.APIContext) {
	var req AuthReq

	err := ctx.Decode(&req)
	if err != nil {
		ctx.Error("failed to encode request", err)
		ctx.WriteFailure(http.StatusBadRequest, "invalid request")
		return
	}

	user := m.getUser(req.Username)

	if user.UserName == "" {
		user, err = entity.NewUser(req.Username, req.Password)
		if err != nil {
			ctx.Error("failed to add user", err)
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
			return
		}

		_, err = m.user.AddUser(ctx, user)
		if err != nil {
			ctx.Error("failed to add user", err)
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
			return
		}
		m.cache.Set(user.UserName, user)
	} else {
		auth := user.Identification(req.Password)
		if !auth {
			ctx.Error("failed to identification user", ErrIdentification)
			ctx.WriteFailure(http.StatusBadRequest, "incorrect username or password")
			return
		}
	}
	
	customClaims := map[string]string{
		"username": user.UserName,
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
