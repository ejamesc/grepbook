package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var (
	ErrUnacceptableFileExtension = errors.New("not an acceptable file extension")
)

type LocalUploader struct {
	uploadFolder string
	logr         appLogger
}

type Uploader interface {
	Upload(filename string, fileReader io.Reader) error
	Delete(filename string) error
}

// CreateUploader takes in a full upload folder path and returns a LocalUploader if the folder exists.
func (a *App) CreateUploader(fullUploadFolderPath string) (*LocalUploader, error) {
	pathExists, err := isFilePathExists(fullUploadFolderPath)
	if !pathExists {
		return nil, os.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	return &LocalUploader{uploadFolder: fullUploadFolderPath, logr: a.logr}, nil
}

// Upload saves a file to the upload folder
func (u *LocalUploader) Upload(filename string, fileReader io.Reader) error {
	if !isAcceptedExtension(filepath.Ext(filename)) {
		return ErrUnacceptableFileExtension
	}
	loc := filepath.Join(u.uploadFolder, filename)
	out, err := os.Create(loc)
	if err != nil {
		return err
	}
	defer out.Close()

	written, err := io.Copy(out, fileReader)
	if err != nil {
		return err
	}
	if u.logr != nil {
		u.logr.Log("%s saved with %d bytes", loc, written)
	}

	return nil
}

// Delete simply deletes the given filename in the upload folder. No checking is done.
func (u *LocalUploader) Delete(filename string) error {
	return os.Remove(filepath.Join(u.uploadFolder, filename))
}

// UploadHandler is the file upload handler
func (a *App) UploadHandler(up Uploader) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		file, header, err := req.FormFile("file")
		if err != nil {
			a.rndr.JSON(w, http.StatusInternalServerError, &APIResponse{Message: "FormFile extraction failed"})
			return new500Error("error retrieving file: ", err)
		}
		defer file.Close()

		user := getUser(req)
		fpath := filepath.Join(a.UploadPath(), strconv.FormatUint(user.ID, 10), header.Filename)
		err = up.Upload(fpath, file)
		if err != nil {
			if err == ErrUnacceptableFileExtension {
				a.rndr.JSON(w, http.StatusBadRequest, &APIResponse{Message: "Only accept files that end in .jpg, .jpeg, .png and .gif"})
				return newError(http.StatusBadRequest, "file extension error: ", err)
			}
			return newError(http.StatusInternalServerError, "uploader error: ", err)
		}

		a.rndr.JSON(w, http.StatusOK, &APIResponse{Message: "File uploaded successfully"})
		return nil
	}
}

func isAcceptedExtension(ext string) bool {
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

func isFilePathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
