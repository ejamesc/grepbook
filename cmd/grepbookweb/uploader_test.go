package main_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	main "github.com/ejamesc/grepbook/cmd/grepbookweb"
	"github.com/julienschmidt/httprouter"
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

func TestUploaderUploadAndDelete(t *testing.T) {
	tmpPath, tmpFilename := "/tmp", "grepbookTest.jpg"
	u, err := app.CreateUploader(tmpPath)
	ok(t, err)

	_, err = u.Upload("badFileExtension.txt", strings.NewReader("quick brown fox"))
	assert(t, err == main.ErrUnacceptableFileExtension, "expect .txt upload to fail with ErrUnacceptableFileExtension, instead got %s", err)
	_, err = u.Upload("badFileExtension", strings.NewReader("quick brown fox"))
	assert(t, err == main.ErrUnacceptableFileExtension, "expect no extension file upload to fail with ErrUnacceptableFileExtension, instead got %s", err)

	savePath, err := u.Upload(tmpFilename, strings.NewReader("quick brown fox"))
	ok(t, err)

	// file should exist
	_, err = os.Stat(filepath.Join(tmpPath, savePath))
	ok(t, err)

	err = u.Delete(savePath)
	ok(t, err)

	// file should have been deleted
	_, err = os.Stat(filepath.Join(tmpPath, savePath))
	equals(t, true, os.IsNotExist(err))
}

type mockUploader struct {
	badExtension bool
	isFail       bool
}

func (u *mockUploader) Upload(filename string, fileReader io.Reader) (string, error) {
	if u.badExtension {
		return "", main.ErrUnacceptableFileExtension
	}
	if u.isFail {
		return "", fmt.Errorf("some err")
	}
	return "", nil
}

func (u *mockUploader) Delete(filename string) error {
	if u.isFail {
		return fmt.Errorf("some err")
	}
	return nil
}

func TestUploadHandler(t *testing.T) {
	mockUploader := &mockUploader{badExtension: false, isFail: false}
	uploadHandler := app.UploadHandler(mockUploader)
	bodyBuf, contentType, err := createFileUploadReader("file", "blah.jpg", []byte("ajsjfajfkalfjalisjd"))
	ok(t, err)

	test := GenerateHandleBodyTesterWithURLParams(t,
		app.Wrap(uploadHandler),
		true,
		httprouter.Params{},
	)
	w := test("POST", bodyBuf, contentType)
	assert(t, w.Code == http.StatusOK, "expect normal file upload to succeed, instead got %d", w.Code)

	mockUploader.badExtension = true
	bodyBuf, contentType, err = createFileUploadReader("file", "blah.txt", []byte("ajsjfajfkalfjalisjd"))
	ok(t, err)
	w = test("POST", bodyBuf, contentType)
	assert(t, w.Code == http.StatusBadRequest, "expect bad request when unacceptable file extension, instead got %d", w.Code)

}

func createFileUploadReader(key, filename string, contents []byte) (*bytes.Buffer, string, error) {
	bodyBuf := &bytes.Buffer{}
	multiWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := multiWriter.CreateFormFile(key, filename)
	if err != nil {
		return bodyBuf, "", err
	}

	fileWriter.Write(contents)
	multiWriter.Close()
	return bodyBuf, multiWriter.FormDataContentType(), nil
}
