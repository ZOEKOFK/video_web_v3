package model

import "github.com/jinzhu/gorm"

type Users struct {
	gorm.Model
	Username  string `gorm:"column:username" json:"username"`
	Password  string `gorm:"column:password" json:"password"`
	Nickname  string `gorm:"column:nickname" json:"nickname"`
	Avatarurl string `gorm:"column:avatar_url" json:"avatar_url"`
}

func (Users) TableName() string {
	return "users"
}
