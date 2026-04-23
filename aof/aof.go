package aof

import (
	"bufio"
	"fmt"
	"os"
	"redis/resp"
	"sync"
	"time"
)

type AOF struct {
	File   *os.File
	reader *bufio.Reader
	mu     sync.Mutex
}

func NewAOF(path string) (*AOF, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	aof := AOF{
		File:   file,
		reader: bufio.NewReader(file),
		mu:     sync.Mutex{},
	}
	go func() {
		for {
			aof.mu.Lock()

			aof.File.Sync()

			aof.mu.Unlock()

			time.Sleep(time.Second)
		}
	}()

	return &aof, nil
}

func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.File.Close()
}

func (a *AOF) Write(value resp.Value) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, err := a.File.Write(value.Marshal())
	if err != nil {
		fmt.Println("Error in writing in file")
		return err
	}
	return nil
}
