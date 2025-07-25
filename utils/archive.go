package utils

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
)

const (
	// MaxZipFileSize is the maximum size of a zip file in bytes
	// generated by the ZipInMemory function.
	MaxZipFileSize = 4 * 1024 * 1024 // 4MB
)

// ZipInMemory creates a zip archive in memory that contains a single file
// with the given file name and content.
// This returns a base64 encoded string representation of the zip archive.
func ZipInMemory(
	fileName string,
	content string,
) (string, error) {
	zipBuffer := bytes.NewBuffer(make([]byte, 0, MaxZipFileSize))
	zipWriter := zip.NewWriter(zipBuffer)
	indexFileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(indexFileWriter, bytes.NewReader([]byte(content)))
	if err != nil {
		return "", err
	}

	err = zipWriter.Close()
	if err != nil {
		return "", err
	}

	zipBytes, err := io.ReadAll(zipBuffer)
	if err != nil {
		return "", err
	}

	if len(zipBytes) > MaxZipFileSize {
		return "", fmt.Errorf(
			"zip file size is too large, the maximum size is %d MB",
			MaxZipFileSize/(1024*1024),
		)
	}
	return base64.StdEncoding.EncodeToString(zipBytes), nil
}
