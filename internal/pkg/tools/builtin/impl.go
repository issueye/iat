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
	var files []string
	for _, entry := range entries {
		info, _ := entry.Info()
		prefix := "F"
		if entry.IsDir() {
			prefix = "D"
		}
		files = append(files, fmt.Sprintf("[%s] %s (%d bytes)", prefix, entry.Name(), info.Size()))
	}
	return strings.Join(files, "\n"), nil
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
