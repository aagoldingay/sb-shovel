package cmd

import (
	"os"
	"testing"

	sbio "github.com/aagoldingay/sb-shovel/io"
	sbmock "github.com/aagoldingay/sb-shovel/mocks"
)

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
