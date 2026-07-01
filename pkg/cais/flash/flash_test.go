package flash

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func base64Decode(v string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(v)
}

func TestSetAndConsume(t *testing.T) {
	rr := httptest.NewRecorder()
	Set(rr, "notice", "Bem-vindo!")

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()

	cookies := res.Cookies()
	var flashCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == CookieName {
			flashCookie = c
			break
		}
	}
	if flashCookie == nil {
		t.Fatal("flash cookie not set")
	}
	if flashCookie.MaxAge != CookieMaxAge {
		t.Errorf("MaxAge = %d, want %d", flashCookie.MaxAge, CookieMaxAge)
	}

	raw, err := base64Decode(flashCookie.Value)
	if err != nil {
		t.Fatalf("decode cookie: %v", err)
	}
	var stored Message
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("unmarshal cookie: %v", err)
	}
	if stored.Kind != "notice" || stored.Message != "Bem-vindo!" {
		t.Errorf("stored = %+v, want notice/Bem-vindo!", stored)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(flashCookie)

	msg, ok := Consume(req)
	if !ok {
		t.Fatal("Consume returned false")
	}
	if msg.Kind != "notice" || msg.Message != "Bem-vindo!" {
		t.Errorf("msg = %+v, want notice/Bem-vindo!", msg)
	}

	rr2 := httptest.NewRecorder()
	Clear(rr2)
	clearRes := rr2.Result()
	defer func() { _ = clearRes.Body.Close() }()

	var cleared *http.Cookie
	for _, c := range clearRes.Cookies() {
		if c.Name == CookieName {
			cleared = c
			break
		}
	}
	if cleared == nil {
		t.Fatal("clear cookie not set")
	}
	if cleared.MaxAge != -1 {
		t.Errorf("clear MaxAge = %d, want -1", cleared.MaxAge)
	}
}
