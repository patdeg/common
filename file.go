package common

import (
	"io"
	"os"

	"golang.org/x/net/context"
)

func GetContent(c context.Context, filename string) (*[]byte, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		Error("Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()
	Info("FILE FOUND : %s", filename)
	buffer := make([]byte, 10*1024*1024)
	n, err := file.Read(buffer)
	if (err == nil) || (err == io.EOF) {
		content := buffer[:n]
		return &content, nil
	}

	Error("Error reading file: %v", err)
	return nil, err

}
