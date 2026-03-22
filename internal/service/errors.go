package service

import (
	"errors"
)

var (
	ErrEmptyLogin                      = errors.New("login is empty")
	ErrPasswordTooShort                = errors.New("password too short")
	ErrLoginAlreadyExist               = errors.New("login already exist")
	ErrInvalidCredentials              = errors.New("invalid credentials")
	ErrInvalidOrderNumber              = errors.New("invalid order number")
	ErrOrderAlreadyExist               = errors.New("order already exist")
	ErrOrderAlreadyExistForAnotherUser = errors.New("order already exist for another user")
	ErrBalanceNotFound                 = errors.New("balance not found")
	ErrNotEnoughBalance                = errors.New("not enough balance")
)
