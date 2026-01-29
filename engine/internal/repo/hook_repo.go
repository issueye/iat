package repo

import (
	"iat/common/model"
	"iat/common/pkg/db"
)

type HookRepo struct{}

func NewHookRepo() *HookRepo {
	return &HookRepo{}
}

func (r *HookRepo) Create(hook *model.Hook) error {
	return db.DB.Create(hook).Error
}

func (r *HookRepo) Get(id uint) (*model.Hook, error) {
	var hook model.Hook
	if err := db.DB.First(&hook, id).Error; err != nil {
		return nil, err
	}
	return &hook, nil
}

func (r *HookRepo) List() ([]model.Hook, error) {
	var hooks []model.Hook
	if err := db.DB.Find(&hooks).Error; err != nil {
		return nil, err
	}
	return hooks, nil
}

func (r *HookRepo) Update(hook *model.Hook) error {
	return db.DB.Save(hook).Error
}

func (r *HookRepo) Delete(id uint) error {
	return db.DB.Delete(&model.Hook{}, id).Error
}

func (r *HookRepo) ListByType(hookType string) ([]model.Hook, error) {
	var hooks []model.Hook
	if err := db.DB.Where("type = ? AND enabled = ?", hookType, true).Find(&hooks).Error; err != nil {
		return nil, err
	}
	return hooks, nil
}

func (r *HookRepo) ListByTarget(targetType string, targetID uint) ([]model.Hook, error) {
	var hooks []model.Hook
	if err := db.DB.Where("target_type = ? AND target_id = ? AND enabled = ?", targetType, targetID, true).Find(&hooks).Error; err != nil {
		return nil, err
	}
	return hooks, nil
}
