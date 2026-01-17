package builtin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// --- File Operations ---

func ResolvePathInBase(baseDir string, userPath string) (string, error) {
	if strings.TrimSpace(userPath) == "" {
		return "", fmt.Errorf("path is required")
	}

	if strings.TrimSpace(baseDir) == "" {
		p, err := filepath.Abs(filepath.Clean(userPath))
		if err != nil {
			return "", err
		}
		return p, nil
	}

	baseAbs, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return "", err
	}

	candidate := userPath
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(baseAbs, candidate)
	}
	candidateAbs, err := filepath.Abs(filepath.Clean(candidate))
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(baseAbs, candidateAbs)
	if err != nil {
		return "", fmt.Errorf("path escapes base directory")
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes base directory")
	}
	return candidateAbs, nil
}

func ReadFile(path string) (string, error) {
	// Check file size first
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %v", err)
	}

	const maxSize = 100 * 1024 // 100KB
	if info.Size() > maxSize {
		return "", fmt.Errorf("file too large (%d bytes), please use read_file_range", info.Size())
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	return string(content), nil
}

func WriteFile(path string, content string) (string, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}
	return "success", nil
}

func ListFiles(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to list files: %v", err)
	}

	ignoredDirs := map[string]bool{
		".git":         true,
		".idea":        true,
		".vscode":      true,
		"node_modules": true,
		"dist":         true,
		"build":        true,
		"vendor":       true,
		"__pycache__":  true,
		"target":       true,
		"bin":          true,
		"obj":          true,
	}

	var files []string
	count := 0
	maxCount := 100 // Reduced from 500 to save tokens and avoid context overflow

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() && ignoredDirs[name] {
			continue
		}

		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}

		if count >= maxCount {
			files = append(files, fmt.Sprintf("... (truncated, total > %d entries)", maxCount))
			break
		}

		info, _ := entry.Info()
		prefix := "F"
		sizeStr := ""
		if entry.IsDir() {
			prefix = "D"
		} else {
			// Simplify size display
			s := info.Size()
			if s < 1024 {
				sizeStr = fmt.Sprintf(" %dB", s)
			} else if s < 1024*1024 {
				sizeStr = fmt.Sprintf(" %.1fKB", float64(s)/1024)
			} else {
				sizeStr = fmt.Sprintf(" %.1fMB", float64(s)/(1024*1024))
			}
		}
		files = append(files, fmt.Sprintf("[%s] %s%s", prefix, name, sizeStr))
		count++
	}

	if len(files) == 0 {
		return "(empty directory)", nil
	}

	return strings.Join(files, "\n"), nil
}

func ReadFileRange(path string, startLine, limit int) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	totalLines := len(lines)

	if startLine < 1 {
		startLine = 1
	}
	if startLine > totalLines {
		return "", fmt.Errorf("start line %d exceeds total lines %d", startLine, totalLines)
	}

	startIndex := startLine - 1
	endIndex := startIndex + limit

	if endIndex > totalLines {
		endIndex = totalLines
	}

	selectedLines := lines[startIndex:endIndex]

	// Add line numbers
	var result []string
	for i, line := range selectedLines {
		result = append(result, fmt.Sprintf("%d: %s", startIndex+i+1, line))
	}

	return strings.Join(result, "\n"), nil
}

func DiffFile(path1, path2 string) (string, error) {
	content1, err := os.ReadFile(path1)
	if err != nil {
		return "", fmt.Errorf("failed to read file 1: %v", err)
	}

	content2, err := os.ReadFile(path2)
	if err != nil {
		return "", fmt.Errorf("failed to read file 2: %v", err)
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(content1), string(content2), false)

	// Format diff
	return dmp.DiffPrettyText(diffs), nil
}

// --- Command Execution ---

func RunCommand(command string, args []string) (string, error) {
	if strings.TrimSpace(command) == "" {
		return "", fmt.Errorf("command is required")
	}

	useShell := false
	if len(args) == 0 {
		if strings.ContainsAny(command, "&|;<>\n\r\t") || strings.Contains(command, "&&") || strings.Contains(command, "||") {
			useShell = true
		} else {
			for _, r := range command {
				if r == ' ' {
					useShell = true
					break
				}
			}
		}
	}

	var cmd *exec.Cmd
	if useShell {
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", command)
		} else {
			cmd = exec.Command("sh", "-lc", command)
		}
	} else {
		cmd = exec.Command(command, args...)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

func RunScript(scriptPath string, args []string) (string, error) {
	// Determine the interpreter based on file extension
	var cmd *exec.Cmd
	if strings.HasSuffix(scriptPath, ".py") {
		cmd = exec.Command("python", append([]string{scriptPath}, args...)...)
	} else if strings.HasSuffix(scriptPath, ".js") {
		cmd = exec.Command("node", append([]string{scriptPath}, args...)...)
	} else if strings.HasSuffix(scriptPath, ".sh") {
		cmd = exec.Command("bash", append([]string{scriptPath}, args...)...)
	} else if strings.HasSuffix(scriptPath, ".go") {
		cmd = exec.Command("go", append([]string{"run", scriptPath}, args...)...)
	} else {
		return "", fmt.Errorf("unsupported script type: %s", scriptPath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("script execution failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

// --- Network Requests ---

func HttpGet(url string) (string, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("http get failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %v", err)
	}
	return string(body), nil
}

func HttpPost(url string, contentType string, body string) (string, error) {
	client := http.Client{Timeout: 10 * time.Second}
	if contentType == "" {
		contentType = "application/json"
	}
	resp, err := client.Post(url, contentType, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("http post failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %v", err)
	}
	return string(respBody), nil
}
