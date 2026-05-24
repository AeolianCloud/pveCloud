package instance

import "testing"

func TestInstanceLifecyclePolicy(t *testing.T) {
	if !CanStart(StatusStopped) || CanStart(StatusRunning) {
		t.Fatal("only stopped instances should be startable")
	}
	if !CanStop(StatusRunning) || CanStop(StatusStopped) {
		t.Fatal("only running instances should be stoppable")
	}
	if CanRelease(StatusReleasing) || CanRelease(StatusReleased) {
		t.Fatal("releasing or released instances must not be releasable again")
	}
	if !CanRelease(StatusError) {
		t.Fatal("error instances should remain releasable for cleanup")
	}
}

func TestTaskPolicyRecognizesWorkerLifecycleTypes(t *testing.T) {
	for _, taskType := range []string{TaskTypeOperationSync, TaskTypeExpiryNotice, TaskTypeExpiryRelease, TaskTypeEmailSend, TaskTypeSMSPlaceholder} {
		if !IsKnownTaskType(taskType) {
			t.Fatalf("task type %q should be known", taskType)
		}
	}
	if IsKnownTaskType("payment_callback") {
		t.Fatal("unopened payment task type must not be accepted")
	}

	for _, status := range []string{TaskStatusPending, TaskStatusRunning, TaskStatusSucceeded, TaskStatusFailed, TaskStatusCancelled} {
		if !IsKnownTaskStatus(status) {
			t.Fatalf("task status %q should be known", status)
		}
	}
	if IsKnownTaskStatus("retrying") {
		t.Fatal("unknown task status must not be accepted")
	}
}

func TestMapVMStatusKeepsOnlyUserVisiblePowerStates(t *testing.T) {
	if got := MapVMStatus(StatusRunning); got != StatusRunning {
		t.Fatalf("running VM should map to running, got %q", got)
	}
	if got := MapVMStatus(StatusStopped); got != StatusStopped {
		t.Fatalf("stopped VM should map to stopped, got %q", got)
	}
	if got := MapVMStatus("paused"); got != StatusError {
		t.Fatalf("unsupported VM status should map to error, got %q", got)
	}
}
