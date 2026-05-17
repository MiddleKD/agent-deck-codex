package session

import (
	"os"
	"path/filepath"
	"strings"
)

// GetCodexHomeDirForInstance resolves CODEX_HOME for an instance.
//
// Priority (most → least specific):
//  1. CODEX_HOME prepended in the instance command string — handled by callers
//  2. CODEX_HOME environment variable
//  3. [profiles.<name>.codex].config_dir
//  4. [codex].config_dir (global)
//  5. ~/.codex (default)
func GetCodexHomeDirForInstance(_ *Instance) string {
	return resolveCodexHomeDir()
}

// resolveCodexHomeDir walks the CODEX_HOME priority chain (env → profile →
// global → default). Command-prefix CODEX_HOME (highest priority) is handled
// by buildCodexCommand directly because it requires the resolved command
// string, not just the config.
func resolveCodexHomeDir() string {
	if envDir := strings.TrimSpace(os.Getenv("CODEX_HOME")); envDir != "" {
		return ExpandPath(envDir)
	}

	userConfig, _ := LoadUserConfig()
	if userConfig != nil {
		profile := GetEffectiveProfile("")
		if dir := userConfig.GetProfileCodexConfigDir(profile); dir != "" {
			return dir
		}
		if userConfig.Codex.ConfigDir != "" {
			return ExpandPath(userConfig.Codex.ConfigDir)
		}
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codex")
}
