package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BaseModel struct {
	ID        string         `gorm:"primaryKey; unique; type:varchar(255)" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	base.ID = uuid.NewString()
	return
}

func (model *BaseModel) AfterCreate(tx *gorm.DB) (err error) {
	fmt.Println("okay base here")
	tx.Preload(clause.Associations).Find(model, "id = ?", model.ID)
	return
}
