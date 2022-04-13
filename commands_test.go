package main

import (
	"os"
	"testing"

	sbio "github.com/aagoldingay/sb-shovel/io"
	sbmock "github.com/aagoldingay/sb-shovel/mocks"
)

func Test_Dump_Fail_EmptyQueue(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 0}

	err := dump(m, "testqueue", false, 5)
	if err == nil {
		t.Error(err)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}

func Test_Dump_Success_OneFile(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 5}

	err := dump(m, "testqueue", true, 5)
	if err != nil {
		t.Error(err)
	}

	// check files
	if _, err := os.Stat("sb-shovel-output/sb_output_000001.txt"); os.IsNotExist(err) {
		t.Errorf("file 1 does not exist")
	}

	c := sbio.ReadFile("sb-shovel-output/sb_output_000001.txt")

	if len(c) != 5 {
		t.Errorf("Unexpected lines in file: %d", len(c))
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}

	// cleanup
	err = os.RemoveAll("sb-shovel-output")
	if err != nil {
		t.Error(err)
	}
}

func Test_Dump_Success_TwoFiles(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 10}
	err := dump(m, "testqueue", false, 5)
	if err != nil {
		t.Error(err)
	}

	// check files
	if _, err := os.Stat("sb-shovel-output/sb_output_000001.txt"); os.IsNotExist(err) {
		t.Errorf("file 1 does not exist")
	}

	c := sbio.ReadFile("sb-shovel-output/sb_output_000001.txt")

	if len(c) != 5 {
		t.Errorf("Unexpected lines in file 1: %d", len(c))
	}

	c = sbio.ReadFile("sb-shovel-output/sb_output_000002.txt")

	if len(c) != 5 {
		t.Errorf("Unexpected lines in file 2: %d", len(c))
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}

	// cleanup
	err = os.RemoveAll("sb-shovel-output")
	if err != nil {
		t.Error(err)
	}
}

func Test_Empty_One_Fail_NoMessages(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 0}
	err := empty(m, "testqueue", false, false, false, false)
	if err.Error() != "no messages to delete" {
		t.Error(err)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}

func Test_Empty_One_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 1}
	err := empty(m, "testqueue", false, false, false, false)
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

func Test_Empty_One_Requeue_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 1, TargetQueueCount: 0}
	err := empty(m, "testqueue", true, false, true, false)
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

func Test_Empty_One_Requeue_Fail_TargetDlq(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 1, TargetQueueCount: 0}
	err := empty(m, "testqueue", false, false, true, false)
	if err.Error() != "cannot requeue messages directly to a dead letter queue" {
		t.Error(err)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}

func Test_Empty_All_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 10}
	err := empty(m, "testqueue", false, true, false, false)
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

func Test_Empty_All_Requeue_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 10, TargetQueueCount: 0}
	err := empty(m, "testqueue", true, true, true, false)
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

func Test_Empty_All_Requeue_Fail_TargetDlq(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 10, TargetQueueCount: 0}
	err := empty(m, "testqueue", false, true, true, false)
	if err.Error() != "cannot requeue messages directly to a dead letter queue" {
		t.Error(err)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}

func Test_SendFromFile_Success(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 1}
	err := sendFromFile(m, "testqueue", "test_files/cmd_send_test.txt")
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
