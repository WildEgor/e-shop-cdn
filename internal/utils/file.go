package utils

import (
	"errors"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"strings"
)

// SanitizeFilename remove some symbols from filename
func SanitizeFilename(filename string) (string, error) {
	if countVal(filename, ".") > 1 {
		return filename, errors.New("filename cannot contain more than one period character")
	}

	// TODO: use regex
	return strings.Replace(strings.Replace(filename, "/", "", -1), `\`, "", -1), nil
}

// ExtractFilename return filename from path
func ExtractFilename(path string) (string, error) {

	p := strings.Split(path, "/")

	if len(p) == 0 {
		return "", errors.New("empty path")
	}

	return p[len(p)-1], nil
}

// UniqueFilename create unique filename
func UniqueFilename(filename string) string {
	parts := strings.Split(filename, ".")
	ext := ""
	bfn := parts[0]

	if len(parts) > 1 {
		ext = parts[len(parts)-1]
		bfn = strings.Join(parts[:len(parts)-1], ".")
	}

	ufn := strings.Join([]string{uuid.New().String(), bfn}, "_")
	if ext != "" {
		ufn += "." + ext
	}
	return ufn
}

// countVal count how many one string in another
func countVal(str string, val string) int {
	var count int

	for _, v := range strings.Split(str, "") {
		if v == val {
			count++
		}
	}

	return count
}

func ReadFileToBuffer(header *multipart.FileHeader) ([]byte, error) {
	fr, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer fr.Close()
	
	fbuf := make([]byte, 512)
	for {
		fbuf = fbuf[:cap(fbuf)]
		if _, err := fr.Read(fbuf); err != nil {
			if err != io.EOF {
				return nil, err
			}
		}
		break
	}

	return fbuf, err
}
