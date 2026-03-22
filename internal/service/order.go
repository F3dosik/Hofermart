package service

import (
	"context"
	"errors"
	"time"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/F3dosik/Hofermart/internal/repository"
	"github.com/F3dosik/Hofermart/internal/worker"
	"github.com/google/uuid"
)

type OrderService interface {
	UploadOrder(ctx context.Context, number string, userID uuid.UUID) error
	GetOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error)
}

type orderService struct {
	repository repository.Repository
	jobChan    chan *worker.ScheduleJob
}

func NewOrderService(repo repository.Repository, jobChan chan *worker.ScheduleJob) OrderService {
	return &orderService{repository: repo, jobChan: jobChan}
}

func (s *orderService) UploadOrder(ctx context.Context, number string, userID uuid.UUID) error {
	if err := validateOrder(number); err != nil {
		return err
	}

	if err := s.repository.UploadOrder(ctx, number, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrOrderAlreadyExistForAnotherUser):
			return ErrOrderAlreadyExistForAnotherUser
		case errors.Is(err, repository.ErrOrderAlreadyExist):
			return ErrOrderAlreadyExist
		default:
			return err
		}
	}

	s.jobChan <- &worker.ScheduleJob{Number: number, RunAt: time.Now(), Attempt: 0}
	return nil
}

func (s *orderService) GetOrders(ctx context.Context, userID uuid.UUID) ([]*model.Order, error) {
	orders, err := s.repository.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
