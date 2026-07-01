package cais

import "net/http"

func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// SetTrigger sets HX-Trigger for client-side events after swap.
func SetTrigger(w http.ResponseWriter, id string) {
	w.Header().Set("HX-Trigger", id)
}

// SetRetarget sets HX-Retarget to change the swap target from the response.
func SetRetarget(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Retarget", selector)
}
