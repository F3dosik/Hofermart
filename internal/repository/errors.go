package repository

import "errors"

var (
	ErrLoginAlreadyExist               = errors.New("login alreadyexist")
	ErrUserNotFound                    = errors.New("user not found")
	ErrOrderNotFound                   = errors.New("order not found")
	ErrOrderAlreadyExist               = errors.New("order already exist")
	ErrOrderAlreadyExistForAnotherUser = errors.New("order already exist for another user")
	ErrNotEnoughBalance                = errors.New("not enough balance")
)
