package main

import (
	"os"
	"testing"

	cc "github.com/aagoldingay/sb-shovel/config"
	sbio "github.com/aagoldingay/sb-shovel/io"
	sbmock "github.com/aagoldingay/sb-shovel/mocks"
)

func Test_Config_Update_New(t *testing.T) {
	cfg, err := cc.NewConfigController()
	if err != nil {
		t.Error(err)
	}

	config(cfg, []string{"update", "TEST_CONFIG_UPDATE_NEW", "TEST_VALUE"})

	v, err := cfg.GetConfigValue("TEST_CONFIG_UPDATE_NEW")

	if err != nil {
		t.Error(err)
	}

	if v != "TEST_VALUE" {
		t.Errorf("value for TEST_CONFIG_UPDATE_NEW was not found")
	}
}

func Test_Config_Update_Existing(t *testing.T) {
	cfg, err := cc.NewConfigController()
	if err != nil {
		t.Error(err)
	}

	cfg.UpdateConfig("TEST_CONFIG_UPDATE_EXISTING", "old_value")
	cfg.SaveConfig()

	config(cfg, []string{"update", "TEST_CONFIG_UPDATE_EXISTING", "new_value"})

	v, err := cfg.GetConfigValue("TEST_CONFIG_UPDATE_EXISTING")

	if err != nil {
		t.Error(err)
	}

	if v != "new_value" {
		t.Errorf("value for TEST_CONFIG_UPDATE_EXISTING was not as expected")
	}
}

func Test_Config_Remove(t *testing.T) {
	cfg, err := cc.NewConfigController()
	if err != nil {
		t.Error(err)
	}

	cfg.UpdateConfig("TEST_CONFIG_REMOVE", "TEST_VALUE")
	cfg.SaveConfig()

	config(cfg, []string{"remove", "TEST_CONFIG_REMOVE"})

	v, err := cfg.GetConfigValue("TEST_CONFIG_REMOVE")

	if err.Error() != cc.ERR_CONFIGEMPTY {
		t.Error(err)
	}

	if v != "" {
		t.Errorf("config returned a value: %s", v)
	}
}

func Test_Config_List(t *testing.T) {
	cfg, err := cc.NewConfigController()
	if err != nil {
		t.Error(err)
	}

	if v := cfg.ListConfig(); v != cc.ERR_CONFIGEMPTY {
		t.Errorf("config wasn't empty: %s", v)
	}
}

func Test_Pull_Fail_EmptyQueue(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 0}

	err := pull(m, "testqueue", false, 5)
	if err == nil {
		t.Error(err)
	}

	if m.SourceQueueClosed != true {
		t.Error("Queue not closed")
	}
}

func Test_Pull_Success_OneFile(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 5}

	err := pull(m, "testqueue", true, 5)
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

func Test_Pull_Success_TwoFiles(t *testing.T) {
	m := &sbmock.MockServiceBusController{SourceQueueCount: 10}
	err := pull(m, "testqueue", false, 5)
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
