package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"
)

// waCallbackIDPrefix marks callback queries that were synthesized from
// WhatsApp interactive replies, so the outbound proxy can skip
// answerCallbackQuery calls for them.
const waCallbackIDPrefix = "wa_"

// WebhookHandler receives WhatsApp Cloud API webhooks, converts incoming
// messages into Telegram-shaped updates and feeds them into the bot's update
// pipeline via its webhook handler.
type WebhookHandler struct {
	TelegramHandler http.Handler
	VerifyToken     string
}

func NewWebhookHandler(tgHandler http.Handler) *WebhookHandler {
	verifyToken := os.Getenv("WHATSAPP_VERIFY_TOKEN")
	if verifyToken == "" {
		verifyToken = "langbot-whatsapp-verify"
	}
	return &WebhookHandler{
		TelegramHandler: tgHandler,
		VerifyToken:     verifyToken,
	}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		if mode == "subscribe" && token == h.VerifyToken {
			fmt.Println("WEBHOOK_VERIFIED")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(challenge))
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	entries, _ := payload["entry"].([]interface{})
	for _, entryInt := range entries {
		entry, _ := entryInt.(map[string]interface{})
		changes, _ := entry["changes"].([]interface{})
		for _, changeInt := range changes {
			change, _ := changeInt.(map[string]interface{})
			value, _ := change["value"].(map[string]interface{})

			messages, _ := value["messages"].([]interface{})
			for _, msgInt := range messages {
				msgObj, ok := msgInt.(map[string]interface{})
				if !ok {
					continue
				}
				if tgUpdate := h.convertMessage(msgObj, value); tgUpdate != nil {
					h.dispatch(tgUpdate)
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

// convertMessage maps a WhatsApp message to a Telegram-shaped update, or nil
// for unsupported message types (media, reactions, statuses...).
func (h *WebhookHandler) convertMessage(msgObj, value map[string]interface{}) map[string]interface{} {
	fromStr, _ := msgObj["from"].(string)
	fromInt, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil || fromInt == 0 {
		return nil
	}

	var contact map[string]interface{}
	if contacts, ok := value["contacts"].([]interface{}); ok && len(contacts) > 0 {
		contact, _ = contacts[0].(map[string]interface{})
	}

	tgUpdate := map[string]interface{}{
		"update_id": time.Now().UnixNano(),
	}

	msgType, _ := msgObj["type"].(string)

	switch msgType {
	case "text":
		textObj, _ := msgObj["text"].(map[string]interface{})
		body, _ := textObj["body"].(string)
		if body == "" {
			return nil
		}
		tgUpdate["message"] = buildTelegramMessageStruct(fromInt, contact, body)

	case "interactive":
		interactiveObj, _ := msgObj["interactive"].(map[string]interface{})
		interType, _ := interactiveObj["type"].(string)

		var callbackData string
		if interType == "button_reply" {
			if reply, ok := interactiveObj["button_reply"].(map[string]interface{}); ok {
				callbackData, _ = reply["id"].(string)
			}
		} else if interType == "list_reply" {
			if reply, ok := interactiveObj["list_reply"].(map[string]interface{}); ok {
				callbackData, _ = reply["id"].(string)
			}
		}
		if callbackData == "" {
			return nil
		}

		tgUpdate["callback_query"] = map[string]interface{}{
			"id": fmt.Sprintf("%s%d", waCallbackIDPrefix, time.Now().UnixNano()),
			"from": map[string]interface{}{
				"id":         fromInt,
				"is_bot":     false,
				"first_name": contactName(contact),
			},
			"message": buildTelegramMessageStruct(fromInt, contact, ""),
			"data":    callbackData,
		}

	default:
		return nil
	}

	return tgUpdate
}

func (h *WebhookHandler) dispatch(tgUpdate map[string]interface{}) {
	tgJSON, err := json.Marshal(tgUpdate)
	if err != nil {
		return
	}

	newReq, _ := http.NewRequest("POST", "/", bytes.NewReader(tgJSON))
	newReq.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	h.TelegramHandler.ServeHTTP(recorder, newReq)
}

func contactName(contact map[string]interface{}) string {
	if contact == nil {
		return "WA_User"
	}
	if profile, ok := contact["profile"].(map[string]interface{}); ok {
		if nm, ok := profile["name"].(string); ok && nm != "" {
			return nm
		}
	}
	return "WA_User"
}

func buildTelegramMessageStruct(fromID int64, contact map[string]interface{}, text string) map[string]interface{} {
	name := contactName(contact)

	parts := strings.Split(name, " ")
	first := parts[0]
	last := ""
	if len(parts) > 1 {
		last = parts[1]
	}

	return map[string]interface{}{
		"message_id": int(time.Now().Unix() % 1000000),
		"from": map[string]interface{}{
			"id":         fromID,
			"is_bot":     false,
			"first_name": first,
			"last_name":  last,
		},
		"chat": map[string]interface{}{
			"id":   fromID,
			"type": "private",
		},
		"date": time.Now().Unix(),
		"text": text,
	}
}
