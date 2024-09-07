package filehandlers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// by having file and fileReader as global, we limit the number of times a file is opened or closed, increasing efficiency
// it also helps in making sure file descriptors do not go out of control when a lot of concurrent requests are made
// since we are using RWMutex, it is thread safe, allowing concurrent reads but not read & writes together
var (
	file       *os.File
	fileReader *bufio.Reader
	fileMutex  sync.RWMutex
)

func InitializeFile(filePath string) error {
	var err error

	if filePath == "" {
		// Create a temporary file
		file, err = os.CreateTemp("", "shared_content_*.txt")
		if err != nil {
			return fmt.Errorf("error creating temporary file: %w", err)
		}
		log.Printf("Created temporary file: %s", file.Name())
	} else {
		// Open the specified file
		file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("error opening file %s: %w", filePath, err)
		}
		log.Printf("Opened file: %s", filePath)
	}

	go WriteToFile()
	fileReader = bufio.NewReader(file)
	return nil
}

func WriteToFile() {
	reader := bufio.NewReader(os.Stdin)
	buffer := make([]byte, 1024*1024) // 1mb buffer

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("EOF reached")
				return
			}
			log.Printf("Error reading from stdin: %v", err)
			continue
		}

		fileMutex.Lock()
		_, err = file.Write(buffer[:n])
		if err != nil {
			log.Printf("Error writing to file: %v", err)
		}

		fileMutex.Unlock()
		// all the reads who were earlier blocked will now start reading data
	}
}

func CloseFile() {
	if file != nil {
		file.Close()
	}
}

func GetFileSize() int64 {
	fileMutex.RLock()
	defer fileMutex.RUnlock()

	info, err := file.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}

func SeekFile(offset int64) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	_, err := file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	fileReader.Reset(file)
	return nil
}

func ReadChunk(ctx context.Context, buffer []byte) (int, error) {
	fileMutex.RLock()
	defer fileMutex.RUnlock()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		return fileReader.Read(buffer)
	}
}
