package mocks

import (
	"errors"
	"fmt"
	sbc "sb-shovel/sbcontroller"
)

type MockServiceBusController struct {
	sbc.Controller

	SourceQueueCount, TargetQueueCount   int
	SourceQueueClosed, TargetQueueClosed bool
}

func (m *MockServiceBusController) DeleteOneMessage(requeue bool) error {
	m.SourceQueueCount = m.SourceQueueCount - 1
	if requeue {
		m.TargetQueueCount++
	}
	return nil
}

func (m *MockServiceBusController) DeleteManyMessages(errChan chan error, requeue bool, total int) {
	if requeue {
		m.TargetQueueCount = m.SourceQueueCount
	}
	m.SourceQueueCount = 0
	errChan <- fmt.Errorf("context canceled")
}

func (m *MockServiceBusController) DisconnectSource() error {
	m.SourceQueueClosed = true
	return nil
}
func (m *MockServiceBusController) DisconnectTarget() error {
	m.TargetQueueClosed = true
	return nil
}

func (m *MockServiceBusController) ReadSourceQueue(outChan chan []string, errChan chan error, maxWrite int) {
	msgs := []string{}
	for i := 0; i < m.SourceQueueCount/5; i++ {
		for j := 0; j < maxWrite; j++ {
			msgs = append(msgs, "hello, world")
		}
		outChan <- msgs
		msgs = []string{}
	}
	errChan <- errors.New(sbc.ERR_QUEUEEMPTY)
}

func (m *MockServiceBusController) SendJsonMessage(q bool, data []byte) error {
	m.SourceQueueCount++
	return nil
}

func (m *MockServiceBusController) SendManyJsonMessages(q bool, data [][]byte) error {
	m.SourceQueueCount += len(data)
	return nil
}

func (m *MockServiceBusController) SetupSourceQueue(name string, dlq, purge bool) error {
	return nil
}

func (m *MockServiceBusController) SetupTargetQueue(name string, dlq, purge bool) error {
	return nil
}

func (m *MockServiceBusController) GetSourceQueueCount() (int, error) {
	return m.SourceQueueCount, nil
}
func (m *MockServiceBusController) GetTargetQueueCount() (int, error) {
	return m.TargetQueueCount, nil
}
