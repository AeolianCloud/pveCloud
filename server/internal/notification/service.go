package notification

import "context"

type Repo interface {
	Create(ctx context.Context, n Notification) (Notification, error)
	ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error)
	MarkRead(ctx context.Context, id uint64, userID uint64) error
	CountUnread(ctx context.Context, userID uint64) (int, error)
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) SendProvisionSuccess(ctx context.Context, userID uint64, instanceNo string) error {
	_, err := s.repo.Create(ctx, Notification{
		UserID: userID,
		Title:  "实例开通成功",
		Body:   "您的实例 " + instanceNo + " 已成功开通。",
		Type:   "provision",
	})
	return err
}

func (s *Service) SendProvisionFailure(ctx context.Context, userID uint64, orderID uint64) error {
	_, err := s.repo.Create(ctx, Notification{
		UserID: userID,
		Title:  "实例开通失败",
		Body:   "您的订单关联实例开通失败，请稍后重试或联系客服。",
		Type:   "provision",
	})
	return err
}

func (s *Service) ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error) {
	return s.repo.ListByUser(ctx, userID, limit)
}

func (s *Service) MarkRead(ctx context.Context, id uint64, userID uint64) error {
	return s.repo.MarkRead(ctx, id, userID)
}

func (s *Service) CountUnread(ctx context.Context, userID uint64) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}
