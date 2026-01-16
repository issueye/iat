package service

import (
	"iat/internal/model"
	"iat/internal/pkg/script"
	"iat/internal/repo"
)

type ScriptService struct {
	repo   *repo.ScriptRepo
	engine *script.ScriptEngine
}

func NewScriptService() *ScriptService {
	return &ScriptService{
		repo:   repo.NewScriptRepo(),
		engine: script.NewScriptEngine(),
	}
}

func (s *ScriptService) CreateScript(name, description, content string) error {
	script := &model.Script{
		Name:        name,
		Description: description,
		Content:     content,
	}
	return s.repo.Create(script)
}

func (s *ScriptService) UpdateScript(id uint, name, description, content string) error {
	script, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	script.Name = name
	script.Description = description
	script.Content = content
	return s.repo.Update(script)
}

func (s *ScriptService) DeleteScript(id uint) error {
	return s.repo.Delete(id)
}

func (s *ScriptService) ListScripts() ([]model.Script, error) {
	return s.repo.List()
}

func (s *ScriptService) RunScript(id uint) (interface{}, error) {
	sc, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	val, err := s.engine.Run(sc.Content)
	if err != nil {
		return nil, err
	}
	return val.Export(), nil
}
