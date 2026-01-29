package repo

import (
	"iat/common/model"
	"iat/common/pkg/consts"
	"iat/common/pkg/db"
)

type MessageRepo struct{}

func NewMessageRepo() *MessageRepo {
	return &MessageRepo{}
}

func (r *MessageRepo) Create(m *model.Message) error {
	return db.DB.Create(m).Error
}

func (r *MessageRepo) ListBySessionID(sessionID uint) ([]model.Message, error) {
	var messages []model.Message
	err := db.DB.Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

func (r *MessageRepo) DeleteBySessionID(sessionID uint) error {
	return db.DB.Where("session_id = ?", sessionID).Delete(&model.Message{}).Error
}

func (r *MessageRepo) DeleteByID(id uint) error {
	return db.DB.Delete(&model.Message{}, id).Error
}

func (r *MessageRepo) UpsertToolCall(sessionID uint, toolCallID string, name string, arguments string) error {
	if toolCallID == "" {
		return db.DB.Create(&model.Message{
			SessionID:     sessionID,
			Role:          consts.RoleTool,
			Category:      consts.MessageCategoryTool,
			ToolCallID:    toolCallID,
			ToolName:      name,
			ToolArgs:      arguments,
			ToolHasResult: false,
			ToolOk:        false,
		}).Error
	}

	var existing model.Message
	err := db.DB.Where("session_id = ? AND role = ? AND tool_call_id = ?", sessionID, consts.RoleTool, toolCallID).First(&existing).Error
	if err == nil {
		existing.Category = consts.MessageCategoryTool
		existing.ToolName = name
		existing.ToolArgs = arguments
		return db.DB.Save(&existing).Error
	}
	// If not found, create new


	return db.DB.Create(&model.Message{
		SessionID:     sessionID,
		Role:          consts.RoleTool,
		Category:      consts.MessageCategoryTool,
		ToolCallID:    toolCallID,
		ToolName:      name,
		ToolArgs:      arguments,
		ToolHasResult: false,
		ToolOk:        false,
	}).Error
}

func (r *MessageRepo) UpsertToolResult(sessionID uint, toolCallID string, name string, output string, ok bool) error {
	if toolCallID == "" {
		return db.DB.Create(&model.Message{
			SessionID:     sessionID,
			Role:          consts.RoleTool,
			Category:      consts.MessageCategoryTool,
			ToolCallID:    toolCallID,
			ToolName:      name,
			ToolOutput:    output,
			ToolHasResult: true,
			ToolOk:        ok,
		}).Error
	}

	var existing model.Message
	err := db.DB.Where("session_id = ? AND role = ? AND tool_call_id = ?", sessionID, consts.RoleTool, toolCallID).First(&existing).Error
	if err == nil {
		existing.Category = consts.MessageCategoryTool
		existing.ToolName = name
		existing.ToolOutput = output
		existing.ToolHasResult = true
		existing.ToolOk = ok
		return db.DB.Save(&existing).Error
	}

	return db.DB.Create(&model.Message{
		SessionID:     sessionID,
		Role:          consts.RoleTool,
		Category:      consts.MessageCategoryTool,
		ToolCallID:    toolCallID,
		ToolName:      name,
		ToolOutput:    output,
		ToolHasResult: true,
		ToolOk:        ok,
	}).Error
}
