package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type sendMessageResponse struct {
	OK     bool               `json:"ok"`
	Result *sendMessageResult `json:"result"`
}

type sendMessageResult struct {
	MessageID int `json:"message_id"`
}

func buildSendMessageBody(chatID, text string) ([]byte, error) {

	request := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}

	body, err := json.Marshal(request)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal telegram request: %w", err)
	}

	return body, nil
}

func buildSendMessageRequest(token, chatID, text string) (*http.Request, error) {

	body, err := buildSendMessageBody(chatID, text)
	if err != nil {
		return nil, fmt.Errorf("failed to build telegram request body: %w", err)
	}

	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		token,
	)

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(body),
	)

	if err != nil {
		return nil, fmt.Errorf("failed create new request to telegram: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func doRequest(client *http.Client, req *http.Request) ([]byte, error) {

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send telegram request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read telegram response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, fmt.Errorf(
			"telegram returned status %s: %s",
			resp.Status,
			body,
		)
	}

	return body, nil
}

func SendMessage(client *http.Client, token, chatID, text string) (int, error) {

	req, err := buildSendMessageRequest(token, chatID, text)

	if err != nil {
		return 0, fmt.Errorf("failed to build send message request: %w", err)
	}

	body, err := doRequest(client, req)

	if err != nil {
		return 0, fmt.Errorf("failed to send telegram message: %w", err)
	}

	messageID, err := parseMessageID(body)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to parse telegram send message response: %w",
			err,
		)
	}

	return messageID, nil
}

func parseMessageID(body []byte) (int, error) {
	var entry sendMessageResponse

	if err := json.Unmarshal(body, &entry); err != nil {
		return 0, fmt.Errorf("failed unmarshal parse telegram response: %w", err)
	}

	if !entry.OK {
		return 0, fmt.Errorf(
			"telegram response is not ok",
		)
	}

	if entry.Result == nil {
		return 0, fmt.Errorf(
			"telegram response does not contain result",
		)
	}

	if entry.Result.MessageID == 0 {
		return 0, fmt.Errorf(
			"telegram response does not contain message_id",
		)
	}

	return entry.Result.MessageID, nil
}
