package io

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

const (
	dirName = "sb-shovel-output"
	prefix  = "sb_output_"
)

func CreateDir() error {
	_, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		err := os.Mkdir(dirName, 0777)
		if err != nil {
			return err
		}
		return nil
	}

	if err != nil {
		return err
	}
	return nil
}

func ReadFile(dir string) [][]byte {
	f, err := os.Open(dir)

	if err == os.ErrNotExist {
		fmt.Println("ERROR: No file found at ", dir)
		return nil
	}

	if err != nil {
		fmt.Println("ERROR: ", err)
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*5120)
	scanner.Buffer(buf, 64*5120)
	data := [][]byte{}

	for scanner.Scan() {
		data = append(data, scanner.Bytes())
	}

	return data
}

func WriteFile(errChannel chan error, suffix int, data []string, wg *sync.WaitGroup) {
	defer wg.Done()

	suffixPattern := map[int]string{1: "00000", 2: "0000", 3: "000", 4: "00", 5: "0"}

	fileName := fmt.Sprintf("%s/%s%s%s.txt", dirName, prefix, suffixPattern[len(fmt.Sprint(suffix))], fmt.Sprint(suffix))
	file, err := os.Create(fileName)

	if err != nil {
		if strings.Contains(err.Error(), "The system cannot find the path specified") {
			errChannel <- fmt.Errorf("cannot locate output directory")
			return
		}
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 64*5120)

	for _, line := range data {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			errChannel <- err
			return
		}
	}

	err = writer.Flush()

	if err != nil {
		errChannel <- err
		return
	}
}
