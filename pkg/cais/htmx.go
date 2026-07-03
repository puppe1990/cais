package cais

import (
	"encoding/json"
	"net/http"
)

func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// SetTrigger sets HX-Trigger for client-side events after swap.
func SetTrigger(w http.ResponseWriter, id string) {
	w.Header().Set("HX-Trigger", id)
}

// SetToast sets HX-Trigger so cais.js shows a transient toast after the HTMX swap.
func SetToast(w http.ResponseWriter, message string) {
	payload, err := json.Marshal(map[string]string{"caisToast": message})
	if err != nil {
		SetTrigger(w, "caisToast")
		return
	}
	w.Header().Set("HX-Trigger", string(payload))
}

// SetRetarget sets HX-Retarget to change the swap target from the response.
func SetRetarget(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Retarget", selector)
}

// SetFocus sets HX-Trigger so cais.js focuses a field after swap (e.g. first invalid input).
func SetFocus(w http.ResponseWriter, selector string) {
	payload, err := json.Marshal(map[string]string{"caisFocus": selector})
	if err != nil {
		return
	}
	w.Header().Set("HX-Trigger", string(payload))
}
