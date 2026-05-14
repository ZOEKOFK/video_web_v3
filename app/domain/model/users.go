package model

import "github.com/jinzhu/gorm"

type Users struct {
	gorm.Model
	Username   string `gorm:"column:username" json:"username"`
	Password   string `gorm:"column:password" json:"password"`
	Nickname   string `gorm:"column:nickname" json:"nickname"`
	Avatarurl  string `gorm:"column:avatar_url" json:"avatar_url"`
	MfaSecret  string `gorm:"column:mfa_secret" json:"-"`
	MfaEnabled bool   `gorm:"column:mfa_enabled;default:false" json:"mfa_enabled"`
}

func (Users) TableName() string {
	return "users"
}
