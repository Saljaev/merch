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

func TestE2EAuth(t *testing.T) {
	URL := "http://127.0.0.1:8080/api"

	username := "testuser"
	password := "secret"
	password1 := "not-secret"

	userReq, _ := json.Marshal(authReq{
		Username: username,
		Password: password,
	})

	userWithAnotherPass, _ := json.Marshal(authReq{
		Username: username,
		Password: password1,
	})

	userReqWithZeroUsername, _ := json.Marshal(authReq{
		Username: "",
		Password: password,
	})

	userReqWithZeroPass, _ := json.Marshal(authReq{
		Username: username,
		Password: "",
	})

	type auth struct {
		method   string
		endPoint string
		body     []byte
	}
	tests := []struct {
		name           string
		url            string
		auth           auth
		wantStatusAuth int
	}{
		{
			name: "Successful add",
			url:  URL,
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				body:     userReq,
			},
			wantStatusAuth: http.StatusOK,
		},
		{
			name: "Try to login with incorrect password",
			url:  URL,
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				body:     userWithAnotherPass,
			},
			wantStatusAuth: http.StatusUnauthorized,
		},
		{
			name: "Unsupported method",
			url:  URL,
			auth: auth{
				method:   http.MethodGet,
				endPoint: "/auth",
			},
			wantStatusAuth: http.StatusMethodNotAllowed,
		},
		{
			name: "Try to login with invalid username",
			url:  URL,
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				body:     userReqWithZeroUsername,
			},
			wantStatusAuth: http.StatusBadRequest,
		},
		{
			name: "Try to login with invalid password",
			url:  URL,
			auth: auth{
				method:   http.MethodPost,
				endPoint: "/auth",
				body:     userReqWithZeroPass,
			},
			wantStatusAuth: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error

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

			assert.NoError(t, err)

			assert.Equal(t, respAuth.StatusCode, tt.wantStatusAuth)
			assert.NotNil(t, authResp)
		})
	}
}
