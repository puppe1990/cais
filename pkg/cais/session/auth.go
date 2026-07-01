package session

import "net/http"

func SignIn(w http.ResponseWriter, store Store, userID int64, opts CookieOptions) error {
	token, err := store.Create(userID)
	if err != nil {
		return err
	}
	SetCookie(w, token, opts)
	return nil
}

func SignOut(w http.ResponseWriter, store Store, r *http.Request) {
	if token := TokenFromRequest(r); token != "" {
		store.Delete(token)
	}
	ClearCookie(w)
}
