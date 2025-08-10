package db

import (
	"time"

	"gorm.io/datatypes"
)

type Image struct {
	ID          uint   `gorm:"primaryKey"`
	Path        string `gorm:"uniqueIndex;not null"`
	FileName    string `gorm:"not null"`
	Ext         string `gorm:"not null"`
	SizeBytes   int64  `gorm:"not null"`
	SHA256      string `gorm:"uniqueIndex;not null"`
	Width       *int
	Height      *int
	CreatedTime *time.Time
	ImportedAt  time.Time `gorm:"autoCreateTime"`

	SourceApp      *string
	ModelName      *string
	ModelHash      *string
	Prompt         *string
	NegativePrompt *string
	Sampler        *string
	Steps          *int
	CFGScale       *float64
	Seed           *string
	Scheduler      *string
	ClipSkip       *int

	NSFW   bool `gorm:"default:false"`
	Hidden bool `gorm:"default:false"`

	RawMetadata datatypes.JSON

	Tags     []*Tag         `gorm:"many2many:image_tags;constraint:OnDelete:CASCADE"`
	UserMeta []UserMetadata `gorm:"constraint:OnDelete:CASCADE"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex;not null"`
}

type ImageTag struct {
	ImageID uint `gorm:"primaryKey"`
	TagID   uint `gorm:"primaryKey"`
}

type UserMetadata struct {
	ID      uint   `gorm:"primaryKey"`
	ImageID uint   `gorm:"index;not null"`
	Key     string `gorm:"not null"`
	Value   string `gorm:"not null"`
}
