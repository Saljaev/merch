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

	username := RandomString(30)
	password := RandomString(30)
	password1 := RandomString(30)

	userReq, err := json.Marshal(authReq{
		Username: username,
		Password: password,
	})

	assert.NoError(t, err)

	userWithAnotherPass, err := json.Marshal(authReq{
		Username: username,
		Password: password1,
	})

	assert.NoError(t, err)

	userReqWithZeroUsername, err := json.Marshal(authReq{
		Username: "",
		Password: password,
	})

	assert.NoError(t, err)

	userReqWithZeroPass, err := json.Marshal(authReq{
		Username: username,
		Password: "",
	})

	assert.NoError(t, err)

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

			req, err = http.NewRequest(tt.auth.method, fmt.Sprintf("%s%s", tt.url, tt.auth.endPoint),
				bytes.NewBuffer(tt.auth.body))
			assert.NoError(t, err)

			client := &http.Client{}
			respAuth, err := client.Do(req)

			assert.NoError(t, err)
			defer func() {
				err = respAuth.Body.Close()
				assert.NoError(t, err)
			}()

			assert.Equal(t, tt.wantStatusAuth, respAuth.StatusCode)

			if respAuth.StatusCode != http.StatusOK {
				return
			}

			var authResponse authResp

			bodyBytes, err := io.ReadAll(respAuth.Body)
			assert.NoError(t, err)

			err = json.Unmarshal(bodyBytes, &authResponse)
			assert.NoError(t, err)

			assert.Equal(t, respAuth.StatusCode, tt.wantStatusAuth)
			assert.NotNil(t, authResponse)
		})
	}
}
