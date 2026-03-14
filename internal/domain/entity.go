package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VideoStatus string

const (
	StatusPending    VideoStatus = "PENDING"
	StatusProcessing VideoStatus = "PROCESSING"
	StatusCompleted  VideoStatus = "COMPLETED"
	StatusError      VideoStatus = "ERROR"
)

type User struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"not null" json:"-"` // Hash da senha
	CreatedAt time.Time `json:"created_at"`
}

type Video struct {
	ID          string      `gorm:"primaryKey" json:"id"`
	UserID      string      `gorm:"index;not null" json:"user_id"`
	FileName    string      `gorm:"not null" json:"file_name"`
	S3InputKey  string      `json:"s3_input_key"`  
	S3OutputKey string      `json:"s3_output_key"` 
	Status      VideoStatus `gorm:"default:'PENDING'" json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (v *Video) BeforeCreate(tx *gorm.DB) (err error) {
	v.ID = uuid.New().String()
	return
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}