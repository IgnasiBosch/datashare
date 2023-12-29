package service

import (
	"archive/zip"
	"bytes"
	"io"
	"mime/multipart"
)

const zipFileName = "archive.zip"

type File struct {
	Name        string
	Content     []byte
	Size        int64
	ContentType string
}

func SaveFile(w io.Writer, content []byte) error {
	_, err := w.Write(content)
	return err
}

func GetFileFromFileHeader(files []*multipart.FileHeader) (*File, error) {
	if len(files) == 1 {
		return getFileFromSingleFileHeader(files[0])
	}
	return getFileFromMultipleFileHeader(files)
}

func getFileFromSingleFileHeader(file *multipart.FileHeader) (*File, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	return &File{
		Name:        file.Filename,
		Content:     content,
		Size:        file.Size,
		ContentType: file.Header.Get("Content-Type"),
	}, nil
}

func getFileFromMultipleFileHeader(files []*multipart.FileHeader) (*File, error) {
	content, err := compressFiles(files)
	if err != nil {
		return nil, err
	}

	return &File{
		Name:        zipFileName,
		Content:     content,
		Size:        int64(len(content)),
		ContentType: "application/zip",
	}, nil

}

// CompressFiles receives a list of files and returns a byte array representing the ZIP file.
func compressFiles(files []*multipart.FileHeader) ([]byte, error) {
	buff := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buff)

	// Process each file.
	for _, file := range files {
		i, err2 := addToZip(zipWriter, file)
		if err2 != nil {
			return i, err2
		}
	}

	// All files have been written to the zip file; we can close it.
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	// Return the buffer's contents (the complete zip file) as a byte array.
	return buff.Bytes(), nil
}

func addToZip(zipWriter *zip.Writer, file *multipart.FileHeader) ([]byte, error) {
	// Add file to zip
	fileWriter, err := zipWriter.Create(file.Filename)
	if err != nil {
		return nil, err
	}

	// Open the file, copy its content into the zip file.
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	if _, err := io.Copy(fileWriter, src); err != nil {
		return nil, err
	}
	return nil, nil
}
