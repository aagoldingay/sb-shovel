package io

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

func Test_ReadFile_Success(t *testing.T) {
	c := ReadFile("../test_files/filewriter_test.json")

	if c == nil {
		t.Error("File read failed")
	}

	// check contents
	// line 2 = 25k characters
	if len(c) > 2 {
		t.Error("Overflow found in file")
	}
}

func Test_ReadFile_Fail_BadPath(t *testing.T) {
	c := ReadFile("../test_files/no_file.json")

	if c != nil {
		t.Error("Test returned unexpected content")
	}
}

func Test_CreateDir_Create(t *testing.T) {
	// setup
	if _, err := os.Stat(dirName); !os.IsNotExist(err) {
		err := helper_deleteDir(t)

		if err != nil {
			t.Errorf("Test setup failed: %s", err.Error())
		}
	}

	// test
	err := CreateDir()

	if err != nil {
		t.Error("Directory creation failed")
	}

	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		t.Error("Directory does not exist")
	}

	// teardown
	err = helper_deleteDir(t)
	if err != nil {
		t.Errorf("Test teardown failed: %s", err.Error())
	}
}

func Test_CreateDir_Exists(t *testing.T) {
	// setup
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := CreateDir()
		if err != nil {
			t.Errorf("Test setup failed: %s", err.Error())
		}
	}

	// test
	err := CreateDir()

	if err != nil {
		t.Error("Directory creation failed")
	}

	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		t.Error("Directory does not exist")
	}

	// teardown
	err = helper_deleteDir(t)
	if err != nil {
		t.Errorf("Test teardown failed: %s", err.Error())
	}
}

func Test_WriteFile_OneFile_Success(t *testing.T) {
	// setup
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := CreateDir()
		if err != nil {
			t.Errorf("Test setup failed: %s", err.Error())
		}
	}

	eChan := make(chan error, 2)
	var wg sync.WaitGroup
	s := []string{"test1", "test2", "test3"}

	// test
	wg.Add(1)
	go WriteFile(eChan, 1, s, &wg)
	wg.Wait()

	if len(eChan) > 0 {
		done := false
		for !done {
			if e, ok := <-eChan; ok {
				if e.Error() == "complete" {
					done = true
					continue
				}
				t.Errorf("Error while writing file: %s", e)
			}
		}
	}
	close(eChan)

	c := ReadFile(fmt.Sprintf("%s/sb_output_000001.txt", dirName))

	if c == nil {
		t.Error("File read failed")
	}

	if len(c) != 3 {
		t.Error("Unexpected number of lines per file")
	}

	// teardown
	err := helper_deleteDir(t)
	if err != nil {
		t.Errorf("Test teardown failed: %s", err.Error())
	}
}

func helper_deleteDir(t *testing.T) error {
	t.Helper()
	err := os.RemoveAll(dirName)
	if err != nil {
		return err
	}
	return nil
}
