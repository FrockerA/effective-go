package service

import (
	"context"
	"errors"
	"time"

	"effective-go/internal/model"

	"github.com/google/uuid"
)

var (
	ErrInvalidDateOrder     = errors.New("start_date must be before end_date")
	ErrInvalidDateFormat    = errors.New("invalid date format, expected MM-YYYY")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	List(ctx context.Context, userID *uuid.UUID, serviceName string) ([]model.Subscription, error)
	CalculateTotal(ctx context.Context, userID *uuid.UUID, serviceName string, periodStart, periodEnd time.Time) (int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
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
		ID:          uuid.New(),
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

func (s *SubscriptionService) CalculateTotal(ctx context.Context, userID *uuid.UUID, serviceName, startDateStr, endDateStr string) (int, error) {
	periodStart, err := ParseDate(startDateStr)
	if err != nil {
		return 0, ErrInvalidDateFormat
	}

	periodEnd, err := ParseDate(endDateStr)
	if err != nil {
		return 0, ErrInvalidDateFormat
	}

	if periodStart.After(periodEnd) {
		return 0, ErrInvalidDateOrder
	}

	return s.repo.CalculateTotal(ctx, userID, serviceName, periodStart, periodEnd)
}

type UpdateDTO struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, dto UpdateDTO) (*model.Subscription, error) {
	sub, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

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

	sub.ServiceName = dto.ServiceName
	sub.Price = dto.Price
	sub.StartDate = startDate
	sub.EndDate = endDate

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}
