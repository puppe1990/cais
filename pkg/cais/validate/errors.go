package validate

import "sort"

type FieldErrors map[string]string

func (e *FieldErrors) Add(field, msg string) {
	if *e == nil {
		*e = make(FieldErrors)
	}
	(*e)[field] = msg
}

func (e FieldErrors) Has(field string) bool {
	_, ok := e[field]
	return ok
}

func (e FieldErrors) First() string {
	if len(e) == 0 {
		return ""
	}
	keys := make([]string, 0, len(e))
	for k := range e {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return e[keys[0]]
}

func (e FieldErrors) Any() bool {
	return len(e) > 0
}
