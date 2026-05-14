package model

import "github.com/jinzhu/gorm"

type Videos struct {
	gorm.Model
	UserID      uint   `gorm:"column:user_id" json:"user_id"`
	Title       string `gorm:"column:title" json:"title"`
	Description string `gorm:"column:description" json:"description"`
	VideoUrl    string `gorm:"column:video_url" json:"video_url"`
	Views       int    `gorm:"column:views;default:0" json:"views"`
	Comments    int    `gorm:"column:comments;default:0" json:"comments"`
	Likes       int    `gorm:"column:likes;default:0" json:"likes"`
}

func (Videos) TableName() string {
	return "videos"
}
