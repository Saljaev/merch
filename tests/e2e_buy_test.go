package tests

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

	type args struct {
		method   string
		endPoint string
		body     string
		item     map[string]int
	}
	type auth struct {
		method   string
		endPoint string
		useAuth  bool
		body     string
	}
	tests := []struct {
		name           string
		url            string
		args           args
		auth           auth
		headers        map[string]string
		wantStatusAuth int
		wantStatusBuy  int
		wantBody       string
	}{
		{
			name: "Successful buy 1 item",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				body:     "",
				item: map[string]int{
					"pen": 1,
				},
			},
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				useAuth:  true,
				body:     `{"username": "testuser", "password": "secret"}`,
			},
			headers:        map[string]string{},
			wantStatusAuth: http.StatusOK,
			wantStatusBuy:  http.StatusOK,
		},
		{
			name: "Successful buy many item",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				body:     "",
				item: map[string]int{
					"pen":       1,
					"book":      1,
					"socks":     2,
					"wallet":    1,
					"powerbank": 4,
				},
			},
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				useAuth:  true,
				body:     `{"username": "testuser", "password": "secret"}`,
			},
			headers:        map[string]string{},
			wantStatusAuth: http.StatusOK,
			wantStatusBuy:  http.StatusOK,
		},
		{
			name: "Try to buy to many items",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				body:     "",
				item: map[string]int{
					"powerbank": 10,
				},
			},
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				useAuth:  true,
				body:     `{"username": "testuser", "password": "secret"}`,
			},
			headers:        map[string]string{},
			wantStatusAuth: http.StatusOK,
			wantStatusBuy:  http.StatusBadRequest,
		},
		{
			name: "Try to buy without auth",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				body:     "",
				item: map[string]int{
					"pen": 1,
				},
			},
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				useAuth:  false,
				body:     `{"username": "testuser", "password": "secret"}`,
			},
			headers:       map[string]string{},
			wantStatusBuy: http.StatusUnauthorized,
		},
		{
			name: "Try to login with incorrect password",
			url:  URL,
			args: args{
				method:   http.MethodGet,
				endPoint: "/buy",
				body:     "",
				item: map[string]int{
					"pen": 1,
				},
			},
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				useAuth:  true,
				body:     `{"username": "testuser", "password": "not-secret"}`,
			},
			headers:        map[string]string{},
			wantStatusAuth: http.StatusUnauthorized,
		},
		{
			name: "Try to buy unsupported method /auth",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/buy",
				body:     "",
				item: map[string]int{
					"pen": 1,
				},
			},
			auth: auth{
				method:   http.MethodGet,
				endPoint: "/auth",
				useAuth:  true,
				body:     `{"username": "testuser", "password": "secret"}`,
			},
			headers:        map[string]string{},
			wantStatusAuth: http.StatusMethodNotAllowed,
		},
		{
			name: "Try to buy unsupported method /buy",
			url:  URL,
			args: args{
				method:   http.MethodPost,
				endPoint: "/buy",
				body:     "",
				item: map[string]int{
					"pen": 1,
				},
			},
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				useAuth:  true,
				body:     `{"username": "testuser", "password": "secret"}`,
			},
			headers:        map[string]string{},
			wantStatusAuth: http.StatusOK,
			wantStatusBuy:  http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var req *http.Request
			var err error

			if tt.auth.useAuth {
				req, err = http.NewRequest(tt.auth.method, fmt.Sprintf("%s%s", tt.url, tt.auth.endPoint), bytes.NewBuffer([]byte(tt.auth.body)))

				assert.NoError(t, err)

				client := &http.Client{}
				respAuth, err := client.Do(req)

				assert.NoError(t, err)
				defer respAuth.Body.Close()

				assert.Equal(t, tt.wantStatusAuth, respAuth.StatusCode)

				if respAuth.StatusCode != http.StatusOK {
					return
				}

				var authResp struct {
					Token string `json:"token"`
				}

				bodyBytes, _ := io.ReadAll(respAuth.Body)

				err = json.Unmarshal(bodyBytes, &authResp)

				tt.headers["Authorization"] = fmt.Sprintf("Bearer %s", authResp.Token)
			}

			for item, count := range tt.args.item {
				for i := 0; i < count; i++ {
					req, err = http.NewRequest(tt.args.method, fmt.Sprintf("%s%s/?item=%s", tt.url, tt.args.endPoint, item), bytes.NewBuffer([]byte(tt.args.body)))

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
