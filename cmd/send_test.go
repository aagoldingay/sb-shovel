package cmd

import (
	"testing"

	sbmock "github.com/aagoldingay/sb-shovel/mocks"
)

func Test_SendFromFile_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 1}
	err := sendFromFile(m, "testqueue", "../test_files/cmd_send_test.txt")
	if err != nil {
		t.Error(err)
	}

	if m.SourceQueueCount != 5 {
		t.Errorf("Queue had unexpected number of messages: %d", m.SourceQueueCount)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}
