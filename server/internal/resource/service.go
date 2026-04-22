package resource

import "context"

type Service struct {
	client VMClient
}

func NewService(client VMClient) *Service {
	return &Service{client: client}
}

func (s *Service) CreateVM(ctx context.Context, req CreateVMRequest) (CreateVMResponse, error) {
	return s.client.CreateVM(ctx, req)
}

func (s *Service) StartVM(ctx context.Context, instanceRef string) error {
	return s.client.StartVM(ctx, instanceRef)
}

func (s *Service) StopVM(ctx context.Context, instanceRef string) error {
	return s.client.StopVM(ctx, instanceRef)
}

func (s *Service) RebootVM(ctx context.Context, instanceRef string) error {
	return s.client.RebootVM(ctx, instanceRef)
}

func (s *Service) ReinstallVM(ctx context.Context, req ReinstallVMRequest) error {
	return s.client.ReinstallVM(ctx, req)
}
