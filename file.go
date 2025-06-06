package common

import (
	"io"
	"os"

	"golang.org/x/net/context"
)

// GetContent reads the file named by filename and returns its contents.
// Any errors encountered are logged and returned.
func GetContent(c context.Context, filename string) (*[]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		Error("Error opening file %s: %v", filename, err)
		return nil, err
	}
	defer file.Close()

	Info("FILE FOUND : %s", filename)
	content, err := io.ReadAll(file)
	if err != nil {
		Error("Error reading file %s: %v", filename, err)
		return nil, err
	}

	return &content, nil

}
