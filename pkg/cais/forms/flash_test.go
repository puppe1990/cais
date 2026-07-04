package forms

import (
	"testing"

	"github.com/puppe1990/cais/pkg/cais/flash"
)

func TestFlashMessage_fromStruct(t *testing.T) {
	if got := FlashMessage(flash.Message{Kind: "notice", Message: "Bem-vindo!"}); got != "Bem-vindo!" {
		t.Errorf("FlashMessage() = %q, want Bem-vindo!", got)
	}
}

func TestFlashMessage_fromPointer(t *testing.T) {
	msg := &flash.Message{Kind: "notice", Message: "Saved!"}
	if got := FlashMessage(msg); got != "Saved!" {
		t.Errorf("FlashMessage() = %q, want Saved!", got)
	}
}

func TestFlashMessage_nilPointer(t *testing.T) {
	if got := FlashMessage((*flash.Message)(nil)); got != "" {
		t.Errorf("FlashMessage(nil) = %q, want empty", got)
	}
}

func TestFlashMessage_nilInterface(t *testing.T) {
	if got := FlashMessage(nil); got != "" {
		t.Errorf("FlashMessage(nil) = %q, want empty", got)
	}
}

func TestFuncs_registersFlashMessage(t *testing.T) {
	if _, ok := Funcs()["flashMessage"]; !ok {
		t.Error("flashMessage missing from Funcs()")
	}
}
