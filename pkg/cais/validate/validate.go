package validate

import (
	"fmt"
	"net/url"
	"strings"
)

// Required reports missing trimmed form fields.
func Required(values map[string]string, keys ...string) error {
	for _, key := range keys {
		if strings.TrimSpace(values[key]) == "" {
			return fmt.Errorf("%s is required", key)
		}
	}
	return nil
}

// URL checks that s is a non-empty http(s) URL.
func URL(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("url is required")
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("url is invalid")
	}
	return nil
}
