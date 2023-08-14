package cmd

import (
	"testing"

	sbmock "github.com/aagoldingay/sb-shovel/mocks"
)

func Test_Requeue_One_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 1, TargetQueueCount: 0}
	err := requeue(m, "testqueue", false, true)
	if err != nil {
		t.Error(err)
	}

	if m.SourceQueueCount != 0 && m.TargetQueueCount != 1 {
		t.Errorf("A queue had unexpected number of messages - source: %d , target: %d",
			m.SourceQueueCount, m.TargetQueueCount)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}

	if m.TargetQueueClosed != true {
		t.Error("Queue not closed")
	}
}

func Test_Requeue_One_Fail_TargetDlq(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 1, TargetQueueCount: 0}
	err := requeue(m, "testqueue", false, false)
	if err.Error() != "cannot requeue messages directly to a dead letter queue" {
		t.Error(err)
	}
}

func Test_Requeue_All_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 10, TargetQueueCount: 0}
	err := requeue(m, "testqueue", true, true)
	if err != nil {
		t.Error(err)
	}

	if m.SourceQueueCount != 0 && m.TargetQueueCount != 10 {
		t.Errorf("A queue had unexpected number of messages - source: %d , target: %d",
			m.SourceQueueCount, m.TargetQueueCount)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}

	if m.TargetQueueClosed != true {
		t.Error("Queue not closed")
	}
}

func Test_Requeue_All_Fail_TargetDlq(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 10, TargetQueueCount: 0}
	err := requeue(m, "testqueue", true, false)
	if err.Error() != "cannot requeue messages directly to a dead letter queue" {
		t.Error(err)
	}
}
