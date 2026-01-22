package grpc

import (
	"encoding/json"
	"iat/internal/service"
	pb "iat/pkg/pb/service"
)

type ChatHandler struct {
	pb.UnimplementedChatServiceServer
	svc *service.ChatService
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) Chat(req *pb.ChatRequest, stream pb.ChatService_ChatServer) error {
	eventChan := make(chan service.ChatEvent)

	// Call service
	err := h.svc.Chat(uint(req.SessionId), req.UserMessage, uint(req.AgentId), req.Mode, eventChan)
	if err != nil {
		close(eventChan)
		return err
	}

	for evt := range eventChan {
		extraJSON := ""
		if evt.Extra != nil {
			b, _ := json.Marshal(evt.Extra)
			extraJSON = string(b)
		}

		pbEvt := &pb.ChatEvent{
			Type:    string(evt.Type),
			Content: evt.Content,
			Extra:   extraJSON,
		}

		if err := stream.Send(pbEvt); err != nil {
			return err
		}
	}

	return nil
}
