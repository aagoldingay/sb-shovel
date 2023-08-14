package cmd

import (
	"testing"

	sbmock "github.com/aagoldingay/sb-shovel/mocks"
)

func Test_Delete_One_Fail_NoMessages(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 0}
	err := delete(m, "testqueue", false, false, false)
	if err.Error() != "no messages to delete" {
		t.Error(err)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}

func Test_Delete_One_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 1}
	err := delete(m, "testqueue", false, false, false)
	if err != nil {
		t.Error(err)
	}

	if m.SourceQueueCount != 0 {
		t.Errorf("Queue had unexpected number of messages: %d", m.SourceQueueCount)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}

func Test_Delete_All_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 10}
	err := delete(m, "testqueue", false, true, false)
	if err != nil {
		t.Error(err)
	}

	if m.SourceQueueCount != 0 {
		t.Errorf("Queue had unexpected number of messages: %d", m.SourceQueueCount)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}
