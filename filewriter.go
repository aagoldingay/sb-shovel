package main

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

func createDir() error {
	_, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		err := os.Mkdir(dirName, 0666)
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

func readFile(dir string) [][]byte {
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

func writeFile(errChannel chan string, i int, data []string, fileSuffix string, wg *sync.WaitGroup) {
	defer wg.Done()
	fileName := fmt.Sprintf("%s/%s%s%s.txt", dirName, prefix, fileSuffix, fmt.Sprint(i))
	file, err := os.Create(fileName)
	if err != nil {
		if strings.Contains(err.Error(), "The system cannot find the path specified") {
			errChannel <- "ERROR: Cannot locate output directory."
			return
		}
	}
	defer file.Close()

	writer := bufio.NewWriterSize(file, 64*1024)
	for _, line := range data {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	err = writer.Flush()
	if err != nil {
		errChannel <- err.Error()
		return
	}
}
