package validate

import (
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
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

// Email checks that s is a valid email address.
func Email(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return fmt.Errorf("email is invalid")
	}
	return nil
}

// Int checks that s is a base-10 integer.
func Int(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("value is required")
	}
	if _, err := strconv.ParseInt(s, 10, 64); err != nil {
		return fmt.Errorf("value must be an integer")
	}
	return nil
}

// Date checks that s is a YYYY-MM-DD date.
func Date(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("date is required")
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return fmt.Errorf("date is invalid")
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
