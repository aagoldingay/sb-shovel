package main

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

func Test_readFile_Success(t *testing.T) {
	c := readFile("test_files/filewriter_test.json")

	if c == nil {
		t.Error("File read failed")
	}

	// check contents
	// line 2 = 25k characters
	if len(c) > 2 {
		t.Error("Overflow found in file")
	}
}

func Test_readFile_BadPath(t *testing.T) {
	c := readFile("test_files/no_file.json")

	if c != nil {
		t.Error("Test returned unexpected content")
	}
}

func Test_createDir_Create(t *testing.T) {
	// setup
	if _, err := os.Stat(dirName); !os.IsNotExist(err) {
		err := helper_deleteDir(t)

		if err != nil {
			t.Errorf("Test setup failed: %s", err.Error())
		}
	}

	// test
	err := createDir()

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

func Test_createDir_Exists(t *testing.T) {
	// setup
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := createDir()
		if err != nil {
			t.Errorf("Test setup failed: %s", err.Error())
		}
	}

	// test
	err := createDir()

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

func Test_writeFile_OneFile_Success(t *testing.T) {
	// setup
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := createDir()
		if err != nil {
			t.Errorf("Test setup failed: %s", err.Error())
		}
	}

	ch := make(chan string)
	var wg sync.WaitGroup
	s := []string{"test1", "test2", "test3"}

	// test
	wg.Add(1)
	writeFile(ch, 1, s, "00000", &wg)
	// wg.Wait()

	if len(ch) > 0 {
		t.Errorf("writeFile returned errors through a channel")
	}

	c := readFile(fmt.Sprintf("%s/sb_output_000001.txt", dirName))

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
