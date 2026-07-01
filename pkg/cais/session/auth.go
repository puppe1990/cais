package session

import "net/http"

func SignIn(w http.ResponseWriter, store Store, r *http.Request, userID int64, opts CookieOptions) error {
	if token := TokenFromRequest(r); token != "" {
		store.Delete(token)
	}
	newToken, err := store.Create(userID)
	if err != nil {
		return err
	}
	SetCookie(w, newToken, opts)
	return nil
}

func SignOut(w http.ResponseWriter, store Store, r *http.Request, opts CookieOptions) {
	if token := TokenFromRequest(r); token != "" {
		store.Delete(token)
	}
	ClearCookie(w, opts)
}
