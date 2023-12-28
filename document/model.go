package document

import (
	"time"
)

const (
	Ready = iota
	Downloaded
	Expired
	MaxFailedAttempts
)

type Document struct {
	ID              string     `gorm:"primaryKey;size:36"`
	Filename        string     `gorm:"not null;size:255"`
	FileContentType string     `gorm:"not null;columnName:file_content_type;size:255"`
	FileSize        int64      `gorm:"not null;columnName:file_size"`
	FailedAttempts  int        `gorm:"not null;columName:failed_attempts;default:0"`
	Status          int        `gorm:"not null;default:0"`
	UploadedAt      time.Time  `gorm:"not null;columName:uploaded_at"`
	DownloadedAt    *time.Time `gorm:"nullable;columName:downloaded_at"`
	UpdatedAt       *time.Time `gorm:"nullable;columName:updated_at"`
	Client          string     `gorm:"not null;size:65"`
}
