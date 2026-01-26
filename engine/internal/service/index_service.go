package service

import (
	"fmt"
	"iat/common/model"
	"iat/engine/pkg/indexdb"
	"iat/engine/internal/repo"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type SessionWithProject struct {
	model.Session
	ProjectName string `json:"projectName"`
}

type IndexService struct {
	projectRepo   *repo.ProjectRepo
	sessionRepo   *repo.SessionRepo
	autoIndexOnce sync.Once
	autoIndexErr  error
}

func NewIndexService() *IndexService {
	return &IndexService{
		projectRepo: repo.NewProjectRepo(),
		sessionRepo: repo.NewSessionRepo(),
	}
}

type IndexResult struct {
	Indexed int    `json:"indexed"`
	Files   int    `json:"files"`
	DBPath  string `json:"dbPath"`
}

func (s *IndexService) IndexProject(projectID uint) (*IndexResult, error) {
	p, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, err
	}
	if err := indexdb.IndexProject(p.ID, p.Name, p.Path); err != nil {
		return nil, err
	}
	ids, _ := indexdb.SearchProjectIDs(p.Name)
	if !containsUint(ids, p.ID) {
		return nil, fmt.Errorf("索引校验失败：无法通过项目名检索到 projectId=%d", p.ID)
	}
	files, ferr := listCommittableFiles(p.Path)
	if ferr != nil {
		var werr error
		files, werr = walkProjectFiles(p.Path)
		if werr != nil {
			return nil, fmt.Errorf("无法获取项目文件列表: %v (git) / %v (walk)", ferr, werr)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("项目目录下没有找到可索引的文件")
	}
	codeInfo, cerr := indexdb.IndexProjectCodeFiles(p.ID, p.Path, files)
	if cerr != nil {
		return nil, cerr
	}
	return &IndexResult{
		Indexed: 1,
		Files:   codeInfo.Files,
		DBPath:  indexdb.OpenedPath(),
	}, nil
}

func (s *IndexService) IndexAllProjects() (*IndexResult, error) {
	projects, err := s.projectRepo.List()
	if err != nil {
		return nil, err
	}
	totalFiles := 0
	for _, p := range projects {
		if err := indexdb.IndexProject(p.ID, p.Name, p.Path); err != nil {
			return nil, err
		}
		files, ferr := listCommittableFiles(p.Path)
		if ferr != nil {
			var werr error
			files, werr = walkProjectFiles(p.Path)
			if werr != nil {
				continue
			}
		}
		if len(files) == 0 {
			continue
		}
		codeInfo, cerr := indexdb.IndexProjectCodeFiles(p.ID, p.Path, files)
		if cerr != nil {
			return nil, cerr
		}
		totalFiles += codeInfo.Files
	}
	return &IndexResult{
		Indexed: len(projects),
		Files:   totalFiles,
		DBPath:  indexdb.OpenedPath(),
	}, nil
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

	// Also search by code content
	codeProjectIDs, _ := indexdb.SearchProjectIDsByCodeContent(q)
	for _, cid := range codeProjectIDs {
		if !containsUint(projectIDs, cid) {
			projectIDs = append(projectIDs, cid)
		}
	}

	if len(projectIDs) == 0 {
		s.autoIndexOnce.Do(func() {
			_, s.autoIndexErr = s.IndexAllProjects()
		})
		if s.autoIndexErr == nil {
			projectIDs, _ = indexdb.SearchProjectIDs(q)
		}
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
			Session:     sess,
			ProjectName: pname,
		})
	}
	return out, nil
}

func containsUint(items []uint, v uint) bool {
	for _, it := range items {
		if it == v {
			return true
		}
	}
	return false
}

func listCommittableFiles(projectPath string) ([]string, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("project path is empty")
	}
	if info, err := os.Stat(projectPath); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("项目路径不存在或不是目录: %s", projectPath)
	}
	cmd := exec.Command("git", "-C", projectPath, "ls-files", "-z")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}
	parts := strings.Split(string(out), "\x00")
	var files []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		files = append(files, filepath.ToSlash(p))
	}
	return files, nil
}

func walkProjectFiles(projectPath string) ([]string, error) {
	var files []string
	root := filepath.Clean(projectPath)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if name == ".git" || name == "node_modules" || name == "dist" || name == "vendor" {
				return filepath.SkipDir
			}
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			return nil
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			return nil
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	return files, err
}
