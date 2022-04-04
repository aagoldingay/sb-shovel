package mocks

import "sb-shovel/sbcontroller"

type MockServiceBusController struct {
	sbcontroller.Controller
}

func (m *MockServiceBusController) GetSourceQueueCount() (int, error) {
	return 1, nil
}
func (m *MockServiceBusController) GetTargetQueueCount() (int, error) {
	return 1, nil
}
