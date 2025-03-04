package web

import (
	"net/http"
)

// AuthLoginHandler renders the login form
func AuthLoginHandler(w http.ResponseWriter, r *http.Request) {
	component := LoginForm()
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// AuthRegisterHandler renders the register form
func AuthRegisterHandler(w http.ResponseWriter, r *http.Request) {
	component := RegisterForm()
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// AuthRecoverHandler renders the recover form
func AuthRecoverHandler(w http.ResponseWriter, r *http.Request) {
	component := RecoverForm()
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
