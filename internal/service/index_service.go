package service

import (
	"fmt"
	"iat/internal/model"
	"iat/internal/pkg/indexdb"
	"iat/internal/repo"
	"strings"
)

type SessionWithProject struct {
	model.Session
	ProjectName string `json:"projectName"`
}

type IndexService struct {
	projectRepo *repo.ProjectRepo
	sessionRepo *repo.SessionRepo
}

func NewIndexService() *IndexService {
	return &IndexService{
		projectRepo: repo.NewProjectRepo(),
		sessionRepo: repo.NewSessionRepo(),
	}
}

func (s *IndexService) IndexProject(projectID uint) error {
	p, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return err
	}
	return indexdb.IndexProject(p.ID, p.Name, p.Path)
}

func (s *IndexService) IndexAllProjects() error {
	projects, err := s.projectRepo.List()
	if err != nil {
		return err
	}
	for _, p := range projects {
		if err := indexdb.IndexProject(p.ID, p.Name, p.Path); err != nil {
			return err
		}
	}
	return nil
}

func (s *IndexService) SearchSessionsByProjectName(query string) ([]SessionWithProject, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}

	projectIDs, err := indexdb.SearchProjectIDs(q)
	if err != nil {
		return nil, err
	}

	if len(projectIDs) == 0 {
		projects, perr := s.projectRepo.List()
		if perr != nil {
			return nil, perr
		}
		for _, p := range projects {
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(q)) {
				projectIDs = append(projectIDs, p.ID)
			}
		}
	}

	if len(projectIDs) == 0 {
		return nil, nil
	}

	sessions, err := s.sessionRepo.ListByProjectIDs(projectIDs)
	if err != nil {
		return nil, err
	}

	projects, err := s.projectRepo.ListByIDs(projectIDs)
	if err != nil {
		return nil, err
	}
	projectNameByID := make(map[uint]string, len(projects))
	for _, p := range projects {
		projectNameByID[p.ID] = p.Name
	}

	out := make([]SessionWithProject, 0, len(sessions))
	for _, sess := range sessions {
		pname := projectNameByID[sess.ProjectID]
		if pname == "" {
			pname = fmt.Sprintf("项目#%d", sess.ProjectID)
		}
		out = append(out, SessionWithProject{
			Session:      sess,
			ProjectName:  pname,
		})
	}
	return out, nil
}

