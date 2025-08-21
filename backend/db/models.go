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

	SourceApp                *string  `json:"sourceApp"`
	ModelID                  *uint    `json:"modelId"`
	Model                    *Model   `json:"model"`
	ModelName                *string  `gorm:"-" json:"modelName,omitempty"`
	ModelHash                *string  `gorm:"-" json:"modelHash,omitempty"`
	Prompt                   *string  `json:"prompt"`
	NegativePrompt           *string  `json:"negativePrompt"`
	Sampler                  *string  `json:"sampler"`
	Steps                    *int     `json:"steps"`
	CFGScale                 *float64 `json:"cfgScale"`
	Seed                     *string  `json:"seed"`
	Scheduler                *string  `json:"scheduler"`
	ClipSkip                 *int     `json:"clipSkip"`
	VariationSeed            *int     `json:"variationSeed"`
	VariationSeedStrength    *float64 `json:"variationSeedStrength"`
	AspectRatio              *string  `json:"aspectRatio"`
	RefinerControlPercentage *float64 `json:"refinerControlPercentage"`
	RefinerUpscale           *float64 `json:"refinerUpscale"`
	RefinerUpscaleMethod     *string  `json:"refinerUpscaleMethod"`

	Rating int  `gorm:"default:0" json:"rating"`
	NSFW   bool `gorm:"default:false" json:"nsfw"`
	Hidden bool `gorm:"default:false" json:"hidden"`

	RawMetadata datatypes.JSON `json:"rawMetadata"`

	Loras      []*Lora      `gorm:"many2many:image_loras;constraint:OnDelete:CASCADE" json:"loras"`
	Embeddings []*Embedding `gorm:"many2many:image_embeddings;constraint:OnDelete:CASCADE" json:"embeddings"`
	Tags       []*Tag       `gorm:"many2many:image_tags;constraint:OnDelete:CASCADE" json:"tags"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"uniqueIndex;not null" json:"name"`
}

type ImageTag struct {
	ImageID uint `gorm:"primaryKey" json:"imageId"`
	TagID   uint `gorm:"primaryKey" json:"tagId"`
}

type ImageLora struct {
	ImageID uint     `gorm:"primaryKey" json:"imageId"`
	LoraID  uint     `gorm:"primaryKey" json:"loraId"`
	Weight  *float64 `json:"weight"`
}

type ImageEmbedding struct {
	ImageID     uint `gorm:"primaryKey" json:"imageId"`
	EmbeddingID uint `gorm:"primaryKey" json:"embeddingId"`
}

type Model struct {
	ID   uint    `gorm:"primaryKey" json:"id"`
	Name string  `gorm:"uniqueIndex;not null" json:"name"`
	Hash *string `gorm:"index" json:"hash"`
}

type Lora struct {
	ID     uint     `gorm:"primaryKey" json:"id"`
	Name   string   `gorm:"uniqueIndex;not null" json:"name"`
	Hash   *string  `gorm:"index" json:"hash"`
	Weight *float64 `gorm:"->;column:weight" json:"weight,omitempty"`
}

type Embedding struct {
	ID   uint    `gorm:"primaryKey" json:"id"`
	Name string  `gorm:"uniqueIndex;not null" json:"name"`
	Hash *string `gorm:"index" json:"hash"`
}

type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `gorm:"not null" json:"value"`
}
