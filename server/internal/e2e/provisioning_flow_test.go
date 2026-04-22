package e2e_test

import "testing"

func TestPaidOrderProvisioningFlow(t *testing.T) {
	harness := &ProvisioningHarness{db: openTestDB(t)}
	result, err := harness.RunPaidProvisioningFlow()
	if err != nil {
		t.Fatalf("run provisioning flow: %v", err)
	}
	if result.OrderStatus != "active" {
		t.Fatalf("expected order status active, got %s", result.OrderStatus)
	}
	if result.TaskStatus != "success" {
		t.Fatalf("expected task status success, got %s", result.TaskStatus)
	}
	if result.InstanceStatus != "running" {
		t.Fatalf("expected instance status running, got %s", result.InstanceStatus)
	}
	if result.InstanceNo == "" {
		t.Fatalf("expected instance no to be set")
	}
}
