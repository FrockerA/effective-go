package service

import (
	"context"
	"errors"
	"time"

	"effective-go/internal/model"

	"github.com/google/uuid"
)

var (
	ErrInvalidDateOrder  = errors.New("start_date must be before end_date")
	ErrInvalidDateFormat = errors.New("invalid date format, expected MM-YYYY")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	List(ctx context.Context, userID *uuid.UUID, serviceName string) ([]model.Subscription, error)
	CalculateTotal(ctx context.Context, userID *uuid.UUID, serviceName string) (int, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("01-2006", dateStr)
}

type CreateDTO struct {
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
}

func (s *SubscriptionService) Create(ctx context.Context, dto CreateDTO) (*model.Subscription, error) {
	startDate, err := ParseDate(dto.StartDate)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	var endDate *time.Time
	if dto.EndDate != nil {
		parsedEndDate, err := ParseDate(*dto.EndDate)
		if err != nil {
			return nil, ErrInvalidDateFormat
		}
		if startDate.After(parsedEndDate) {
			return nil, ErrInvalidDateOrder
		}
		endDate = &parsedEndDate
	}

	sub := &model.Subscription{
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		UserID:      dto.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, userID *uuid.UUID, serviceName string) ([]model.Subscription, error) {
	return s.repo.List(ctx, userID, serviceName)
}

func (s *SubscriptionService) CalculateTotal(ctx context.Context, userID *uuid.UUID, serviceName string) (int, error) {
	return s.repo.CalculateTotal(ctx, userID, serviceName)
}
