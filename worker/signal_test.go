package worker

import (
	"testing"
)

func TestContinueSignalOperation(t *testing.T) {
	var sig Signal
	sig |= ContinueWithRetrySignal

	if sig&ContinueWithRetrySignal == 0 {
		t.Errorf("wrong signal operation")
	}
}
