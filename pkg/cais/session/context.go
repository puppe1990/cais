package session

import (
	"context"
	"net/http"
)

type ctxKey struct{}

func WithUserID(r *http.Request, id int64) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, id))
}

func UserID(r *http.Request) (int64, bool) {
	id, ok := r.Context().Value(ctxKey{}).(int64)
	return id, ok
}
