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

type mockUploader struct {
	badExtension bool
	isFail       bool
}

func (u *mockUploader) Upload(filename string, fileReader io.Reader) error {
	if u.badExtension {
		return main.ErrUnacceptableFileExtension
	}
	if u.isFail {
		return fmt.Errorf("some err")
	}
	return nil
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
	bodyBuf := &bytes.Buffer{}
	multiWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := multiWriter.CreateFormFile("file", "blah.jpg")
	ok(t, err)

	fileWriter.Write([]byte("asfajgoiejofcaogjiaofioaejfi"))
	multiWriter.Close()
	test := GenerateHandleBodyTesterWithURLParams(t,
		app.Wrap(uploadHandler),
		true,
		httprouter.Params{},
		multiWriter.FormDataContentType(),
	)
	w := test("POST", bodyBuf)
	assert(t, w.Code == http.StatusOK, "expect normal file upload to succeed, instead got %d", w.Code)

}
