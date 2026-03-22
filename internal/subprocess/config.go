package subprocess

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/severity1/claude-agent-sdk-go/internal/cli"
	"github.com/severity1/claude-agent-sdk-go/internal/control"
	"github.com/severity1/claude-agent-sdk-go/internal/shared"
)

// generateMcpConfigFile creates a temporary MCP config file from options.McpServers.
// Returns the file path. The file is stored in t.mcpConfigFile for cleanup.
func (t *Transport) generateMcpConfigFile() (string, error) {
	// Build servers map, stripping Instance field from SDK servers for CLI serialization
	// The CLI doesn't need the Go instance - it routes mcp_message requests to the SDK
	serversForCLI := make(map[string]any)
	for name, config := range t.options.McpServers {
		if sdkConfig, ok := config.(*shared.McpSdkServerConfig); ok {
			// SDK servers: only send type and name to CLI
			serversForCLI[name] = map[string]any{
				"type": string(sdkConfig.Type),
				"name": sdkConfig.Name,
			}
		} else {
			// External servers: pass as-is
			serversForCLI[name] = config
		}
	}

	// Create the MCP config structure matching Claude CLI expected format
	mcpConfig := map[string]interface{}{
		"mcpServers": serversForCLI,
	}

	// Marshal to JSON
	configData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal MCP config: %w", err)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "claude_mcp_config_*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write config data
	if _, err := tmpFile.Write(configData); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write MCP config: %w", err)
	}

	// Sync to ensure data is written
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to sync MCP config file: %w", err)
	}

	// Store for cleanup later
	t.mcpConfigFile = tmpFile

	return tmpFile.Name(), nil
}

// GetValidator returns the stream validator for diagnostic purposes.
// This allows clients to check for validation issues like missing tool results.
func (t *Transport) GetValidator() *shared.StreamValidator {
	return t.validator
}

// SetModel changes the AI model during a streaming session.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
func (t *Transport) SetModel(ctx context.Context, model *string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return fmt.Errorf("SetModel not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return fmt.Errorf("control protocol not initialized")
	}

	return t.protocol.SetModel(ctx, model)
}

// SetPermissionMode changes the permission mode during a streaming session.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
func (t *Transport) SetPermissionMode(ctx context.Context, mode string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return fmt.Errorf("SetPermissionMode not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return fmt.Errorf("control protocol not initialized")
	}

	return t.protocol.SetPermissionMode(ctx, mode)
}

// RewindFiles reverts tracked files to their state at a specific user message.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
// Returns error if not connected, not in streaming mode, or protocol not initialized.
func (t *Transport) RewindFiles(ctx context.Context, userMessageID string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return fmt.Errorf("RewindFiles not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return fmt.Errorf("control protocol not initialized")
	}

	return t.protocol.RewindFiles(ctx, userMessageID)
}

// GetMcpStatus returns the status of all configured MCP servers.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
func (t *Transport) GetMcpStatus(ctx context.Context) ([]shared.McpServerStatusEntry, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return nil, fmt.Errorf("GetMcpStatus not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return nil, fmt.Errorf("control protocol not initialized")
	}

	result, err := t.protocol.GetMcpStatus(ctx)
	if err != nil {
		return nil, err
	}

	// Parse the raw response into typed status entries
	return parseMcpStatusResponse(result)
}

// parseMcpStatusResponse converts the raw control protocol response to typed entries.
func parseMcpStatusResponse(result any) ([]shared.McpServerStatusEntry, error) {
	resultSlice, ok := result.([]any)
	if !ok {
		// Try as a map with a "servers" key
		if resultMap, ok := result.(map[string]any); ok {
			if servers, ok := resultMap["servers"]; ok {
				resultSlice, ok = servers.([]any)
				if !ok {
					return nil, fmt.Errorf("unexpected mcp_status response format")
				}
			}
		}
		if resultSlice == nil {
			return nil, fmt.Errorf("unexpected mcp_status response format")
		}
	}

	entries := make([]shared.McpServerStatusEntry, 0, len(resultSlice))
	for _, item := range resultSlice {
		entryMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		entry := shared.McpServerStatusEntry{}
		if name, ok := entryMap["name"].(string); ok {
			entry.Name = name
		}
		if status, ok := entryMap["status"].(string); ok {
			entry.Status = shared.McpServerConnectionStatus(status)
		}
		if errMsg, ok := entryMap["error"].(string); ok {
			entry.Error = &errMsg
		}
		if scope, ok := entryMap["scope"].(string); ok {
			entry.Scope = &scope
		}
		if config, ok := entryMap["config"].(map[string]any); ok {
			entry.Config = config
		}
		if serverInfo, ok := entryMap["serverInfo"].(map[string]any); ok {
			si := &shared.McpServerInfo{}
			if name, ok := serverInfo["name"].(string); ok {
				si.Name = name
			}
			if version, ok := serverInfo["version"].(string); ok {
				si.Version = version
			}
			entry.ServerInfo = si
		}
		if tools, ok := entryMap["tools"].([]any); ok {
			for _, toolItem := range tools {
				toolMap, ok := toolItem.(map[string]any)
				if !ok {
					continue
				}
				ti := shared.McpToolInfo{}
				if name, ok := toolMap["name"].(string); ok {
					ti.Name = name
				}
				if desc, ok := toolMap["description"].(string); ok {
					ti.Description = &desc
				}
				entry.Tools = append(entry.Tools, ti)
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// ReconnectMcpServer reconnects a disconnected MCP server by name.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
func (t *Transport) ReconnectMcpServer(ctx context.Context, serverName string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return fmt.Errorf("ReconnectMcpServer not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return fmt.Errorf("control protocol not initialized")
	}

	return t.protocol.ReconnectMcpServer(ctx, serverName)
}

// ToggleMcpServer enables or disables an MCP server by name.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
func (t *Transport) ToggleMcpServer(ctx context.Context, serverName string, enabled bool) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return fmt.Errorf("ToggleMcpServer not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return fmt.Errorf("control protocol not initialized")
	}

	return t.protocol.ToggleMcpServer(ctx, serverName, enabled)
}

// SetMcpServers dynamically replaces the set of MCP servers.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
func (t *Transport) SetMcpServers(ctx context.Context, servers map[string]any) (map[string]any, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return nil, fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return nil, fmt.Errorf("SetMcpServers not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return nil, fmt.Errorf("control protocol not initialized")
	}

	result, err := t.protocol.SetMcpServers(ctx, servers)
	if err != nil {
		return nil, err
	}

	if resultMap, ok := result.(map[string]any); ok {
		return resultMap, nil
	}

	return nil, nil
}

// StopTask stops a running background task by its ID.
// This method requires control protocol integration which is only available
// in streaming mode (when closeStdin is false).
func (t *Transport) StopTask(ctx context.Context, taskID string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	// Control protocol integration is only available in streaming mode
	if t.closeStdin {
		return fmt.Errorf("StopTask not available in one-shot mode")
	}

	// Delegate to control protocol
	if t.protocol == nil {
		return fmt.Errorf("control protocol not initialized")
	}

	return t.protocol.StopTask(ctx, taskID)
}

// buildProtocolOptions constructs control protocol options from transport configuration.
// This extracts callback wiring logic from Connect to reduce cyclomatic complexity.
func (t *Transport) buildProtocolOptions() []control.ProtocolOption {
	var opts []control.ProtocolOption

	// Wire permission callback if configured
	if t.options != nil && t.options.CanUseTool != nil {
		// Create adapter that converts between shared.Options (any types)
		// and control package (strongly-typed) to avoid import cycles
		optionsCallback := t.options.CanUseTool
		opts = append(opts,
			control.WithCanUseToolCallback(func(
				ctx context.Context,
				toolName string,
				input map[string]any,
				permCtx control.ToolPermissionContext,
			) (control.PermissionResult, error) {
				// Call the Options callback with any-typed permCtx
				result, err := optionsCallback(ctx, toolName, input, permCtx)
				if err != nil {
					return nil, err
				}

				// Convert result back to strongly-typed PermissionResult
				if pr, ok := result.(control.PermissionResult); ok {
					return pr, nil
				}

				// Fallback: deny if result type is unexpected
				return control.NewPermissionResultDeny("invalid permission result type"), nil
			}))
	}

	// Wire hooks if configured
	if t.options != nil && t.options.Hooks != nil {
		// Convert from any to strongly-typed hooks map
		if hooks, ok := t.options.Hooks.(map[control.HookEvent][]control.HookMatcher); ok {
			opts = append(opts, control.WithHooks(hooks))
		}
	}

	// Wire SDK MCP servers to protocol (Issue #7)
	if t.options != nil && len(t.options.McpServers) > 0 {
		sdkServers := make(map[string]control.McpServer)
		for name, config := range t.options.McpServers {
			if sdkConfig, ok := config.(*shared.McpSdkServerConfig); ok && sdkConfig.Instance != nil {
				sdkServers[name] = sdkConfig.Instance
			}
		}
		if len(sdkServers) > 0 {
			opts = append(opts, control.WithSdkMcpServers(sdkServers))
		}
	}

	return opts
}

// hasSdkMcpServers checks if any SDK MCP servers are configured.
// Returns true if at least one SDK server with a valid Instance exists.
func (t *Transport) hasSdkMcpServers() bool {
	if t.options == nil || len(t.options.McpServers) == 0 {
		return false
	}
	for _, config := range t.options.McpServers {
		if sdkConfig, ok := config.(*shared.McpSdkServerConfig); ok && sdkConfig.Instance != nil {
			return true
		}
	}
	return false
}

// buildEnvironment constructs the environment variables for the subprocess.
// This extracts environment setup logic from Connect to reduce cyclomatic complexity.
func (t *Transport) buildEnvironment() []string {
	env := os.Environ()

	// Set entrypoint to identify SDK to CLI
	env = append(env, "CLAUDE_CODE_ENTRYPOINT="+t.entrypoint)

	// Enable file checkpointing if requested (matches Python SDK)
	if t.options != nil && t.options.EnableFileCheckpointing {
		env = append(env, "CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING=true")
	}

	// Add user-specified environment variables
	if t.options != nil && t.options.ExtraEnv != nil {
		for key, value := range t.options.ExtraEnv {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return env
}

// prepareMcpConfig generates MCP config file if needed and returns modified options.
// Returns the original options unchanged if no MCP servers are configured.
func (t *Transport) prepareMcpConfig() (*shared.Options, error) {
	if t.options == nil || len(t.options.McpServers) == 0 {
		return t.options, nil
	}

	mcpConfigPath, err := t.generateMcpConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to generate MCP config file: %w", err)
	}

	// Create shallow copy with mcp-config in ExtraArgs
	optsCopy := *t.options
	if optsCopy.ExtraArgs == nil {
		optsCopy.ExtraArgs = make(map[string]*string)
	} else {
		extraArgsCopy := make(map[string]*string, len(optsCopy.ExtraArgs)+1)
		for k, v := range optsCopy.ExtraArgs {
			extraArgsCopy[k] = v
		}
		optsCopy.ExtraArgs = extraArgsCopy
	}
	optsCopy.ExtraArgs["mcp-config"] = &mcpConfigPath
	return &optsCopy, nil
}

// emitCLIVersionWarning performs a non-blocking CLI version check and emits
// a warning via StderrCallback if the CLI version is outdated.
func (t *Transport) emitCLIVersionWarning(ctx context.Context) {
	if warning := cli.CheckCLIVersion(ctx, t.cliPath); warning != "" {
		if t.options != nil && t.options.StderrCallback != nil {
			t.options.StderrCallback(warning)
		}
	}
}
