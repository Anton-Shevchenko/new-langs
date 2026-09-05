package whatsapp

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
)

func multipartRequest(t *testing.T, url string, fields map[string]string) *http.Request {
	t.Helper()
	buf := bytes.NewBuffer(nil)
	form := multipart.NewWriter(buf)
	for k, v := range fields {
		if err := form.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	form.Close()

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", form.FormDataContentType())
	return req
}

// roundTripRecorder captures the request the proxy forwards to the transport.
type roundTripRecorder struct {
	req  *http.Request
	body []byte
}

func (r *roundTripRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	r.req = req
	if req.Body != nil {
		r.body, _ = io.ReadAll(req.Body)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{}`))),
		Header:     make(http.Header),
	}, nil
}

func TestProxyReroutesWhatsAppSendMessage(t *testing.T) {
	recorder := &roundTripRecorder{}
	p := &ProxyClient{Transport: recorder, WhatsAppToken: "token", WhatsAppPhoneNumber: "12345"}

	req := multipartRequest(t, "https://api.telegram.org/botTOKEN/sendMessage", map[string]string{
		"chat_id":    "380671234567",
		"text":       "<b>Hallo</b> Welt",
		"parse_mode": "HTML",
	})

	resp, err := p.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if recorder.req == nil || recorder.req.URL.Host != "graph.facebook.com" {
		t.Fatalf("expected request to graph.facebook.com, got %v", recorder.req)
	}

	var waBody map[string]interface{}
	if err := json.Unmarshal(recorder.body, &waBody); err != nil {
		t.Fatal(err)
	}
	if waBody["to"] != "380671234567" {
		t.Errorf("wrong recipient: %v", waBody["to"])
	}
	text := waBody["text"].(map[string]interface{})["body"].(string)
	if text != "*Hallo* Welt" {
		t.Errorf("HTML not converted: %q", text)
	}
}

func TestProxyPassesThroughTelegramChats(t *testing.T) {
	recorder := &roundTripRecorder{}
	p := &ProxyClient{Transport: recorder, WhatsAppToken: "token", WhatsAppPhoneNumber: "12345"}

	req := multipartRequest(t, "https://api.telegram.org/botTOKEN/sendMessage", map[string]string{
		"chat_id": "123456",
		"text":    "hi",
	})

	resp, err := p.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if recorder.req == nil || recorder.req.URL.Host != "api.telegram.org" {
		t.Fatalf("telegram request should pass through, got %v", recorder.req)
	}
	if len(recorder.body) == 0 {
		t.Error("forwarded request lost its body")
	}
}

func TestExtractKeyboardButtons(t *testing.T) {
	inline := `{"inline_keyboard":[[{"text":"Yes","callback_data":"test_yes"}],[{"text":"No","callback_data":"test_no"}]]}`
	buttons := extractKeyboardButtons(inline)
	if len(buttons) != 2 || buttons[0].ID != "test_yes" || buttons[1].Title != "No" {
		t.Errorf("unexpected inline buttons: %+v", buttons)
	}

	replyKb := `{"keyboard":[[{"text":"📋 List"},{"text":"Test me"}]]}`
	buttons = extractKeyboardButtons(replyKb)
	if len(buttons) != 2 || buttons[0].ID != "📋 List" {
		t.Errorf("unexpected reply buttons: %+v", buttons)
	}
}

func TestBuildWhatsAppInteractiveUsesListForManyButtons(t *testing.T) {
	var buttons []waButton
	for i := 0; i < 6; i++ {
		buttons = append(buttons, waButton{ID: string(rune('a' + i)), Title: string(rune('A' + i))})
	}
	body := buildWhatsAppInteractive("123", "pick one", buttons)
	interactive := body["interactive"].(map[string]interface{})
	if interactive["type"] != "list" {
		t.Errorf("expected list type, got %v", interactive["type"])
	}

	body = buildWhatsAppInteractive("123", "pick one", buttons[:3])
	interactive = body["interactive"].(map[string]interface{})
	if interactive["type"] != "button" {
		t.Errorf("expected button type, got %v", interactive["type"])
	}
}

func TestHTMLToWhatsApp(t *testing.T) {
	in := "🇩🇪 <b>Haus</b>\n<i>das</i> &amp; <code>x</code><br><u>y</u>"
	got := htmlToWhatsApp(in)
	want := "🇩🇪 *Haus*\n_das_ & ```x```\ny"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
