package db

import (
	"time"

	"gorm.io/datatypes"
)

type Image struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Path        string     `gorm:"uniqueIndex;not null" json:"path"`
	FileName    string     `gorm:"not null" json:"fileName"`
	Ext         string     `gorm:"not null" json:"ext"`
	SizeBytes   int64      `gorm:"not null" json:"sizeBytes"`
	SHA256      string     `gorm:"uniqueIndex;not null" json:"sha256"`
	Width       *int       `json:"width"`
	Height      *int       `json:"height"`
	CreatedTime *time.Time `json:"createdTime"`
	ImportedAt  time.Time  `gorm:"autoCreateTime" json:"importedAt"`

	SourceApp      *string  `json:"sourceApp"`
	ModelName      *string  `json:"modelName"`
	ModelHash      *string  `json:"modelHash"`
	Prompt         *string  `json:"prompt"`
	NegativePrompt *string  `json:"negativePrompt"`
	Sampler        *string  `json:"sampler"`
	Steps          *int     `json:"steps"`
	CFGScale       *float64 `json:"cfgScale"`
	Seed           *string  `json:"seed"`
	Scheduler      *string  `json:"scheduler"`
	ClipSkip       *int     `json:"clipSkip"`

	Rating int  `gorm:"default:0" json:"rating"`
	NSFW   bool `gorm:"default:false" json:"nsfw"`
	Hidden bool `gorm:"default:false" json:"hidden"`

	RawMetadata datatypes.JSON `json:"rawMetadata"`

	Loras    []Lora         `gorm:"constraint:OnDelete:CASCADE" json:"loras"`
	Tags     []*Tag         `gorm:"many2many:image_tags;constraint:OnDelete:CASCADE" json:"tags"`
	UserMeta []UserMetadata `gorm:"constraint:OnDelete:CASCADE" json:"userMeta"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"`
}

type ImageTag struct {
	ImageID uint `gorm:"primaryKey" json:"imageId"`
	TagID   uint `gorm:"primaryKey" json:"tagId"`
}

type Lora struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	ImageID uint   `gorm:"index;not null" json:"imageId"`
	Name    string `json:"name"`
	Hash    string `json:"hash"`
}

type UserMetadata struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	ImageID uint   `gorm:"index;not null" json:"imageId"`
	Key     string `gorm:"not null" json:"key"`
	Value   string `gorm:"not null" json:"value"`
}

type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `gorm:"not null" json:"value"`
}
