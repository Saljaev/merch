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

func TestE2EBuy(t *testing.T) {
	URL := "http://127.0.0.1:8080/api"

	username := "testuser"
	password := "secret"

	userReq, _ := json.Marshal(authReq{
		Username: username,
		Password: password,
	})

	reqHeaders := map[string]string{
		"Authorization": "",
	}

	type args struct {
		method   string
		endPoint string
		item     map[string]int
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
		name          string
		url           string
		args          args
		auth          auth
		headers       map[string]string
		wantAuth      bool
		wantStatusBuy int
	}{
		{
			name: "Successful buy 1 item",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				item: map[string]int{
					"pen": 1,
				},
			},
			auth:          reqAuth,
			headers:       reqHeaders,
			wantAuth:      true,
			wantStatusBuy: http.StatusOK,
		},
		{
			name: "Successful buy many item",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				item: map[string]int{
					"pen":       1,
					"book":      1,
					"socks":     2,
					"wallet":    1,
					"powerbank": 4,
				},
			},
			auth:          reqAuth,
			headers:       reqHeaders,
			wantAuth:      true,
			wantStatusBuy: http.StatusOK,
		},
		{
			name: "Try to buy item that costs more than user's coins",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				item: map[string]int{
					"powerbank": 10,
				},
			},
			auth:          reqAuth,
			headers:       reqHeaders,
			wantAuth:      true,
			wantStatusBuy: http.StatusBadRequest,
		},
		{
			name: "Try to buy item that not in shop",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				item: map[string]int{
					"item-not-in-shop": 1,
				},
			},
			auth:          reqAuth,
			headers:       reqHeaders,
			wantAuth:      true,
			wantStatusBuy: http.StatusBadRequest,
		},
		{
			name: "Try to buy without token",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				item: map[string]int{
					"pen": 1,
				},
			},
			auth:          reqAuth,
			headers:       map[string]string{},
			wantAuth:      false,
			wantStatusBuy: http.StatusUnauthorized,
		},
		{
			name: "Try to buy with invalid token",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				item: map[string]int{
					"pen": 1,
				},
			},
			auth:          reqAuth,
			headers:       map[string]string{},
			wantAuth:      true,
			wantStatusBuy: http.StatusUnauthorized,
		},
		{
			name: "Try to buy unsupported method /buy",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/buy",
				item: map[string]int{
					"pen": 1,
				},
			},
			auth:          reqAuth,
			headers:       reqHeaders,
			wantAuth:      true,
			wantStatusBuy: http.StatusMethodNotAllowed,
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

				bodyBytes, _ := io.ReadAll(respAuth.Body)

				var authResponse authResp

				err = json.Unmarshal(bodyBytes, &authResponse)

				_, ok := tt.headers["Authorization"]
				if ok {
					tt.headers["Authorization"] = fmt.Sprintf("Bearer %s", authResponse.Token)
				} else {
					tt.headers["Authorization"] = "Bearer token"
				}
			}

			for item, count := range tt.args.item {
				for i := 0; i < count; i++ {
					req, err = http.NewRequest(tt.args.method, fmt.Sprintf("%s%s/?item=%s", tt.url, tt.args.endPoint, item), nil)

					for k, v := range tt.headers {
						req.Header.Set(k, v)
					}

					assert.NoError(t, err)

					client1 := &http.Client{}
					respBuy, err := client1.Do(req)
					defer respBuy.Body.Close()

					assert.NoError(t, err)

					assert.Equal(t, tt.wantStatusBuy, respBuy.StatusCode)
				}
			}
		})
	}
}
