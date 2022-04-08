package mocks

import (
	"testing"

	"github.com/aagoldingay/sb-shovel/sbcontroller"
)

func CalculateQueueSize(c sbcontroller.Controller) int {
	s, err := c.GetSourceQueueCount()
	if err != nil {
		return 0
	}
	t, err := c.GetTargetQueueCount()
	if err != nil {
		return 0
	}
	return s + t
}

func Test_MockController_CalculateQueueSize(t *testing.T) {
	m := &MockServiceBusController{SourceQueueCount: 1, TargetQueueCount: 1}

	c := CalculateQueueSize(m)
	if c != 2 {
		t.Fail()
	}
}
