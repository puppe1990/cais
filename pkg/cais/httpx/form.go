package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ParseFormOrJSON fills r.Form (and r.PostForm) from urlencoded/multipart forms
// or a JSON object body so FormValue works for both classic HTML forms and
// Inertia useForm posts (application/json).
func ParseFormOrJSON(r *http.Request) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	ct := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil && ct != "" {
		// Fall through to ParseForm for malformed types.
		mediaType = ""
	}
	if mediaType == "application/json" {
		return parseJSONForm(r)
	}
	return r.ParseForm()
}

func parseJSONForm(r *http.Request) error {
	// Preserve query params already (or about to be) in Form.
	if err := r.ParseForm(); err != nil {
		return err
	}
	if r.Body == nil {
		return nil
	}
	defer r.Body.Close()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MiB
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}

	if r.PostForm == nil {
		r.PostForm = make(url.Values, len(data))
	}
	if r.Form == nil {
		r.Form = make(url.Values, len(data))
	}
	for k, v := range data {
		s := formValueString(v)
		r.PostForm.Set(k, s)
		r.Form.Set(k, s)
	}
	return nil
}

func formValueString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case json.Number:
		return x.String()
	default:
		return fmt.Sprint(x)
	}
}

// FormTruthy reports whether a form/JSON field value is true.
// HTML checkboxes send "on"; Inertia JSON posts send "true".
func FormTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "on", "yes":
		return true
	default:
		return false
	}
}
