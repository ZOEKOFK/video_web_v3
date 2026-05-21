package grpc

import (
	"context"
	"errors"
	"log"
	"strconv"

	"google.golang.org/grpc/metadata"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	socialpb "github.com/ZOEKOFK/video_web_v3/app/pb/social"
	"github.com/ZOEKOFK/video_web_v3/app/usecase"
)

type ChatGrpc struct {
	socialpb.UnimplementedChatServiceServer
	chatUseCase usecase.ChatUseCase
}

func NewChatGrpc(chatUseCase usecase.ChatUseCase) *ChatGrpc {
	return &ChatGrpc{
		chatUseCase: chatUseCase,
	}
}

func GetUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, nil
	}

	userIDs := md.Get(UserIDMetadataKey)
	if len(userIDs) == 0 || userIDs[0] == "" {
		return 0, nil
	}

	userID, err := strconv.ParseUint(userIDs[0], 10, 64)
	if err != nil {
		log.Printf("[ERROR] Failed to parse user_id from context: %v", err)
		return 0, err
	}

	return userID, nil
}

func (s *ChatGrpc) SendMessage(ctx context.Context, in *socialpb.SendMessageRequest) (*common.CommonResponse, error) {
	senderID := uint(in.SenderId)

	contextUserID, err := GetUserIDFromContext(ctx)
	if err == nil && contextUserID > 0 {
		if in.SenderId != int64(contextUserID) {
			log.Printf("[WARNING] Sender ID mismatch: request=%d, token=%d", in.SenderId, contextUserID)
			return FailResponse("SendMessage", common.ErrorCode_OPERATION_FORBIDDEN, errors.New("sender_id does not match authenticated user")), nil
		}
		senderID = uint(contextUserID)
	}

	log.Printf("[SendMessage] sender=%d, receiver=%d, content=%s", senderID, in.ReceiverId, in.Content)

	msg, err := s.chatUseCase.SendMessage(senderID, uint(in.ReceiverId), in.Content, model.MessageType(in.Type))
	if err != nil {
		return FailResponse("SendMessage", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	return SuccessResponse("SendMessage", &common.Data{
		ChatMessage: &common.ChatMessage{
			Id:         int64(msg.ID),
			SenderId:   int64(msg.SenderID),
			ReceiverId: int64(msg.ReceiverID),
			Content:    msg.Content,
			Timestamp:  msg.Timestamp,
			Type:       int32(msg.Type),
			IsAi:       msg.IsAI,
		},
	}), nil
}

func (s *ChatGrpc) GetChatHistory(ctx context.Context, in *socialpb.GetChatHistoryRequest) (*common.CommonResponse, error) {
	log.Printf("[GetChatHistory] user=%d, peer=%d", in.UserId, in.PeerId)

	messages, err := s.chatUseCase.GetChatHistory(uint(in.UserId), uint(in.PeerId), int(in.Page), int(in.PageSize))
	if err != nil {
		return FailResponse("GetChatHistory", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	var msgResponses []*common.ChatMessage
	for _, msg := range messages {
		msgResponses = append(msgResponses, &common.ChatMessage{
			Id:         int64(msg.ID),
			SenderId:   int64(msg.SenderID),
			ReceiverId: int64(msg.ReceiverID),
			Content:    msg.Content,
			Timestamp:  msg.Timestamp,
			Type:       int32(msg.Type),
			IsAi:       msg.IsAI,
		})
	}

	return SuccessResponse("GetChatHistory", &common.Data{
		ChatMessageList: msgResponses,
	}), nil
}

func (s *ChatGrpc) GetChatSessions(ctx context.Context, in *socialpb.ChatSessionRequest) (*common.CommonResponse, error) {
	log.Printf("[GetChatSessions] user=%d", in.UserId)

	sessions, err := s.chatUseCase.GetChatSessions(uint(in.UserId))
	if err != nil {
		return FailResponse("GetChatSessions", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	var sessionResponses []*common.ChatSession
	for _, session := range sessions {
		sessionResponses = append(sessionResponses, &common.ChatSession{
			PeerId:      int64(session.PeerID),
			LastMessage: session.LastMessage,
			LastTime:    session.LastTime,
			UnreadCount: int32(session.UnreadCount),
		})
	}

	return SuccessResponse("GetChatSessions", &common.Data{
		ChatSessionList: sessionResponses,
	}), nil
}

func (s *ChatGrpc) GetUnreadMessages(ctx context.Context, in *socialpb.GetUnreadMessagesRequest) (*common.CommonResponse, error) {
	log.Printf("[GetUnreadMessages] user=%d", in.UserId)

	messages, err := s.chatUseCase.GetUnreadMessages(uint(in.UserId))
	if err != nil {
		return FailResponse("GetUnreadMessages", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	senderIDs := make(map[uint]bool)
	var msgResponses []*common.ChatMessage
	for _, msg := range messages {
		senderIDs[uint(msg.SenderID)] = true
		msgResponses = append(msgResponses, &common.ChatMessage{
			Id:         int64(msg.ID),
			SenderId:   int64(msg.SenderID),
			ReceiverId: int64(msg.ReceiverID),
			Content:    msg.Content,
			Timestamp:  msg.Timestamp,
			Type:       int32(msg.Type),
			IsAi:       msg.IsAI,
		})
	}

	for senderID := range senderIDs {
		s.chatUseCase.ClearUnread(uint(in.UserId), senderID)
	}

	return SuccessResponse("GetUnreadMessages", &common.Data{
		ChatMessageList: msgResponses,
	}), nil
}
