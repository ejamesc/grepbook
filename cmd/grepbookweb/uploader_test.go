package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	main "github.com/ejamesc/grepbook/cmd/grepbookweb"
)

func TestCreateUploader(t *testing.T) {
	u, err := app.CreateUploader("/tmp")
	ok(t, err)
	assert(t, u != nil, "expect CreateUploader to return an Uploader, instead got %+v", u)

	badFolder := "/bal25u9eari3o"
	u, err = app.CreateUploader(badFolder)
	assert(t, err != nil, "expect error because %s shouldn't exist", badFolder)
	assert(t, u == nil, "expect CreateUploader to return nil, instead got %+v", u)
}

func TestUploaderUpload(t *testing.T) {
	tmpPath, tmpFilename := "/tmp", "grepbookTest.jpg"
	u, err := app.CreateUploader(tmpPath)
	ok(t, err)

	err = u.Upload("badFileExtension.txt", strings.NewReader("quick brown fox"))
	assert(t, err == main.ErrUnacceptableFileExtension, "expect .txt upload to fail with ErrUnacceptableFileExtension, instead got %s", err)
	err = u.Upload("badFileExtension", strings.NewReader("quick brown fox"))
	assert(t, err == main.ErrUnacceptableFileExtension, "expect no extension file upload to fail with ErrUnacceptableFileExtension, instead got %s", err)

	err = u.Upload(tmpFilename, strings.NewReader("quick brown fox"))
	ok(t, err)

	// file should exist
	_, err = os.Stat(filepath.Join(tmpPath, tmpFilename))
	ok(t, err)

	err = u.Delete(tmpFilename)
	ok(t, err)

	// file should have been deleted
	_, err = os.Stat(filepath.Join(tmpPath, tmpFilename))
	equals(t, true, os.IsNotExist(err))
}
