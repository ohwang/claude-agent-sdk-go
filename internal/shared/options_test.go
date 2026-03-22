package shared

import (
	"encoding/json"
	"testing"
)

// TestOptionsDefaults tests Options struct default values using table-driven approach
func TestOptionsDefaults(t *testing.T) {
	options := NewOptions()

	tests := []struct {
		name     string
		field    interface{}
		expected interface{}
	}{
		{"MaxThinkingTokens", options.MaxThinkingTokens, 8000},
		{"ContinueConversation", options.ContinueConversation, false},
		{"MaxTurns", options.MaxTurns, 0},
		{"AllowedTools_initialized", options.AllowedTools == nil, false},
		{"AllowedTools_empty", len(options.AllowedTools), 0},
		{"DisallowedTools_initialized", options.DisallowedTools == nil, false},
		{"DisallowedTools_empty", len(options.DisallowedTools), 0},
		{"Betas_initialized", options.Betas == nil, false},
		{"Betas_empty", len(options.Betas), 0},
		{"AddDirs_initialized", options.AddDirs == nil, false},
		{"AddDirs_empty", len(options.AddDirs), 0},
		{"McpServers_initialized", options.McpServers == nil, false},
		{"McpServers_empty", len(options.McpServers), 0},
		{"ExtraArgs_initialized", options.ExtraArgs == nil, false},
		{"ExtraArgs_empty", len(options.ExtraArgs), 0},
		{"ExtraEnv_initialized", options.ExtraEnv == nil, false},
		{"ExtraEnv_empty", len(options.ExtraEnv), 0},
		{"ForkSession", options.ForkSession, false},
		{"SettingSources_initialized", options.SettingSources == nil, false},
		{"SettingSources_empty", len(options.SettingSources), 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertOptionsField(t, test.field, test.expected, test.name)
		})
	}

	// Test nil pointer fields
	nilTests := []struct {
		name  string
		check func() bool
	}{
		{"SystemPrompt", func() bool { return options.SystemPrompt == nil }},
		{"AppendSystemPrompt", func() bool { return options.AppendSystemPrompt == nil }},
		{"Model", func() bool { return options.Model == nil }},
		{"PermissionMode", func() bool { return options.PermissionMode == nil }},
		{"PermissionPromptToolName", func() bool { return options.PermissionPromptToolName == nil }},
		{"Resume", func() bool { return options.Resume == nil }},
		{"Settings", func() bool { return options.Settings == nil }},
		{"Cwd", func() bool { return options.Cwd == nil }},
		{"Thinking", func() bool { return options.Thinking == nil }},
		{"Effort", func() bool { return options.Effort == nil }},
	}

	for _, test := range nilTests {
		t.Run(test.name+"_nil", func(t *testing.T) {
			if !test.check() {
				t.Errorf("Expected %s to be nil", test.name)
			}
		})
	}
}

// TestOptionsValidation tests critical validation edge cases
func TestOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Options
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid_options",
			setup: func() *Options {
				opts := NewOptions()
				opts.AllowedTools = []string{"Read", "Write"}
				return opts
			},
			wantErr: false,
		},
		{
			name: "negative_thinking_tokens",
			setup: func() *Options {
				opts := NewOptions()
				opts.MaxThinkingTokens = -100
				return opts
			},
			wantErr: true,
			errMsg:  "MaxThinkingTokens must be non-negative, got -100",
		},
		{
			name: "conflicting_tools",
			setup: func() *Options {
				opts := NewOptions()
				opts.AllowedTools = []string{"Read", "Write"}
				opts.DisallowedTools = []string{"Write", "Bash"}
				return opts
			},
			wantErr: true,
			errMsg:  "tool 'Write' cannot be in both AllowedTools and DisallowedTools",
		},
		{
			name: "negative_max_turns",
			setup: func() *Options {
				opts := NewOptions()
				opts.MaxTurns = -5
				return opts
			},
			wantErr: true,
			errMsg:  "MaxTurns must be non-negative, got -5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := test.setup()
			err := options.Validate()
			assertValidationError(t, err, test.wantErr, test.errMsg)
		})
	}
}

// TestMcpServerTypes tests MCP server configuration interface compliance
func TestMcpServerTypes(t *testing.T) {
	tests := []struct {
		name         string
		config       McpServerConfig
		expectedType McpServerType
	}{
		{
			name: "stdio_server",
			config: &McpStdioServerConfig{
				Type:    McpServerTypeStdio,
				Command: "node",
				Args:    []string{"server.js"},
			},
			expectedType: McpServerTypeStdio,
		},
		{
			name: "sse_server",
			config: &McpSSEServerConfig{
				Type: McpServerTypeSSE,
				URL:  "https://example.com/sse",
			},
			expectedType: McpServerTypeSSE,
		},
		{
			name: "http_server",
			config: &McpHTTPServerConfig{
				Type: McpServerTypeHTTP,
				URL:  "https://example.com/http",
			},
			expectedType: McpServerTypeHTTP,
		},
		{
			name: "claudeai_proxy_server",
			config: &McpClaudeAIProxyServerConfig{
				Type: McpServerTypeClaudeAIProxy,
				URL:  "https://proxy.claude.ai",
				ID:   "proxy-123",
			},
			expectedType: McpServerTypeClaudeAIProxy,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertMcpServerType(t, test.config, test.expectedType)
		})
	}
}

// TestPermissionModeConstants tests permission mode constant values
func TestPermissionModeConstants(t *testing.T) {
	tests := []struct {
		mode     PermissionMode
		expected string
	}{
		{PermissionModeDefault, "default"},
		{PermissionModeAcceptEdits, "acceptEdits"},
		{PermissionModePlan, "plan"},
		{PermissionModeBypassPermissions, "bypassPermissions"},
	}

	for _, test := range tests {
		t.Run(string(test.mode), func(t *testing.T) {
			assertOptionsField(t, string(test.mode), test.expected, "PermissionMode")
		})
	}
}

// TestSettingSourceConstants tests setting source constant values
func TestSettingSourceConstants(t *testing.T) {
	tests := []struct {
		source   SettingSource
		expected string
	}{
		{SettingSourceUser, "user"},
		{SettingSourceProject, "project"},
		{SettingSourceLocal, "local"},
	}

	for _, test := range tests {
		t.Run(string(test.source), func(t *testing.T) {
			assertOptionsField(t, string(test.source), test.expected, "SettingSource")
		})
	}
}

// Helper functions

// assertOptionsField verifies field values with proper error reporting
func assertOptionsField(t *testing.T, actual, expected interface{}, fieldName string) {
	t.Helper()
	// Handle nil pointer comparisons properly
	if expected == nil {
		if actual != nil {
			t.Errorf("Expected %s = nil, got %v", fieldName, actual)
		}
		return
	}
	if actual != expected {
		t.Errorf("Expected %s = %v, got %v", fieldName, expected, actual)
	}
}

// assertValidationError verifies validation error behavior
func assertValidationError(t *testing.T, err error, wantErr bool, expectedMsg string) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
		return
	}
	if wantErr && expectedMsg != "" && err.Error() != expectedMsg {
		t.Errorf("error = %v, expected message %q", err, expectedMsg)
	}
}

// assertMcpServerType verifies MCP server configuration types
func assertMcpServerType(t *testing.T, config McpServerConfig, expectedType McpServerType) {
	t.Helper()
	if config.GetType() != expectedType {
		t.Errorf("Expected server type %s, got %s", expectedType, config.GetType())
	}
}

// TestSandboxSettingsDefaults tests that Sandbox is nil by default
func TestSandboxSettingsDefaults(t *testing.T) {
	options := NewOptions()

	if options.Sandbox != nil {
		t.Errorf("Expected Sandbox to be nil by default, got %+v", options.Sandbox)
	}
}

// TestSandboxSettingsTypes tests sandbox type definitions and JSON serialization
func TestSandboxSettingsTypes(t *testing.T) {
	// Test that all sandbox types are properly defined
	sandbox := &SandboxSettings{
		Enabled:                   true,
		AutoAllowBashIfSandboxed:  true,
		ExcludedCommands:          []string{"docker", "git"},
		AllowUnsandboxedCommands:  false,
		EnableWeakerNestedSandbox: false,
		Network: &SandboxNetworkConfig{
			AllowUnixSockets:    []string{"/var/run/docker.sock"},
			AllowAllUnixSockets: false,
			AllowLocalBinding:   true,
		},
		IgnoreViolations: &SandboxIgnoreViolations{
			File:    []string{"/tmp/*"},
			Network: []string{"localhost"},
		},
	}

	// Verify fields are accessible and correct
	if !sandbox.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if !sandbox.AutoAllowBashIfSandboxed {
		t.Error("Expected AutoAllowBashIfSandboxed to be true")
	}
	if len(sandbox.ExcludedCommands) != 2 {
		t.Errorf("Expected 2 ExcludedCommands, got %d", len(sandbox.ExcludedCommands))
	}
	if sandbox.Network == nil {
		t.Error("Expected Network to be set")
	}
	if sandbox.Network != nil && !sandbox.Network.AllowLocalBinding {
		t.Error("Expected Network.AllowLocalBinding to be true")
	}
	if sandbox.IgnoreViolations == nil {
		t.Error("Expected IgnoreViolations to be set")
	}
}

// =============================================================================
// WI-1: Thinking Configuration & Effort
// =============================================================================

// TestThinkingConfigInterface tests that all ThinkingConfig types satisfy the interface
func TestThinkingConfigInterface(t *testing.T) {
	tests := []struct {
		name   string
		config ThinkingConfig
	}{
		{"adaptive", ThinkingAdaptive{}},
		{"enabled_no_budget", ThinkingEnabled{}},
		{"enabled_with_budget", ThinkingEnabled{BudgetTokens: intPtr(10000)}},
		{"disabled", ThinkingDisabled{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Verify it satisfies the interface (compile-time check via assignment)
			var _ ThinkingConfig = test.config
		})
	}
}

// TestThinkingEnabledBudgetTokens tests ThinkingEnabled with and without budget
func TestThinkingEnabledBudgetTokens(t *testing.T) {
	// Without budget
	te := ThinkingEnabled{}
	if te.BudgetTokens != nil {
		t.Errorf("Expected nil BudgetTokens, got %v", te.BudgetTokens)
	}

	// With budget
	budget := 5000
	teWithBudget := ThinkingEnabled{BudgetTokens: &budget}
	if teWithBudget.BudgetTokens == nil {
		t.Fatal("Expected non-nil BudgetTokens")
	}
	if *teWithBudget.BudgetTokens != 5000 {
		t.Errorf("Expected BudgetTokens = 5000, got %d", *teWithBudget.BudgetTokens)
	}
}

// TestEffortConstants tests effort level constant values
func TestEffortConstants(t *testing.T) {
	tests := []struct {
		effort   Effort
		expected string
	}{
		{EffortLow, "low"},
		{EffortMedium, "medium"},
		{EffortHigh, "high"},
		{EffortMax, "max"},
	}

	for _, test := range tests {
		t.Run(string(test.effort), func(t *testing.T) {
			if string(test.effort) != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, string(test.effort))
			}
		})
	}
}

// TestOptionsThinkingAndEffort tests that Thinking and Effort fields work on Options
func TestOptionsThinkingAndEffort(t *testing.T) {
	opts := NewOptions()

	// Set thinking
	opts.Thinking = ThinkingAdaptive{}
	if opts.Thinking == nil {
		t.Error("Expected Thinking to be set")
	}

	// Set effort
	effort := EffortHigh
	opts.Effort = &effort
	if opts.Effort == nil {
		t.Fatal("Expected Effort to be set")
	}
	if *opts.Effort != EffortHigh {
		t.Errorf("Expected Effort = %q, got %q", EffortHigh, *opts.Effort)
	}
}

// =============================================================================
// WI-7: AgentDefinition Missing Fields
// =============================================================================

// TestAgentDefinitionNewFields tests the new fields on AgentDefinition
func TestAgentDefinitionNewFields(t *testing.T) {
	maxTurns := 10
	agent := AgentDefinition{
		Description:     "Test agent",
		Prompt:          "You are a test agent.",
		Tools:           []string{"Read", "Write"},
		DisallowedTools: []string{"Bash"},
		Model:           AgentModelSonnet,
		McpServers:      []any{"server1", map[string]any{"command": "node", "args": []string{"server.js"}}},
		Skills:          []string{"skill1", "skill2"},
		MaxTurns:        &maxTurns,
	}

	if len(agent.DisallowedTools) != 1 || agent.DisallowedTools[0] != "Bash" {
		t.Errorf("Expected DisallowedTools = [Bash], got %v", agent.DisallowedTools)
	}
	if len(agent.McpServers) != 2 {
		t.Errorf("Expected 2 McpServers, got %d", len(agent.McpServers))
	}
	if len(agent.Skills) != 2 {
		t.Errorf("Expected 2 Skills, got %d", len(agent.Skills))
	}
	if agent.MaxTurns == nil || *agent.MaxTurns != 10 {
		t.Errorf("Expected MaxTurns = 10, got %v", agent.MaxTurns)
	}
}

// TestAgentDefinitionJSON tests JSON serialization of AgentDefinition with new fields
func TestAgentDefinitionJSON(t *testing.T) {
	maxTurns := 5
	agent := AgentDefinition{
		Description:     "JSON test",
		Prompt:          "Test prompt",
		DisallowedTools: []string{"Bash"},
		Skills:          []string{"coder"},
		MaxTurns:        &maxTurns,
	}

	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["disallowedTools"] == nil {
		t.Error("Expected disallowedTools in JSON")
	}
	if result["skills"] == nil {
		t.Error("Expected skills in JSON")
	}
	if result["maxTurns"] == nil {
		t.Error("Expected maxTurns in JSON")
	}
	// Omitted empty fields should not appear
	if result["tools"] != nil {
		t.Error("Expected tools to be omitted when empty")
	}
	if result["mcpServers"] != nil {
		t.Error("Expected mcpServers to be omitted when empty")
	}
}

// =============================================================================
// WI-9: MCP Tool Annotations
// =============================================================================

// TestToolAnnotations tests ToolAnnotations struct
func TestToolAnnotations(t *testing.T) {
	readOnly := true
	destructive := false
	openWorld := true

	annotations := &ToolAnnotations{
		ReadOnly:    &readOnly,
		Destructive: &destructive,
		OpenWorld:   &openWorld,
	}

	if annotations.ReadOnly == nil || !*annotations.ReadOnly {
		t.Error("Expected ReadOnly = true")
	}
	if annotations.Destructive == nil || *annotations.Destructive {
		t.Error("Expected Destructive = false")
	}
	if annotations.OpenWorld == nil || !*annotations.OpenWorld {
		t.Error("Expected OpenWorld = true")
	}
}

// TestToolAnnotationsJSON tests JSON serialization of ToolAnnotations
func TestToolAnnotationsJSON(t *testing.T) {
	readOnly := true
	annotations := &ToolAnnotations{
		ReadOnly: &readOnly,
	}

	data, err := json.Marshal(annotations)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["readOnly"] != true {
		t.Errorf("Expected readOnly = true, got %v", result["readOnly"])
	}
	// Omitted fields should not appear
	if result["destructive"] != nil {
		t.Error("Expected destructive to be omitted")
	}
	if result["openWorld"] != nil {
		t.Error("Expected openWorld to be omitted")
	}
}

// TestMcpToolDefinitionWithAnnotations tests McpToolDefinition with annotations
func TestMcpToolDefinitionWithAnnotations(t *testing.T) {
	readOnly := true
	def := McpToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]any{"type": "object"},
		Annotations: &ToolAnnotations{ReadOnly: &readOnly},
	}

	if def.Annotations == nil {
		t.Fatal("Expected Annotations to be set")
	}
	if !*def.Annotations.ReadOnly {
		t.Error("Expected ReadOnly = true")
	}

	// Test JSON serialization includes annotations
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["annotations"] == nil {
		t.Error("Expected annotations in JSON output")
	}
}

// =============================================================================
// WI-12: SystemPromptPreset
// =============================================================================

// TestSystemPromptPreset tests SystemPromptPreset struct
func TestSystemPromptPreset(t *testing.T) {
	appendText := "Be concise."
	preset := SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
		Append: &appendText,
	}

	if preset.Type != "preset" {
		t.Errorf("Expected Type = preset, got %s", preset.Type)
	}
	if preset.Preset != "claude_code" {
		t.Errorf("Expected Preset = claude_code, got %s", preset.Preset)
	}
	if preset.Append == nil || *preset.Append != "Be concise." {
		t.Errorf("Expected Append = 'Be concise.', got %v", preset.Append)
	}
}

// TestSystemPromptPresetJSON tests JSON serialization
func TestSystemPromptPresetJSON(t *testing.T) {
	// With append
	appendText := "Extra instructions"
	preset := SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
		Append: &appendText,
	}

	data, err := json.Marshal(preset)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["type"] != "preset" {
		t.Errorf("Expected type = preset, got %v", result["type"])
	}
	if result["preset"] != "claude_code" {
		t.Errorf("Expected preset = claude_code, got %v", result["preset"])
	}
	if result["append"] != "Extra instructions" {
		t.Errorf("Expected append = 'Extra instructions', got %v", result["append"])
	}

	// Without append (should omit)
	presetNoAppend := SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
	}

	data, err = json.Marshal(presetNoAppend)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result2 map[string]any
	if err := json.Unmarshal(data, &result2); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result2["append"] != nil {
		t.Error("Expected append to be omitted when nil")
	}
}

// =============================================================================
// WI-18: Sandbox Filesystem & Ripgrep
// =============================================================================

// TestSandboxFilesystemConfig tests SandboxFilesystemConfig struct
func TestSandboxFilesystemConfig(t *testing.T) {
	managedOnly := true
	fs := &SandboxFilesystemConfig{
		AllowWrite:                []string{"/tmp", "/workspace"},
		DenyWrite:                 []string{"/etc"},
		DenyRead:                  []string{"/secrets"},
		AllowRead:                 []string{"/data"},
		AllowManagedReadPathsOnly: &managedOnly,
	}

	if len(fs.AllowWrite) != 2 {
		t.Errorf("Expected 2 AllowWrite paths, got %d", len(fs.AllowWrite))
	}
	if len(fs.DenyWrite) != 1 {
		t.Errorf("Expected 1 DenyWrite path, got %d", len(fs.DenyWrite))
	}
	if len(fs.DenyRead) != 1 {
		t.Errorf("Expected 1 DenyRead path, got %d", len(fs.DenyRead))
	}
	if len(fs.AllowRead) != 1 {
		t.Errorf("Expected 1 AllowRead path, got %d", len(fs.AllowRead))
	}
	if fs.AllowManagedReadPathsOnly == nil || !*fs.AllowManagedReadPathsOnly {
		t.Error("Expected AllowManagedReadPathsOnly = true")
	}
}

// TestSandboxRipgrepConfig tests SandboxRipgrepConfig struct
func TestSandboxRipgrepConfig(t *testing.T) {
	rg := &SandboxRipgrepConfig{
		Command: "/usr/bin/rg",
		Args:    []string{"--max-depth", "5"},
	}

	if rg.Command != "/usr/bin/rg" {
		t.Errorf("Expected Command = /usr/bin/rg, got %s", rg.Command)
	}
	if len(rg.Args) != 2 {
		t.Errorf("Expected 2 Args, got %d", len(rg.Args))
	}
}

// TestSandboxSettingsWithNewFields tests SandboxSettings with filesystem and ripgrep
func TestSandboxSettingsWithNewFields(t *testing.T) {
	sandbox := &SandboxSettings{
		Enabled: true,
		Filesystem: &SandboxFilesystemConfig{
			AllowWrite: []string{"/workspace"},
		},
		Ripgrep: &SandboxRipgrepConfig{
			Command: "rg",
		},
		EnableWeakerNetworkIsolation: true,
	}

	if sandbox.Filesystem == nil {
		t.Error("Expected Filesystem to be set")
	}
	if sandbox.Ripgrep == nil {
		t.Error("Expected Ripgrep to be set")
	}
	if !sandbox.EnableWeakerNetworkIsolation {
		t.Error("Expected EnableWeakerNetworkIsolation = true")
	}
}

// TestSandboxNetworkConfigNewFields tests new fields on SandboxNetworkConfig
func TestSandboxNetworkConfigNewFields(t *testing.T) {
	managedOnly := true
	network := &SandboxNetworkConfig{
		AllowedDomains:          []string{"example.com", "api.example.com"},
		AllowManagedDomainsOnly: &managedOnly,
	}

	if len(network.AllowedDomains) != 2 {
		t.Errorf("Expected 2 AllowedDomains, got %d", len(network.AllowedDomains))
	}
	if network.AllowManagedDomainsOnly == nil || !*network.AllowManagedDomainsOnly {
		t.Error("Expected AllowManagedDomainsOnly = true")
	}
}

// TestSandboxSettingsJSONWithNewFields tests JSON serialization with new fields
func TestSandboxSettingsJSONWithNewFields(t *testing.T) {
	sandbox := &SandboxSettings{
		Enabled: true,
		Filesystem: &SandboxFilesystemConfig{
			AllowWrite: []string{"/workspace"},
		},
		Ripgrep: &SandboxRipgrepConfig{
			Command: "rg",
			Args:    []string{"--max-depth", "3"},
		},
		EnableWeakerNetworkIsolation: true,
		Network: &SandboxNetworkConfig{
			AllowedDomains: []string{"example.com"},
		},
	}

	data, err := json.Marshal(sandbox)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["filesystem"] == nil {
		t.Error("Expected filesystem in JSON")
	}
	if result["ripgrep"] == nil {
		t.Error("Expected ripgrep in JSON")
	}
	if result["enableWeakerNetworkIsolation"] != true {
		t.Error("Expected enableWeakerNetworkIsolation = true in JSON")
	}
}

// =============================================================================
// WI-21: McpClaudeAIProxy
// =============================================================================

// TestMcpClaudeAIProxyServerConfig tests the ClaudeAI proxy config type
func TestMcpClaudeAIProxyServerConfig(t *testing.T) {
	config := &McpClaudeAIProxyServerConfig{
		Type: McpServerTypeClaudeAIProxy,
		URL:  "https://proxy.claude.ai/mcp",
		ID:   "proxy-abc-123",
	}

	if config.GetType() != McpServerTypeClaudeAIProxy {
		t.Errorf("Expected type %s, got %s", McpServerTypeClaudeAIProxy, config.GetType())
	}
	if config.URL != "https://proxy.claude.ai/mcp" {
		t.Errorf("Expected URL = https://proxy.claude.ai/mcp, got %s", config.URL)
	}
	if config.ID != "proxy-abc-123" {
		t.Errorf("Expected ID = proxy-abc-123, got %s", config.ID)
	}
}

// TestMcpClaudeAIProxyServerConfigJSON tests JSON serialization
func TestMcpClaudeAIProxyServerConfigJSON(t *testing.T) {
	config := &McpClaudeAIProxyServerConfig{
		Type: McpServerTypeClaudeAIProxy,
		URL:  "https://proxy.claude.ai",
		ID:   "test-id",
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["type"] != "claudeai-proxy" {
		t.Errorf("Expected type = claudeai-proxy, got %v", result["type"])
	}
	if result["url"] != "https://proxy.claude.ai" {
		t.Errorf("Expected url = https://proxy.claude.ai, got %v", result["url"])
	}
	if result["id"] != "test-id" {
		t.Errorf("Expected id = test-id, got %v", result["id"])
	}
}

// TestMcpServerTypeClaudeAIProxyConstant tests the constant value
func TestMcpServerTypeClaudeAIProxyConstant(t *testing.T) {
	if McpServerTypeClaudeAIProxy != "claudeai-proxy" {
		t.Errorf("Expected McpServerTypeClaudeAIProxy = 'claudeai-proxy', got %q", McpServerTypeClaudeAIProxy)
	}
}

// =============================================================================
// Helpers
// =============================================================================

func intPtr(n int) *int {
	return &n
}
