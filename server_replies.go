package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/agronskiy/telegram-page-checker-bot/internal/config"
	"github.com/agronskiy/telegram-page-checker-bot/internal/pipres"
)

// Create a struct to conform to the JSON body
// of the send message request
// https://core.telegram.org/bots/api#sendmessage
type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func sayResult(singleUrl *config.SingleURL,
	adminId int64,
	result pipres.PipelineResult,
	needSayNegative bool,
) error {
	if result != pipres.SlotAvailable && !needSayNegative {
		return nil
	}

	// Create the request body struct
	var text string
	if result == pipres.SlotAvailable {
		text = fmt.Sprintf("✅ Есть слот, беги регистрироваться!\n🆔: %s\n🔗: %s", singleUrl.Name, singleUrl.Url)
	} else if result == pipres.SlotNotAvailable {
		text = fmt.Sprintf("🤷 Слотов пока нет, ждем...\n🆔: %s\n🔗: %s", singleUrl.Name, singleUrl.Url)
	} else if result == pipres.NoRescheduleTasks {
		text = fmt.Sprintf("🤔 Не нашел слотов для переноса!\n🆔: %s\n🔗: %s", singleUrl.Name, singleUrl.Url)
	} else {
		text = fmt.Sprintf("🤔 Возможно, слот уже зарегистрирован?\n🆔: %s\n🔗: %s", singleUrl.Name, singleUrl.Url)
	}

	for _, recipient_id := range []int64{singleUrl.UserID, adminId} {
		reqBody := &sendMessageReqBody{
			ChatID: recipient_id,
			Text:   text,
		}
		// Create the JSON body from the struct
		reqBytes, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}

		// Send a post request with your token
		var bot_url string = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.ApiKey)
		res, err := http.Post(bot_url, "application/json", bytes.NewBuffer(reqBytes))
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return errors.New("unexpected status" + res.Status)
		}
	}

	return nil
}
