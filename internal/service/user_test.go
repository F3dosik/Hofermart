package service

import (
	"context"
	"errors"
	"testing"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testUser = &model.User{
	ID:    uuid.New(),
	Login: "alice",
}

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		login     string
		password  string
		setupMock func(*mockRepository)
		wantToken bool
		wantErr   error
	}{
		{
			name:     "success",
			login:    "alice",
			password: "secret123",
			setupMock: func(m *mockRepository) {
				m.On("CreateUserWithBalance", ctx, "alice", mock.AnythingOfType("string")).
					Return(testUser, nil)
			},
			wantToken: true,
		},
		{
			name:      "empty login",
			login:     "",
			password:  "secret123",
			setupMock: func(m *mockRepository) {},
			wantErr:   ErrEmptyLogin,
		},
		{
			name:      "password too short",
			login:     "alice",
			password:  "short",
			setupMock: func(m *mockRepository) {},
			wantErr:   ErrPasswordTooShort,
		},
		{
			name:     "login already exists",
			login:    "alice",
			password: "secret123",
			setupMock: func(m *mockRepository) {
				m.On("CreateUserWithBalance", ctx, "alice", mock.AnythingOfType("string")).
					Return(nil, repository.ErrLoginAlreadyExist)
			},
			wantErr: ErrLoginAlreadyExist,
		},
		{
			name:     "repository unexpected error",
			login:    "alice",
			password: "secret123",
			setupMock: func(m *mockRepository) {
				m.On("CreateUserWithBalance", ctx, "alice", mock.AnythingOfType("string")).
					Return(nil, errors.New("db is down"))
			},
			wantErr: errors.New("any"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			tt.setupMock(repo)

			svc := newTestUserService(repo)
			token, err := svc.Register(ctx, tt.login, tt.password)

			if tt.wantToken {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			} else {
				assert.Error(t, err)
				if isSentinel(tt.wantErr) {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				assert.Empty(t, token)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()
	validHash, err := hashPassword("secret123")
	if err != nil {
		panic(err)
	}

	userWithHash := &model.User{
		ID:       uuid.New(),
		Login:    "alice",
		Password: validHash,
	}

	tests := []struct {
		name      string
		login     string
		password  string
		setupMock func(*mockRepository)
		wantToken bool
		wantErr   error
	}{
		{
			name:     "success",
			login:    "alice",
			password: "secret123",
			setupMock: func(m *mockRepository) {
				m.On("GetUserByLogin", ctx, "alice").
					Return(userWithHash, nil)
			},
			wantToken: true,
		},
		{
			name:      "empty login",
			login:     "",
			password:  "secret123",
			setupMock: func(m *mockRepository) {},
			wantErr:   ErrEmptyLogin,
		},
		{
			name:     "user not found",
			login:    "ghost",
			password: "secret123",
			setupMock: func(m *mockRepository) {
				m.On("GetUserByLogin", ctx, "ghost").
					Return(nil, repository.ErrUserNotFound)
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			login:    "alice",
			password: "wrongpass",
			setupMock: func(m *mockRepository) {
				m.On("GetUserByLogin", ctx, "alice").
					Return(userWithHash, nil)
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:     "repository unexpected error",
			login:    "alice",
			password: "secret123",
			setupMock: func(m *mockRepository) {
				m.On("GetUserByLogin", ctx, "alice").
					Return(nil, errors.New("db is down"))
			},
			wantErr: errors.New("any"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			tt.setupMock(repo)

			svc := newTestUserService(repo)
			token, err := svc.Login(ctx, tt.login, tt.password)

			if tt.wantToken {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			} else {
				assert.Error(t, err)
				if isSentinel(tt.wantErr) {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				assert.Empty(t, token)
			}

			repo.AssertExpectations(t)
		})
	}
}

func isSentinel(err error) bool {
	return errors.Is(err, ErrEmptyLogin) ||
		errors.Is(err, ErrPasswordTooShort) ||
		errors.Is(err, ErrLoginAlreadyExist) ||
		errors.Is(err, ErrInvalidCredentials)
}
