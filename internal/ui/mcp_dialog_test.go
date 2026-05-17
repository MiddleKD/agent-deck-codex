package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMCPDialog_TypeJumpAvailable(t *testing.T) {
	dialog := NewMCPDialog()
	dialog.visible = true
	dialog.scope = MCPScopeLocal
	dialog.column = MCPColumnAvailable
	dialog.localAvailable = []MCPItem{
		{Name: "alpha"},
		{Name: "delta"},
		{Name: "docs"},
		{Name: "zeta"},
	}
	dialog.localAvailableIdx = 0

	_, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if dialog.localAvailableIdx != 1 {
		t.Fatalf("expected jump to delta (index 1), got %d", dialog.localAvailableIdx)
	}

	_, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if dialog.localAvailableIdx != 2 {
		t.Fatalf("expected jump to docs (index 2), got %d", dialog.localAvailableIdx)
	}
}

func TestMCPDialog_TypeJumpWrapAround(t *testing.T) {
	dialog := NewMCPDialog()
	dialog.visible = true
	dialog.scope = MCPScopeLocal
	dialog.column = MCPColumnAvailable
	dialog.localAvailable = []MCPItem{
		{Name: "alpha"},
		{Name: "delta"},
		{Name: "docs"},
	}
	dialog.localAvailableIdx = 2

	_, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if dialog.localAvailableIdx != 1 {
		t.Fatalf("expected wrapped jump to delta (index 1), got %d", dialog.localAvailableIdx)
	}
}

func TestMCPDialog_TypeJumpResetOnScopeSwitch(t *testing.T) {
	dialog := NewMCPDialog()
	dialog.visible = true
	dialog.tool = "claude"
	dialog.scope = MCPScopeLocal
	dialog.column = MCPColumnAvailable
	dialog.localAvailable = []MCPItem{{Name: "docs"}}
	dialog.globalAvailable = []MCPItem{{Name: "zeta"}, {Name: "alpha"}}

	_, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if dialog.typeJumpBuf != "d" {
		t.Fatalf("expected jump buffer d, got %q", dialog.typeJumpBuf)
	}

	_, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyTab})
	if dialog.scope != MCPScopeGlobal {
		t.Fatalf("expected scope to switch to global, got %v", dialog.scope)
	}
	if dialog.typeJumpBuf != "" {
		t.Fatalf("expected jump buffer reset on scope switch, got %q", dialog.typeJumpBuf)
	}

	_, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	if dialog.globalAvailableIdx != 0 {
		t.Fatalf("expected jump in global list to zeta (index 0), got %d", dialog.globalAvailableIdx)
	}
}

// TestMCPDialog_Apply_CodexRemovesMCPFromConfigToml is a regression test for
// the bug where removing an MCP from a codex session in the MCP dialog didn't
// update .codex/config.toml (Apply only wrote .mcp.json and SkipMCPRegenerate
// prevented the regeneration step that would translate .mcp.json → .codex/config.toml).
func TestMCPDialog_Apply_CodexRemovesMCPFromConfigToml(t *testing.T) {
	oldEnv := os.Getenv("AGENTDECK_MANAGE_MCP_JSON")
	os.Setenv("AGENTDECK_MANAGE_MCP_JSON", "true")
	defer os.Setenv("AGENTDECK_MANAGE_MCP_JSON", oldEnv)

	tmpDir, err := os.MkdirTemp("", "mcp-dialog-codex-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	codexConfigDir := filepath.Join(tmpDir, ".codex")
	if err := os.MkdirAll(codexConfigDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-condition: .codex/config.toml has both mcp-a and mcp-b enabled
	initial := "[defaults]\nmodel = \"claude-sonnet-4-5\"\n\n" +
		"# BEGIN AGENTDECK CODEX MCP\n" +
		"[mcp_servers.mcp-a]\ncommand = \"npx\"\nargs = [\"-y\", \"mcp-a\"]\n\n" +
		"[mcp_servers.mcp-b]\ncommand = \"npx\"\nargs = [\"-y\", \"mcp-b\"]\n" +
		"# END AGENTDECK CODEX MCP\n"
	configPath := filepath.Join(codexConfigDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Simulate user removing mcp-b: dialog shows only mcp-a in localAttached
	dialog := NewMCPDialog()
	dialog.tool = "codex"
	dialog.projectPath = tmpDir
	dialog.scope = MCPScopeLocal
	dialog.localAttached = []MCPItem{{Name: "mcp-a"}}
	dialog.localChanged = true

	if err := dialog.Apply(); err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read .codex/config.toml after Apply: %v", err)
	}
	output := string(data)

	if strings.Contains(output, "mcp-b") {
		t.Error("mcp-b should have been removed from .codex/config.toml but is still present")
	}
	if !strings.Contains(output, "# BEGIN AGENTDECK CODEX MCP") {
		t.Error("AGENTDECK markers should still be present in .codex/config.toml")
	}
	if !strings.Contains(output, `model = "claude-sonnet-4-5"`) {
		t.Error("existing [defaults] config should be preserved in .codex/config.toml")
	}
}
