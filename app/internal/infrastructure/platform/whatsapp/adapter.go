package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type ProxyClient struct {
	Transport           http.RoundTripper
	WhatsAppToken       string
	WhatsAppPhoneNumber string
}

func NewProxyClient() *ProxyClient {
	return &ProxyClient{
		Transport:           http.DefaultTransport,
		WhatsAppToken:       os.Getenv("WHATSAPP_TOKEN"),
		WhatsAppPhoneNumber: os.Getenv("WHATSAPP_PHONE_ID"),
	}
}

func (p *ProxyClient) RoundTrip(req *http.Request) (*http.Response, error) {

	if strings.Contains(req.URL.Host, "api.telegram.org") {

		var bodyBytes []byte
		if req.Body != nil {
			bodyBytes, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		if len(bodyBytes) > 0 {
			var payload map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &payload); err == nil {
				if chatID, ok := payload["chat_id"].(float64); ok && isWhatsAppID(int64(chatID)) {

					if strings.Contains(req.URL.Path, "/sendMessage") {
						return p.sendWhatsAppMessage(int64(chatID), payload)
					} else if strings.Contains(req.URL.Path, "/editMessageText") {
						return p.sendWhatsAppMessage(int64(chatID), payload)
					} else if strings.Contains(req.URL.Path, "/deleteMessage") {

						return fakeTelegramResponse(true), nil
					}
				}
			}
		}
	}

	return p.Transport.RoundTrip(req)
}

func isWhatsAppID(id int64) bool {

	return id > 10000000000
}

func fakeTelegramResponse(success bool) *http.Response {
	respBody := `{"ok": true, "result": {"message_id": 1}}`
	if !success {
		respBody = `{"ok": false, "error_code": 400, "description": "Failed"}`
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(respBody)),
		Header:     make(http.Header),
	}
}

func (p *ProxyClient) sendWhatsAppMessage(chatID int64, payload map[string]interface{}) (*http.Response, error) {
	if p.WhatsAppToken == "" || p.WhatsAppPhoneNumber == "" {
		return fakeTelegramResponse(false), fmt.Errorf("WhatsApp credentials missing")
	}

	url := fmt.Sprintf("https://graph.facebook.com/v17.0/%s/messages", p.WhatsAppPhoneNumber)
	text, _ := payload["text"].(string)

	var waBody map[string]interface{}

	if replyMarkup, ok := payload["reply_markup"].(map[string]interface{}); ok {
		if inlineKeyboard, hasKB := replyMarkup["inline_keyboard"].([]interface{}); hasKB && len(inlineKeyboard) > 0 {
			waBody = buildWhatsAppInteractive(fmt.Sprintf("%d", chatID), text, inlineKeyboard)
		} else if keyboard, hasReply := replyMarkup["keyboard"].([]interface{}); hasReply && len(keyboard) > 0 {
			waBody = buildWhatsAppInteractive(fmt.Sprintf("%d", chatID), text, keyboard)
		} else {
			waBody = buildWhatsAppText(fmt.Sprintf("%d", chatID), text)
		}
	} else {
		waBody = buildWhatsAppText(fmt.Sprintf("%d", chatID), text)
	}

	waBytes, err := json.Marshal(waBody)
	if err != nil {
		return fakeTelegramResponse(false), err
	}

	waReq, _ := http.NewRequest("POST", url, bytes.NewBuffer(waBytes))
	waReq.Header.Set("Authorization", "Bearer "+p.WhatsAppToken)
	waReq.Header.Set("Content-Type", "application/json")

	waResp, err := p.Transport.RoundTrip(waReq)
	if err != nil {
		return fakeTelegramResponse(false), err
	}
	defer waResp.Body.Close()

	return fakeTelegramResponse(true), nil
}

func buildWhatsAppText(to string, text string) map[string]interface{} {
	return map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "text",
		"text": map[string]interface{}{
			"preview_url": false,
			"body":        text,
		},
	}
}

func buildWhatsAppInteractive(to string, text string, tgKeyboard []interface{}) map[string]interface{} {

	buttons := []map[string]interface{}{}
	var listItems []map[string]interface{}

	for _, rowInt := range tgKeyboard {
		row, ok := rowInt.([]interface{})
		if !ok {
			continue
		}
		for _, btnInt := range row {
			btn, ok := btnInt.(map[string]interface{})
			if !ok {
				continue
			}
			textObj := btn["text"].(string)
			var callback string
			if cb, hasCb := btn["callback_data"].(string); hasCb {
				callback = cb
			} else {
				callback = textObj
			}

			if len(callback) > 255 {
				callback = callback[:255]
			}

			buttons = append(buttons, map[string]interface{}{
				"type": "reply",
				"reply": map[string]interface{}{
					"id":    callback,
					"title": truncate(textObj, 20),
				},
			})

			listItems = append(listItems, map[string]interface{}{
				"id":          callback,
				"title":       truncate(textObj, 24),
				"description": "",
			})
		}
	}

	if len(buttons) <= 3 && len(buttons) > 0 {
		return map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                to,
			"type":              "interactive",
			"interactive": map[string]interface{}{
				"type": "button",
				"body": map[string]interface{}{
					"text": text,
				},
				"action": map[string]interface{}{
					"buttons": buttons,
				},
			},
		}
	}

	return map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]interface{}{
			"type": "list",
			"body": map[string]interface{}{
				"text": text,
			},
			"action": map[string]interface{}{
				"button": "Options",
				"sections": []map[string]interface{}{
					{
						"title": "Available Actions",
						"rows":  listItems,
					},
				},
			},
		},
	}
}

func truncate(s string, l int) string {
	if len(s) > l {
		return s[:l]
	}
	return s
}
