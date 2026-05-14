package model

import "github.com/jinzhu/gorm"

type Like struct {
	gorm.Model
	UserID   uint `gorm:"column:user_id;index" json:"user_id"`
	TargetID uint `gorm:"column:target_id;index" json:"target_id"`
	Type     int  `gorm:"column:type" json:"type"`
}

func (Like) TableName() string {
	return "likes"
}
