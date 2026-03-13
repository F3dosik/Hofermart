package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/F3dosik/Hofermart/internal/service"
)

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func parseRequest(r *http.Request) (*authRequest, error) {
	var user authRequest

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	user, err := parseRequest(r)
	if err != nil {
		h.logger.Debugw("can't decode JSON body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := h.userService.Register(r.Context(), user.Login, user.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoginAlreadyExist):
			h.logger.Debugw("register: login already taken", "login", user.Login)
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		case errors.Is(err, service.ErrEmptyLogin),
			errors.Is(err, service.ErrPasswordTooShort):
			h.logger.Debugw("register", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		default:
			h.logger.Errorw("register: internal error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	h.setTokenCookie(w, token)

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	user, err := parseRequest(r)
	if err != nil {
		h.logger.Debugw("can't decode JSON body", "error", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), user.Login, user.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyLogin):
			h.logger.Debugw("login", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidCredentials):
			h.logger.Debugw("login", "error", err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		default:
			h.logger.Errorw("login: internal error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	h.setTokenCookie(w, token)

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})
}
