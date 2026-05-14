package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexEncodeValue(t *testing.T) {
	data := map[string]interface{}{
		"mcp_servers": map[string]interface{}{
			"context7": map[string]interface{}{
				"command": "npx",
				"args":    []interface{}{"-y", "@upstash/context7-mcp"},
				"env":     map[string]interface{}{"MY_VAR": "my_value"},
			},
			"my-http": map[string]interface{}{
				"transport": "streamable_http",
				"url":       "http://localhost:3001/mcp",
			},
		},
	}

	encoded, err := codexEncodeValue(data)
	if err != nil {
		t.Fatalf("codexEncodeValue failed: %v", err)
	}

	output := string(encoded)

	// Verify TOML structure: [mcp_servers.NAME] sections
	if !strings.Contains(output, "[mcp_servers.context7]") {
		t.Error("Missing [mcp_servers.context7] section")
	}
	if !strings.Contains(output, "[mcp_servers.my-http]") {
		t.Error("Missing [mcp_servers.my-http] section")
	}

	// Verify values
	if !strings.Contains(output, `command = "npx"`) {
		t.Error("Missing command = \"npx\"")
	}
	if !strings.Contains(output, `transport = "streamable_http"`) {
		t.Error("Missing transport = \"streamable_http\"")
	}
	if !strings.Contains(output, "MY_VAR") {
		t.Error("Missing env var MY_VAR")
	}
}

func TestGetCodexConfigPath(t *testing.T) {
	path := getCodexConfigPath("/home/user/myproject")
	expected := "/home/user/myproject/.codex/config.toml"
	if path != expected {
		t.Errorf("getCodexConfigPath(/home/user/myproject) = %q, want %q", path, expected)
	}

	old := os.Getenv("CODEX_HOME")
	os.Setenv("CODEX_HOME", "/custom/codex")
	defer os.Setenv("CODEX_HOME", old)

	path = getCodexConfigPath("/home/user/myproject")
	expected = "/custom/codex/.codex/config.toml"
	if path != expected {
		t.Errorf("getCodexConfigPath with CODEX_HOME = %q, want %q", path, expected)
	}
}

func TestReadFileOrEmpty(t *testing.T) {
	content, err := readFileOrEmpty("/nonexistent/file.txt")
	if err != nil {
		t.Fatalf("readFileOrEmpty nonexistent: %v", err)
	}
	if content != "" {
		t.Errorf("readFileOrEmpty nonexistent = %q, want empty", content)
	}

	tmpDir, err := os.MkdirTemp("", "readfile-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world\n"), 0644); err != nil {
		t.Fatal(err)
	}

	content, err = readFileOrEmpty(testFile)
	if err != nil {
		t.Fatalf("readFileOrEmpty existing: %v", err)
	}
	if content != "hello world\n" {
		t.Errorf("readFileOrEmpty = %q, want %q", content, "hello world\n")
	}
}

func TestWriteCodexConfig_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "codex-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("AGENTDECK_MANAGE_MCP_JSON", "true")
	defer os.Setenv("AGENTDECK_MANAGE_MCP_JSON", "")

	err = WriteCodexConfig(tmpDir, []string{})
	if err != nil {
		t.Fatalf("WriteCodexConfig failed: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".codex", "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "# BEGIN AGENTDECK CODEX MCP") {
		t.Error("Missing begin marker")
	}
	if !strings.Contains(output, "# END AGENTDECK CODEX MCP") {
		t.Error("Missing end marker")
	}
}

func TestWriteCodexConfig_PreservesExisting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "codex-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, ".codex", "config.toml")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	existing := `[defaults]
model = "claude-sonnet-4-5-20260508"
theme = "dark"
`
	if err := os.WriteFile(configPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("AGENTDECK_MANAGE_MCP_JSON", "true")
	defer os.Setenv("AGENTDECK_MANAGE_MCP_JSON", "")

	err = WriteCodexConfig(tmpDir, []string{})
	if err != nil {
		t.Fatalf("WriteCodexConfig failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, `model = "claude-sonnet-4-5-20260508"`) {
		t.Error("Existing model setting was lost")
	}
	if !strings.Contains(output, `theme = "dark"`) {
		t.Error("Existing theme setting was lost")
	}
}

func TestGetCodexMCPInfo(t *testing.T) {
	// Create config with MCP entries between markers
	tmpDir, err := os.MkdirTemp("", "codex-mcpinfo-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, ".codex", "config.toml")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	content := `[defaults]
model = "claude-sonnet-4-5-20260508"

# BEGIN AGENTDECK CODEX MCP
[mcp_servers.context7]
command = "npx"
args = ["-y", "@upstash/context7-mcp"]

[mcp_servers.my-http]
transport = "streamable_http"
url = "http://localhost:3001/mcp"
# END AGENTDECK CODEX MCP

[other]
key = "value"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info := GetCodexMCPInfo(tmpDir)
	if info == nil {
		t.Fatal("GetCodexMCPInfo returned nil")
	}
	if len(info.LocalMCPs) != 2 {
		t.Fatalf("LocalMCPs count = %d, want 2", len(info.LocalMCPs))
	}
	if info.LocalMCPs[0].Name != "context7" || info.LocalMCPs[1].Name != "my-http" {
		t.Errorf("LocalMCPs = %+v, want [context7, my-http]", info.LocalMCPs)
	}
}

func TestWriteCodexConfig_UpdateMarkers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "codex-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, ".codex", "config.toml")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	existing := `[defaults]
model = "claude-sonnet-4-5-20260508"

# BEGIN AGENTDECK CODEX MCP
[mcp_servers.old_mcp]
command = "npx"
args = ["old", "args"]
# END AGENTDECK CODEX MCP

[other]
key = "value"
`
	if err := os.WriteFile(configPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("AGENTDECK_MANAGE_MCP_JSON", "true")
	defer os.Setenv("AGENTDECK_MANAGE_MCP_JSON", "")

	err = WriteCodexConfig(tmpDir, []string{})
	if err != nil {
		t.Fatalf("WriteCodexConfig failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	output := string(data)

	if strings.Contains(output, "[mcp_servers.old_mcp]") {
		t.Error("Old MCP block was not replaced")
	}

	beginCount := strings.Count(output, "# BEGIN AGENTDECK CODEX MCP")
	endCount := strings.Count(output, "# END AGENTDECK CODEX MCP")
	if beginCount != 1 {
		t.Errorf("Begin marker count = %d, want 1", beginCount)
	}
	if endCount != 1 {
		t.Errorf("End marker count = %d, want 1", endCount)
	}

	if !strings.Contains(output, `model = "claude-sonnet-4-5-20260508"`) {
		t.Error("model setting lost")
	}
	if !strings.Contains(output, `key = "value"`) {
		t.Error("other.key lost")
	}
}
