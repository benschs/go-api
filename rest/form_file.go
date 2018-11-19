package rest

import (
	"net/http"
	"strings"
)

// FileInfo of a file found in a multiPart form request.
type FileInfo struct {
	FileSize    int64
	FileName    string
	ContentType string
	File        []byte
}

type MultipartFormValue string

// ParseRequestFormFile parses the multipart form of the request and looks for a file in the provided formName.
// Returns the file, fileName, content type and file size.
func ParseRequestFormFile(r *http.Request, formName string) (*FileInfo, error) {
	fi := &FileInfo{}

	// Parse request form
	e := r.ParseMultipartForm(32 << 20)
	if e != nil {
		return nil, e
	}

	// Get file from request
	file, fileHeader, e := r.FormFile(formName)
	if e != nil {
		return nil, e
	}
	defer file.Close()

	// Get file size
	fi.FileSize, e = file.Seek(0, 2)
	if e != nil {
		return nil, e
	}
	// Reset file read to beginning
	_, e = file.Seek(0, 0)
	if e != nil {
		return nil, e
	}

	// Read file
	buf := make([]byte, fi.FileSize)
	_, e = file.Read(buf)
	if e != nil {
		return nil, e
	}
	// Reset file read to beginning
	_, e = file.Seek(0, 0)
	if e != nil {
		return nil, e
	}

	fileSplit := strings.Split(fileHeader.Filename, ".")
	if len(fileSplit) < 2 {
		return nil, e
	}

	// Get file name
	fi.FileName = fileHeader.Filename

	// Check file type
	buffer := make([]byte, 512)
	n, e := file.Read(buffer)
	if e != nil {
		return nil, e
	}
	// Reset file read to beginning
	_, e = file.Seek(0, 0)
	if e != nil {
		return nil, e
	}

	// Always returns a valid content-type and "application/octet-stream" if no others seemed to match.
	fi.ContentType = http.DetectContentType(buffer[:n])
	fi.File = buf

	return fi, nil
}
