package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"
)

type WebhookHandler struct {
	TelegramHandler http.Handler
	VerifyToken     string
}

func NewWebhookHandler(tgHandler http.Handler) *WebhookHandler {
	return &WebhookHandler{
		TelegramHandler: tgHandler,
		VerifyToken:     "langbot-whatsapp-verify",
	}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

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

	if r.Method == "POST" {
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

				if messages, ok := value["messages"].([]interface{}); ok && len(messages) > 0 {
					msgObj := messages[0].(map[string]interface{})
					contacts := value["contacts"].([]interface{})
					contact := contacts[0].(map[string]interface{})

					fromStr, _ := msgObj["from"].(string)
					fromInt, _ := strconv.ParseInt(fromStr, 10, 64)

					msgType, _ := msgObj["type"].(string)

					tgUpdate := map[string]interface{}{}
					tgUpdate["update_id"] = time.Now().Unix()

					if msgType == "text" {
						textObj := msgObj["text"].(map[string]interface{})
						body := textObj["body"].(string)

						tgUpdate["message"] = buildTelegramMessageStruct(fromInt, contact, body)

					} else if msgType == "interactive" {

						interactiveObj := msgObj["interactive"].(map[string]interface{})
						interType := interactiveObj["type"].(string)
						var callbackData string
						if interType == "button_reply" {
							callbackData = interactiveObj["button_reply"].(map[string]interface{})["id"].(string)
						} else if interType == "list_reply" {
							callbackData = interactiveObj["list_reply"].(map[string]interface{})["id"].(string)
						}

						tgUpdate["callback_query"] = map[string]interface{}{
							"id": fmt.Sprintf("%d", time.Now().Unix()),
							"from": map[string]interface{}{
								"id":         fromInt,
								"is_bot":     false,
								"first_name": contact["profile"].(map[string]interface{})["name"],
							},
							"message": buildTelegramMessageStruct(fromInt, contact, ""),
							"data":    callbackData,
						}
					}

					if len(tgUpdate) > 1 {

						tgJSON, _ := json.Marshal(tgUpdate)

						newReq, _ := http.NewRequest("POST", "/", bytes.NewReader(tgJSON))
						newReq.Header.Set("Content-Type", "application/json")

						recorder := httptest.NewRecorder()
						h.TelegramHandler.ServeHTTP(recorder, newReq)
					}
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}

func buildTelegramMessageStruct(fromID int64, contact map[string]interface{}, text string) map[string]interface{} {
	name := "WA_User"
	if profile, ok := contact["profile"].(map[string]interface{}); ok {
		if nm, _ok := profile["name"].(string); _ok {
			name = nm
		}
	}

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
