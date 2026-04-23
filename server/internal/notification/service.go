package notification

import "context"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SendProvisionSuccess(ctx context.Context, userID uint64, instanceNo string) error {
	return nil
}

func (s *Service) SendProvisionFailure(ctx context.Context, userID uint64, orderID uint64) error {
	return nil
}
