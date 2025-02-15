package merch

import (
	"merch/internal/api/handlers/auth"
	"merch/internal/usecase/usecase"
)

type MerchHandler struct {
	user usecase.UserUseCase
	jwt  *auth.JWTManager
}

func NewMerchHandler(u usecase.UserUseCase, j *auth.JWTManager) *MerchHandler {
	return &MerchHandler{
		user: u,
		jwt:  j,
	}
}
