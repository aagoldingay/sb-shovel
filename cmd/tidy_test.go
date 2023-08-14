package cmd

import (
	"testing"

	sbmock "github.com/aagoldingay/sb-shovel/mocks"
)

func Test_Tidy_Invalid_Regex(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 5}

	err := tidy(m, "testqueue", "(?<", false, false)
	if err.Error() != "error parsing regexp: invalid or unsupported Perl syntax: `(?<`" {
		t.Error(err)
	}
}

func Test_Tidy_No_Execute(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 5}

	execute := false
	err := tidy(m, "testqueue", "ab+c", false, execute)

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

func Test_Tidy_Execute_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 5}

	execute := true
	err := tidy(m, "testqueue", "ab+c", false, execute)

	if err != nil {
		t.Error(err)
	}

	if m.SourceQueueCount != 3 {
		t.Errorf("Queue had unexpected number of messages: %d", m.SourceQueueCount)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}
