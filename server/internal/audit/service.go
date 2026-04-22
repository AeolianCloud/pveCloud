package audit

import "context"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Record(ctx context.Context, event string, businessID uint64) error {
	return nil
}
