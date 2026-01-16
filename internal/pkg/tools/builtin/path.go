package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolvePathInBase(baseDir string, inputPath string) (string, error) {
	if strings.TrimSpace(baseDir) == "" {
		return strings.TrimSpace(inputPath), nil
	}

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base dir: %v", err)
	}
	baseAbs = filepath.Clean(baseAbs)

	p := strings.TrimSpace(inputPath)
	if p == "" {
		p = "."
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(baseAbs, p)
	}

	pAbs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}
	pAbs = filepath.Clean(pAbs)

	rel, err := filepath.Rel(baseAbs, pAbs)
	if err != nil {
		return "", fmt.Errorf("failed to validate path: %v", err)
	}
	rel = filepath.Clean(rel)
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path outside project: %s", inputPath)
	}

	return pAbs, nil
}

