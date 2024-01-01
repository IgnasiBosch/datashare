package document

import (
	"dataShare/core"
	"dataShare/service"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	dataFolder          = "./datafiles/"
	maxUploadFileSize   = 100 << 20 // 100 MiB
	totalUploadsPerHour = 5
)

type Handler struct {
	c  echo.Context
	DB *gorm.DB
	e  *service.Encryption
}

func NewHandler(c echo.Context, db *gorm.DB, e *service.Encryption) *Handler {
	return &Handler{c: c, DB: db, e: e}
}

// getTotalFileSize calculates the total file size of multiple multipart.FileHeaders.
func getTotalFileSize(files []*multipart.FileHeader) int {
	totalFileSize := 0
	for _, v := range files {
		totalFileSize += int(v.Size)
	}
	return totalFileSize
}

func (h *Handler) validateFiles(form *multipart.Form) ([]*multipart.FileHeader, *core.Error) {

	files := form.File["files"]
	if totalFileSize := getTotalFileSize(files); totalFileSize > maxUploadFileSize {
		return nil, core.NewError(http.StatusBadRequest, 1005, "File size is too large")
	}
	if len(files) == 0 {
		return nil, core.NewError(http.StatusBadRequest, 1010, "No files uploaded")
	}
	return files, nil
}

func (h *Handler) Encrypt(form *multipart.Form) (*core.IDKey, error) {

	files, e := h.validateFiles(form)
	if e != nil {
		return nil, e
	}

	file, err := service.GetFileFromFileHeader(files)
	if err != nil {
		return nil, core.NewError(http.StatusBadRequest, 1020, "Can't get file from header")
	}

	ipAddr := h.c.RealIP()
	client := h.e.HashString(ipAddr)
	dr := NewRepositoryImp(h.DB)
	total, err := dr.GetTotalUsage(client)
	if err != nil {
		return nil, core.NewError(http.StatusUnprocessableEntity, 1030, "Can't process file")
	}
	if total >= totalUploadsPerHour {
		return nil, core.NewError(http.StatusUnprocessableEntity, 1040, "Allowed number of uploads per hour exceeded")
	}

	document := Document{
		ID:              service.NewID("doc"),
		Filename:        file.Name,
		FileSize:        file.Size,
		FileContentType: file.ContentType,
		Status:          Ready,
		UploadedAt:      time.Now(),
		Client:          client,
	}

	passphrase := service.NewKey()
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		dr := NewRepositoryImp(tx)
		err = dr.Save(&document)
		if err != nil {
			return err
		}

		documentContent, err := h.e.Encrypt(passphrase, file.Content)
		if err != nil {
			return err
		}
		targetPath := dataFolder + document.ID
		targetFile, err := os.Create(targetPath)
		defer targetFile.Close()
		if err != nil {
			return err
		}
		err = service.SaveFile(targetFile, documentContent)

		return err

	})
	if err != nil {
		return nil, core.NewError(http.StatusUnprocessableEntity, 1030, "Can't process file")
	}

	return &core.IDKey{
		ID:  document.ID,
		Key: passphrase,
	}, nil
}

func (h *Handler) Check(ID string) error {
	dr := NewRepositoryImp(h.DB)
	d, err := dr.FindById(ID)
	if err != nil {
		return core.NewError(http.StatusUnprocessableEntity, 2000, "Can't find document")
	}

	if d.Status == Downloaded {
		return core.NewError(http.StatusUnprocessableEntity, 2010, "Document was already downloaded")
	}

	if d.Status == Expired {
		return core.NewError(http.StatusUnprocessableEntity, 2020, "Document was expired")
	}

	if d.Status == MaxFailedAttempts {
		return core.NewError(http.StatusUnprocessableEntity, 2030, "Document reached max failed attempts")
	}

	return nil
}

func (h *Handler) Decrypt(ip *core.IDKey) ([]byte, *Document, error) {
	dr := NewRepositoryImp(h.DB)
	d, err := dr.FindById(ip.ID)
	if err != nil {
		return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2000, "Can't find document")
	}

	if d.Status == Downloaded {
		return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2010, "Document was already downloaded")
	}

	if d.Status == Expired {
		return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2020, "Document was expired")
	}

	if d.Status == MaxFailedAttempts {
		return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2030, "Document reached max failed attempts")
	}

	targetPath := dataFolder + ip.ID
	src, err := os.Open(targetPath)
	if err != nil {
		return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2040, "Can't open file")
	}
	ciphertext, err := io.ReadAll(src)
	if err != nil {
		return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2050, "Can't read file")
	}
	now := time.Now()
	documentContent, err := h.e.Decrypt(ip.Key, ciphertext)
	if err != nil {

		d.FailedAttempts++
		d.UpdatedAt = &now
		if d.FailedAttempts >= 3 {
			d.Status = MaxFailedAttempts
			err = os.Remove(targetPath)
			if err != nil {
				return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2060, "Can't remove file")
			}
		}

		err = h.DB.Transaction(func(tx *gorm.DB) error {
			dr := NewRepositoryImp(tx)
			err := dr.Update(d)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2070, "Can't update file")
		}

		return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2080, "Wrong key, try again")
	}

	if err != nil {
		return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2070, "Can't update file")
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		dr := NewRepositoryImp(tx)
		d.Status = Downloaded
		d.DownloadedAt = &now
		d.UpdatedAt = &now
		err := dr.Update(d)
		if err != nil {
			return err
		}
		err = os.Remove(targetPath)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, nil, core.NewError(http.StatusUnprocessableEntity, 2080, "Can't remove file")
	}

	//h.c.Response().WriteHeader("Content-Disposition", "attachment; filename="+d.Filename)
	//h.c.Response().WriteHeader("Content-Type", d.FileContentType)
	//h.c.Response().WriteHeader("Accept-Length", fmt.Sprintf("%d", d.FileSize))
	//_, err = h.c.Writer.Write(documentContent)
	//if err != nil {
	//	return core.NewError(http.StatusUnprocessableEntity, 2080, "Can't remove file")
	//}

	return documentContent, d, nil

}
