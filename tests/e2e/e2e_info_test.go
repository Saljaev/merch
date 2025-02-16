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

func TestE2EInfoInventory(t *testing.T) {
	URL := "http://127.0.0.1:8080/api"

	username := RandomString(30)
	password := RandomString(30)

	userReq, err := json.Marshal(authReq{
		Username: username,
		Password: password,
	})
	assert.NoError(t, err)

	store := map[string]int{
		"pen":    3,
		"wallet": 1,
	}

	headers := map[string]string{}
	err = buyItems(store, userReq, URL, headers)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", URL, "/info"), nil)
	assert.NoError(t, err)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	dataInfo, err := client.Do(req)
	assert.NoError(t, err)

	type Inventory struct {
		Type     string `json:"type"`
		Quantity int    `json:"quantity"`
	}

	type History struct {
		FromUser string `json:"fromUser,omitempty"`
		ToUser   string `json:"toUser,omitempty"`
		Amount   int    `json:"amount"`
	}

	type CoinHistory struct {
		Received []History `json:"received"`
		Sent     []History `json:"sent"`
	}

	var respInfo struct {
		Coins       int         `json:"coins"`
		Inventory   []Inventory `json:"inventory"`
		CoinHistory CoinHistory `json:"coinHistory"`
	}

	bodyBytes, err := io.ReadAll(dataInfo.Body)
	assert.NoError(t, err)

	err = json.Unmarshal(bodyBytes, &respInfo)
	assert.NoError(t, err)

	for _, item := range respInfo.Inventory {
		_, ok := store[item.Type]
		assert.NotEqual(t, ok, false)
	}
}

func TestE2EInfoTransaction(t *testing.T) {
	URL := "http://127.0.0.1:8080/api"

	username := RandomString(30)
	username1 := RandomString(30)
	password := RandomString(30)

	userReq, err := json.Marshal(authReq{
		Username: username,
		Password: password,
	})
	assert.NoError(t, err)

	user1Req, err := json.Marshal(authReq{
		Username: username1,
		Password: password,
	})
	assert.NoError(t, err)

	headers := map[string]string{}
	err = doTransfer(username, URL, headers, userReq, user1Req, 100)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", URL, "/info"), nil)
	assert.NoError(t, err)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	dataInfo, err := client.Do(req)
	assert.NoError(t, err)

	type Inventory struct {
		Type     string `json:"type"`
		Quantity int    `json:"quantity"`
	}

	type History struct {
		FromUser string `json:"fromUser,omitempty"`
		ToUser   string `json:"toUser,omitempty"`
		Amount   int    `json:"amount"`
	}

	type CoinHistory struct {
		Received []History `json:"received"`
		Sent     []History `json:"sent"`
	}

	var respInfo struct {
		Coins       int         `json:"coins"`
		Inventory   []Inventory `json:"inventory"`
		CoinHistory CoinHistory `json:"coinHistory"`
	}

	bodyBytes, err := io.ReadAll(dataInfo.Body)
	assert.NoError(t, err)

	err = json.Unmarshal(bodyBytes, &respInfo)
	assert.NoError(t, err)

	assert.NotNil(t, t, respInfo)
}

func doTransfer(toUsername, URL string, headers map[string]string, toUser, fromUser []byte, amount int) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", URL, "/auth"), bytes.NewBuffer((toUser)))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", URL, "/auth"), bytes.NewBuffer((fromUser)))
	if err != nil {
		return err
	}

	client = &http.Client{}
	respAuth, err := client.Do(req)
	if err != nil {
		return err
	}

	var authResponse authResp
	bodyBytes, err := io.ReadAll(respAuth.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyBytes, &authResponse)
	if err != nil {
		return err
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %s", authResponse.Token)

	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", URL, "/sendCoin"),
		bytes.NewBuffer(createTransferReq(toUsername, amount)))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if err != nil {
		return err
	}

	client = &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func buyItems(store map[string]int, user []byte, URL string, headers map[string]string) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", URL, "/auth"), bytes.NewBuffer((user)))
	if err != nil {
		return err
	}

	client := &http.Client{}
	respAuth, err := client.Do(req)
	if err != nil {
		return err
	}

	var authResponse authResp
	bodyBytes, err := io.ReadAll(respAuth.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodyBytes, &authResponse)
	if err != nil {
		return err
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %s", authResponse.Token)

	for item, count := range store {
		for i := 0; i < count; i++ {
			req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s/?item=%s", URL, "/buy", item), nil)
			if err != nil {
				return err
			}

			for k, v := range headers {
				req.Header.Set(k, v)
			}

			client1 := &http.Client{}
			respBuy, err := client1.Do(req)
			defer func() {
				err = respBuy.Body.Close()
				if err != nil {
					return
				}
			}()

			if err != nil {
				return err
			}
		}
	}

	return nil
}
