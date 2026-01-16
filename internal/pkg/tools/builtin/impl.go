package builtin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// --- File Operations ---

func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	return string(content), nil
}

func WriteFile(path string, content string) (string, error) {
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
	cmd := exec.Command(command, args...)
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
