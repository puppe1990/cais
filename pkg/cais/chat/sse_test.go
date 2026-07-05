package chat

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWriteStream_emitsStreamEvent(t *testing.T) {
	rr := httptest.NewRecorder()
	if err := WriteStream(rr, LiveBubble("typing")); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: stream\n") {
		t.Errorf("missing stream event, got %q", body)
	}
	if !strings.Contains(body, "data: ") {
		t.Error("missing data line")
	}
	if !strings.Contains(body, "data-cais-live") {
		t.Error("data should include live bubble")
	}
}

func TestWriteMessage_emitsMessageEvent(t *testing.T) {
	rr := httptest.NewRecorder()
	html := MessageBubble(RoleAssistant, "done", timeFromTest())
	if err := WriteMessage(rr, html); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: message\n") {
		t.Errorf("missing message event, got %q", body)
	}
	if !strings.Contains(body, "done") {
		t.Error("data should include message HTML")
	}
}

func timeFromTest() time.Time {
	return time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
}
