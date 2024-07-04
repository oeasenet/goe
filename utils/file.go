package utils

import (
	"io"
	"os"
)

func FilePathToIOReader(filePath string) (io.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
