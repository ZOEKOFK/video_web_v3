package model

import (
	"github.com/jinzhu/gorm"
)

type Comment struct {
	gorm.Model
	UserID   uint      `gorm:"column:user_id;index" json:"user_id"`
	VideoID  uint      `gorm:"column:video_id;index" json:"video_id"`
	ParentID uint      `gorm:"column:parent_id;default:0;index" json:"parent_id"`
	Content  string    `gorm:"column:content;size:500" json:"content"`
	Likes    int64     `gorm:"column:likes;default:0" json:"likes"`
	User     *Users    `gorm:"-" json:"user_info,omitempty"`
	Replies  []Comment `gorm:"-" json:"reply_list,omitempty"`
}

func (c *Comment) TableName() string {
	return "comments"
}
