package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/F3dosik/Hofermart/internal/model"
	"github.com/go-resty/resty/v2"
)

type AccrualClient interface {
	GetAccrual(ctx context.Context, orderNumber string) (*model.AccrualResponse, error)
}

type accrualClient struct {
	url    string
	client *resty.Client
}

func NewAccrual(url string) AccrualClient {
	return &accrualClient{
		url:    url,
		client: resty.New(),
	}
}

func (c *accrualClient) GetAccrual(ctx context.Context, orderNumber string) (*model.AccrualResponse, error) {
	fullURL, err := url.JoinPath(c.url, "api/orders", orderNumber)
	if err != nil {
		return nil, fmt.Errorf("get accrual:%w: %w", ErrBuildURL, err)
	}
	var accrualResponse model.AccrualResponse
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&accrualResponse).
		Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("get accrual: %w: %w", ErrRequestExec, err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
	case http.StatusNoContent:
		return nil, ErrOrderNotFound
	case http.StatusTooManyRequests:
		retryAfter, err := parseRetryAfter(resp)
		if err != nil {
			return nil, fmt.Errorf("get accrual: parse Retry-After: %w", err)
		}
		return nil, &ErrRateLimit{RetryAfter: retryAfter}
	default:
		return nil, fmt.Errorf("get accrual: unexpected status code: %d", resp.StatusCode())
	}

	return &accrualResponse, nil

}

func parseRetryAfter(resp *resty.Response) (time.Duration, error) {
	value := resp.Header().Get("Retry-After")
	if value == "" {
		return 0, fmt.Errorf("Retry-After header not present")
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	if t, err := time.Parse(http.TimeFormat, value); err == nil {
		return time.Until(t), nil
	}
	return 0, fmt.Errorf("invalid Retry-After format: %s", value)
}
