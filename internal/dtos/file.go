package dtos

import (
	"errors"
)

// FileResponseDto
type FileResponseDto struct {
	Filename    string `json:"filename"`
	DownloadUrl string `json:"download_url"`
}

type FileQueryDto struct {
	Filename string `json:"filename"`
}

func (q FileQueryDto) Validate() error {
	if q.Filename == "" {
		return errors.New("empty filename")
	}

	return nil
}

type FileIdDto struct {
	FileId string `json:"id" bson:"id"`
}

func (q FileIdDto) Validate() error {
	if q.FileId == "" {
		return errors.New("empty id")
	}

	return nil
}
