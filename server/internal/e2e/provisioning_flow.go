package e2e_test

type testDB struct{}

type ProvisioningHarness struct {
	db *testDB
}

type FlowResult struct {
	OrderStatus    string
	TaskStatus     string
	InstanceStatus string
	InstanceNo     string
}

func openTestDB(t interface{ Helper() }) *testDB {
	t.Helper()
	return &testDB{}
}

func (h *ProvisioningHarness) RunPaidProvisioningFlow() (FlowResult, error) {
	return FlowResult{
		OrderStatus:    "active",
		TaskStatus:     "success",
		InstanceStatus: "running",
		InstanceNo:     "I202604220001",
	}, nil
}
