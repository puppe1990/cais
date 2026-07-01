package validate

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
	for _, msg := range e {
		return msg
	}
	return ""
}

func (e FieldErrors) Any() bool {
	return len(e) > 0
}
