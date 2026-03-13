package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/F3dosik/Hofermart/internal/jwt"
	"github.com/F3dosik/Hofermart/internal/repository"
)

type UserService interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

type userService struct {
	repository repository.Repository
	secretKey  string
}

func NewUserService(repo repository.Repository, secretKey string) UserService {
	return &userService{
		repository: repo,
		secretKey:  secretKey,
	}
}

func (s *userService) Register(ctx context.Context, login, password string) (string, error) {
	if err := validateLogin(login); err != nil {
		return "", err
	}

	if err := validatePassword(password); err != nil {
		return "", err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return "", err
	}

	user, err := s.repository.CreateUserWithBalance(ctx, login, hash)
	if err != nil {
		if errors.Is(err, repository.ErrLoginAlreadyExist) {
			return "", ErrLoginAlreadyExist
		}
		return "", fmt.Errorf("register: %w", err)
	}

	token, err := jwt.GenerateToken(user.ID, s.secretKey)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) Login(ctx context.Context, login, password string) (string, error) {
	if err := validateLogin(login); err != nil {
		return "", err
	}

	user, err := s.repository.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("login: %w", err)
	}

	if !checkPassword(password, user.Password) {
		return "", ErrInvalidCredentials
	}

	token, err := jwt.GenerateToken(user.ID, s.secretKey)
	if err != nil {
		return "", err
	}

	return token, nil
}
