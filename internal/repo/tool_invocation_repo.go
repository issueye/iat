package repo

import (
	"iat/internal/model"
	"iat/internal/pkg/db"
)

type ToolInvocationRepo struct{}

func NewToolInvocationRepo() *ToolInvocationRepo {
	return &ToolInvocationRepo{}
}

func (r *ToolInvocationRepo) ListBySessionID(sessionID uint) ([]model.ToolInvocation, error) {
	var items []model.ToolInvocation
	err := db.DB.Where("session_id = ?", sessionID).Order("created_at asc").Find(&items).Error
	return items, err
}

func (r *ToolInvocationRepo) DeleteBySessionID(sessionID uint) error {
	return db.DB.Where("session_id = ?", sessionID).Delete(&model.ToolInvocation{}).Error
}

func (r *ToolInvocationRepo) UpsertCall(sessionID uint, toolCallID string, name string, arguments string) error {
	if toolCallID == "" {
		return db.DB.Create(&model.ToolInvocation{
			SessionID:  sessionID,
			ToolCallID: "",
			Name:       name,
			Arguments:  arguments,
			HasResult:  false,
			Ok:         false,
		}).Error
	}

	var existing model.ToolInvocation
	err := db.DB.Where("session_id = ? AND tool_call_id = ?", sessionID, toolCallID).First(&existing).Error
	if err == nil {
		existing.Name = name
		existing.Arguments = arguments
		return db.DB.Save(&existing).Error
	}
	// If not found, create new


	return db.DB.Create(&model.ToolInvocation{
		SessionID:  sessionID,
		ToolCallID: toolCallID,
		Name:       name,
		Arguments:  arguments,
		HasResult:  false,
		Ok:         false,
	}).Error
}

func (r *ToolInvocationRepo) UpsertResult(sessionID uint, toolCallID string, name string, output string, ok bool) error {
	if toolCallID == "" {
		return db.DB.Create(&model.ToolInvocation{
			SessionID:  sessionID,
			ToolCallID: "",
			Name:       name,
			Arguments:  "",
			Output:     output,
			HasResult:  true,
			Ok:         ok,
		}).Error
	}

	var existing model.ToolInvocation
	err := db.DB.Where("session_id = ? AND tool_call_id = ?", sessionID, toolCallID).First(&existing).Error
	if err == nil {
		existing.Name = name
		existing.Output = output
		existing.HasResult = true
		existing.Ok = ok
		return db.DB.Save(&existing).Error
	}

	return db.DB.Create(&model.ToolInvocation{
		SessionID:  sessionID,
		ToolCallID: toolCallID,
		Name:       name,
		Arguments:  "",
		Output:     output,
		HasResult:  true,
		Ok:         ok,
	}).Error
}
