package audit

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, event string, businessID uint64) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.Record(ctx, event, "order", businessID, nil)
}
