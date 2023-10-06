package storage

import (
	"io"
	"time"
)

type File struct {
	Group      string
	Name       string
	Content    io.Reader
	UploadDate time.Time
}
