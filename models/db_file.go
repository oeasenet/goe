package models

import "go.oease.dev/goe/modules/mongodb"

func (u *GoeFile) ColName() string {
	return "goe_files"
}

type FileType string

const (
	FileTypeImage       FileType = "image"
	FileTypeVideo       FileType = "video"
	FileTypeAudio       FileType = "audio"
	FileTypeArchive     FileType = "archive"
	FileTypeApplication FileType = "application"
	FileTypeDocument    FileType = "document"
	FileTypeOther       FileType = "other"
)

type GoeFile struct {
	mongodb.DefaultModel `bson:",inline"`
	Extension            string   `json:"extension" bson:"extension"`
	Filename             string   `json:"filename" bson:"filename"`
	Hash                 string   `json:"hash" bson:"hash"`
	MimeType             string   `json:"mime_type" bson:"mime_type"`
	Size                 int64    `json:"size" bson:"size"`
	Type                 FileType `json:"type" bson:"type"`
	UploadedName         string   `json:"uploaded_name" bson:"uploaded_name"`
}
