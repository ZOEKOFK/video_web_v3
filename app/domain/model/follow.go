package model

import "github.com/jinzhu/gorm"

type Follow struct {
	gorm.Model
	FollowerID uint `gorm:"column:follower_id;unique_index:idx_follow" json:"follower_id"`
	FollowedID uint `gorm:"column:followed_id;unique_index:idx_follow" json:"followed_id"`
}

func (Follow) TableName() string {
	return "follows"
}
