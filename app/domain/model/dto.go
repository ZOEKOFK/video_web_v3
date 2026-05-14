package model

import (
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
)

func UserToPb(user *Users) *common.User {
	if user == nil {
		return nil
	}
	return &common.User{
		Id:        int64(user.ID),
		Username:  user.Username,
		AvatarUrl: user.Avatarurl,
		Nickname:  user.Nickname,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func UserListToPb(users []*Users) []*common.User {
	pbUsers := make([]*common.User, 0, len(users))
	for _, u := range users {
		pbUsers = append(pbUsers, UserToPb(u))
	}
	return pbUsers
}

func VideoToPb(video *Videos) *common.Video {
	if video == nil {
		return nil
	}
	return &common.Video{
		Id:          int64(video.ID),
		Title:       video.Title,
		Description: video.Description,
		VideoUrl:    video.VideoUrl,
		Views:       int32(video.Views),
		Likes:       int32(video.Likes),
		Comments:    int32(video.Comments),
		CreatedAt:   video.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   video.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func VideoListToPb(videos []*Videos) []*common.Video {
	pbVideos := make([]*common.Video, 0, len(videos))
	for _, v := range videos {
		pbVideos = append(pbVideos, VideoToPb(v))
	}
	return pbVideos
}

func CommentToPb(comment *Comment) *common.Comment {
	if comment == nil {
		return nil
	}
	pbComment := &common.Comment{
		Id:        int64(comment.ID),
		UserId:    int64(comment.UserID),
		VideoId:   int64(comment.VideoID),
		ParentId:  int64(comment.ParentID),
		Content:   comment.Content,
		Likes:     comment.Likes,
		CreatedAt: comment.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: comment.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if comment.User != nil {
		pbComment.UserInfo = UserToPb(comment.User)
	}
	if len(comment.Replies) > 0 {
		pbComment.ReplyList = make([]*common.Comment, 0, len(comment.Replies))
		for i := range comment.Replies {
			pbComment.ReplyList = append(pbComment.ReplyList, CommentToPb(&comment.Replies[i]))
		}
	}
	return pbComment
}

func CommentListToPb(comments []*Comment) []*common.Comment {
	pbComments := make([]*common.Comment, 0, len(comments))
	for _, c := range comments {
		pbComments = append(pbComments, CommentToPb(c))
	}
	return pbComments
}

func FollowToUser(follow *Follow) *common.User {
	if follow == nil {
		return nil
	}
	return &common.User{
		Id:        int64(follow.FollowedID),
		CreatedAt: follow.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func FollowerToUser(follow *Follow) *common.User {
	if follow == nil {
		return nil
	}
	return &common.User{
		Id:        int64(follow.FollowerID),
		CreatedAt: follow.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func FollowListToUsers(follows []*Follow, isFollower bool) []*common.User {
	pbUsers := make([]*common.User, 0, len(follows))
	for _, f := range follows {
		if isFollower {
			pbUsers = append(pbUsers, FollowerToUser(f))
		} else {
			pbUsers = append(pbUsers, FollowToUser(f))
		}
	}
	return pbUsers
}
