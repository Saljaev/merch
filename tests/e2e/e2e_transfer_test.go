package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func createTransferReq(username string, amount int) []byte {
	type transferCoin struct {
		ToUser string `json:"toUser"`
		Amount int    `json:"amount"`
	}

	req := transferCoin{
		ToUser: username,
		Amount: amount,
	}

	data, _ := json.Marshal(&req)

	return data
}

func TestE2ETransfer(t *testing.T) {
	URL := "http://127.0.0.1:8080/api"

	username := "user"
	username1 := "user1"
	password := "secret"

	validCoins := 100
	toManyCoins := 10000
	invalidCoins := -1

	userReq, _ := json.Marshal(authReq{
		Username: username,
		Password: password,
	})

	user1Req, _ := json.Marshal(authReq{
		Username: username1,
		Password: password,
	})

	reqHeaders := map[string]string{
		"Authorization": "",
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", URL, "/auth"), bytes.NewBuffer((user1Req)))
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}

	type args struct {
		method   string
		endPoint string
		body     []byte
	}
	type auth struct {
		method   string
		endPoint string
		body     []byte
	}

	reqAuth := auth{
		method:   http.MethodPost,
		endPoint: "/auth",
		body:     userReq,
	}
	tests := []struct {
		name               string
		url                string
		args               args
		auth               auth
		headers            map[string]string
		wantAuth           bool
		wantStatusTransfer int
	}{
		{
			name: "Successful transfer coin",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/sendCoin",
				body:     createTransferReq(username1, validCoins),
			},
			auth:               reqAuth,
			headers:            reqHeaders,
			wantAuth:           true,
			wantStatusTransfer: http.StatusOK,
		},
		{
			name: "Transfer coins than more user's coins",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/sendCoin",
				body:     createTransferReq(username1, toManyCoins),
			},
			auth:               reqAuth,
			headers:            reqHeaders,
			wantAuth:           true,
			wantStatusTransfer: http.StatusBadRequest,
		},
		{
			name: "Transfer to yourself",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/sendCoin",
				body:     createTransferReq(username, validCoins),
			},
			auth:               reqAuth,
			headers:            reqHeaders,
			wantAuth:           true,
			wantStatusTransfer: http.StatusBadRequest,
		},
		{
			name: "Transfer to not existing user",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/sendCoin",
				body:     createTransferReq("not user", validCoins),
			},
			auth:               reqAuth,
			headers:            reqHeaders,
			wantAuth:           true,
			wantStatusTransfer: http.StatusBadRequest,
		},
		{
			name: "Transfer with invalid username",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/sendCoin",
				body:     createTransferReq("", validCoins),
			},
			auth:               reqAuth,
			headers:            reqHeaders,
			wantAuth:           true,
			wantStatusTransfer: http.StatusBadRequest,
		},
		{
			name: "Transfer with invalid coins amount",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/sendCoin",
				body:     createTransferReq(username1, invalidCoins),
			},
			auth:               reqAuth,
			headers:            reqHeaders,
			wantAuth:           true,
			wantStatusTransfer: http.StatusBadRequest,
		},
		{
			name: "Transfer with unsupported method",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/sendCoin",
				body:     createTransferReq(username1, invalidCoins),
			},
			auth:               reqAuth,
			headers:            reqHeaders,
			wantAuth:           true,
			wantStatusTransfer: http.StatusMethodNotAllowed,
		},
		{
			name: "Transfer without token",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/sendCoin",
				body:     createTransferReq(username1, invalidCoins),
			},
			auth:               reqAuth,
			headers:            map[string]string{},
			wantAuth:           false,
			wantStatusTransfer: http.StatusUnauthorized,
		},
		{
			name: "Transfer with invalid token",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/sendCoin",
				body:     createTransferReq(username1, invalidCoins),
			},
			auth:               reqAuth,
			headers:            map[string]string{},
			wantAuth:           true,
			wantStatusTransfer: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tt.wantAuth {
				req, err = http.NewRequest(tt.auth.method, fmt.Sprintf("%s%s", tt.url, tt.auth.endPoint), bytes.NewBuffer(tt.auth.body))

				assert.NoError(t, err)

				client := &http.Client{}
				respAuth, err := client.Do(req)

				assert.NoError(t, err)
				defer respAuth.Body.Close()

				var authResp struct {
					Token string `json:"token"`
				}

				bodyBytes, _ := io.ReadAll(respAuth.Body)

				err = json.Unmarshal(bodyBytes, &authResp)

				_, ok := tt.headers["Authorization"]
				if ok {
					tt.headers["Authorization"] = fmt.Sprintf("Bearer %s", authResp.Token)
				} else {
					tt.headers["Authorization"] = "Bearer token"
				}
			}

			req, err = http.NewRequest(tt.args.method, fmt.Sprintf("%s%s", tt.url, tt.args.endPoint), bytes.NewBuffer(tt.args.body))

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			assert.NoError(t, err)

			client1 := &http.Client{}
			respBuy, err := client1.Do(req)
			defer respBuy.Body.Close()

			assert.NoError(t, err)

			assert.Equal(t, tt.wantStatusTransfer, respBuy.StatusCode)
		})
	}
}
