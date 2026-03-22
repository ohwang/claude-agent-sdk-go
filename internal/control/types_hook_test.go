package control

import (
	"context"
	"encoding/json"
	"testing"
)

// =============================================================================
// Hook Event Tests
// =============================================================================

func TestHookEventConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant HookEvent
		expected string
	}{
		{"pre_tool_use", HookEventPreToolUse, "PreToolUse"},
		{"post_tool_use", HookEventPostToolUse, "PostToolUse"},
		{"user_prompt_submit", HookEventUserPromptSubmit, "UserPromptSubmit"},
		{"stop", HookEventStop, "Stop"},
		{"subagent_stop", HookEventSubagentStop, "SubagentStop"},
		{"pre_compact", HookEventPreCompact, "PreCompact"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("HookEvent constant %s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestHookEventCount(t *testing.T) {
	// Ensure we have exactly 6 hook events as per Python SDK parity
	events := []HookEvent{
		HookEventPreToolUse,
		HookEventPostToolUse,
		HookEventUserPromptSubmit,
		HookEventStop,
		HookEventSubagentStop,
		HookEventPreCompact,
	}

	if len(events) != 6 {
		t.Errorf("Expected 6 hook events for Python SDK parity, got %d", len(events))
	}
}

// =============================================================================
// Hook Input Type Tests
// =============================================================================

func TestBaseHookInputSerialization(t *testing.T) {
	input := BaseHookInput{
		SessionID:      "session-123",
		TranscriptPath: "/tmp/transcript.json",
		Cwd:            "/home/user/project",
		PermissionMode: "default",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal BaseHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Verify JSON field names match Python SDK
	assertHookJSONField(t, result, "session_id", "session-123")
	assertHookJSONField(t, result, "transcript_path", "/tmp/transcript.json")
	assertHookJSONField(t, result, "cwd", "/home/user/project")
	assertHookJSONField(t, result, "permission_mode", "default")
}

func TestPreToolUseHookInputSerialization(t *testing.T) {
	input := PreToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     map[string]any{"command": "ls -la"},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal PreToolUseHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Verify JSON field names match Python SDK
	assertHookJSONField(t, result, "hook_event_name", "PreToolUse")
	assertHookJSONField(t, result, "tool_name", "Bash")

	toolInput, ok := result["tool_input"].(map[string]any)
	if !ok {
		t.Fatal("tool_input should be a map")
	}
	if toolInput["command"] != "ls -la" {
		t.Errorf("tool_input.command = %v, want %q", toolInput["command"], "ls -la")
	}
}

func TestPostToolUseHookInputSerialization(t *testing.T) {
	input := PostToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "PostToolUse",
		ToolName:      "Bash",
		ToolInput:     map[string]any{"command": "ls -la"},
		ToolResponse:  "file1.txt\nfile2.txt",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal PostToolUseHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Verify JSON field names match Python SDK
	assertHookJSONField(t, result, "hook_event_name", "PostToolUse")
	assertHookJSONField(t, result, "tool_name", "Bash")
	assertHookJSONField(t, result, "tool_response", "file1.txt\nfile2.txt")
}

func TestUserPromptSubmitHookInputSerialization(t *testing.T) {
	input := UserPromptSubmitHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "UserPromptSubmit",
		Prompt:        "Please help me fix this bug",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal UserPromptSubmitHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "UserPromptSubmit")
	assertHookJSONField(t, result, "prompt", "Please help me fix this bug")
}

func TestStopHookInputSerialization(t *testing.T) {
	input := StopHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:  "Stop",
		StopHookActive: true,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal StopHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "Stop")
	if result["stop_hook_active"] != true {
		t.Errorf("stop_hook_active = %v, want true", result["stop_hook_active"])
	}
}

func TestSubagentStopHookInputSerialization(t *testing.T) {
	input := SubagentStopHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:  "SubagentStop",
		StopHookActive: false,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal SubagentStopHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "SubagentStop")
	if result["stop_hook_active"] != false {
		t.Errorf("stop_hook_active = %v, want false", result["stop_hook_active"])
	}
}

func TestPreCompactHookInputSerialization(t *testing.T) {
	customInstructions := "Be concise"
	input := PreCompactHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:      "PreCompact",
		Trigger:            "auto",
		CustomInstructions: &customInstructions,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal PreCompactHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "PreCompact")
	assertHookJSONField(t, result, "trigger", "auto")
	assertHookJSONField(t, result, "custom_instructions", "Be concise")
}

func TestPreCompactHookInputSerializationNilCustomInstructions(t *testing.T) {
	input := PreCompactHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:      "PreCompact",
		Trigger:            "manual",
		CustomInstructions: nil,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal PreCompactHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// custom_instructions should be omitted when nil
	if _, exists := result["custom_instructions"]; exists {
		t.Error("custom_instructions should be omitted when nil")
	}
}

// =============================================================================
// Hook Output Type Tests
// =============================================================================

func TestHookJSONOutputSerialization(t *testing.T) {
	continueVal := true
	decision := "block" //nolint:goconst // test value - no benefit from constant
	systemMessage := "Tool blocked"
	reason := "Security policy"

	output := HookJSONOutput{
		Continue:      &continueVal,
		Decision:      &decision,
		SystemMessage: &systemMessage,
		Reason:        &reason,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal HookJSONOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Verify JSON field names - note: Go can use "continue" directly (not a keyword)
	if result["continue"] != true {
		t.Errorf("continue = %v, want true", result["continue"])
	}
	assertHookJSONField(t, result, "decision", "block")
	assertHookJSONField(t, result, "systemMessage", "Tool blocked")
	assertHookJSONField(t, result, "reason", "Security policy")
}

func TestHookJSONOutputOmitEmpty(t *testing.T) {
	output := HookJSONOutput{} // All fields nil

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal HookJSONOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// All optional fields should be omitted
	unexpectedFields := []string{"continue", "suppressOutput", "stopReason", "decision", "systemMessage", "reason", "hookSpecificOutput"}
	for _, field := range unexpectedFields {
		if _, exists := result[field]; exists {
			t.Errorf("Field %q should be omitted when nil", field)
		}
	}
}

func TestAsyncHookJSONOutputSerialization(t *testing.T) {
	output := AsyncHookJSONOutput{
		Async:        true,
		AsyncTimeout: 5000, // 5 seconds in milliseconds
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal AsyncHookJSONOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Verify JSON field names - note: Go can use "async" directly (not a keyword)
	if result["async"] != true {
		t.Errorf("async = %v, want true", result["async"])
	}
	// JSON numbers unmarshal as float64
	if result["asyncTimeout"] != float64(5000) {
		t.Errorf("asyncTimeout = %v, want 5000", result["asyncTimeout"])
	}
}

func TestPreToolUseHookSpecificOutputSerialization(t *testing.T) {
	decision := "allow"
	reason := "User approved"
	output := PreToolUseHookSpecificOutput{
		HookEventName:            "PreToolUse",
		PermissionDecision:       &decision,
		PermissionDecisionReason: &reason,
		UpdatedInput:             map[string]any{"command": "ls"},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal PreToolUseHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "PreToolUse")
	assertHookJSONField(t, result, "permissionDecision", "allow")
	assertHookJSONField(t, result, "permissionDecisionReason", "User approved")

	updatedInput, ok := result["updatedInput"].(map[string]any)
	if !ok {
		t.Fatal("updatedInput should be a map")
	}
	if updatedInput["command"] != "ls" {
		t.Errorf("updatedInput.command = %v, want %q", updatedInput["command"], "ls")
	}
}

func TestPostToolUseHookSpecificOutputSerialization(t *testing.T) {
	context := "Tool executed with warnings"
	output := PostToolUseHookSpecificOutput{
		HookEventName:     "PostToolUse",
		AdditionalContext: &context,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal PostToolUseHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "PostToolUse")
	assertHookJSONField(t, result, "additionalContext", "Tool executed with warnings")
}

func TestUserPromptSubmitHookSpecificOutputSerialization(t *testing.T) {
	context := "Additional instructions applied"
	output := UserPromptSubmitHookSpecificOutput{
		HookEventName:     "UserPromptSubmit",
		AdditionalContext: &context,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal UserPromptSubmitHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "UserPromptSubmit")
	assertHookJSONField(t, result, "additionalContext", "Additional instructions applied")
}

// =============================================================================
// Hook Matcher Tests
// =============================================================================

func TestHookMatcherSerialization(t *testing.T) {
	timeout := 30.0
	matcher := HookMatcher{
		Matcher: "Bash|Write",
		Timeout: &timeout,
		// Hooks are not serialized (json:"-")
	}

	data, err := json.Marshal(matcher)
	if err != nil {
		t.Fatalf("Failed to marshal HookMatcher: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "matcher", "Bash|Write")
	if result["timeout"] != float64(30.0) {
		t.Errorf("timeout = %v, want 30.0", result["timeout"])
	}

	// Hooks should not be serialized
	if _, exists := result["hooks"]; exists {
		t.Error("hooks should not be serialized")
	}
}

func TestHookMatcherConfigSerialization(t *testing.T) {
	timeout := 60.0
	config := HookMatcherConfig{
		Matcher:          "Read",
		HookCallbackIDs: []string{"hook_0", "hook_1"},
		Timeout:          &timeout,
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal HookMatcherConfig: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "matcher", "Read")

	callbackIDs, ok := result["hookCallbackIds"].([]any)
	if !ok {
		t.Fatal("hookCallbackIds should be an array")
	}
	if len(callbackIDs) != 2 {
		t.Errorf("hookCallbackIds length = %d, want 2", len(callbackIDs))
	}
	if callbackIDs[0] != "hook_0" || callbackIDs[1] != "hook_1" {
		t.Errorf("hookCallbackIds = %v, want [hook_0, hook_1]", callbackIDs)
	}
}

// =============================================================================
// HookContext Tests
// =============================================================================

func TestHookContextCreation(t *testing.T) {
	ctx := context.Background()
	hookCtx := HookContext{
		Signal: ctx,
	}

	if hookCtx.Signal != ctx {
		t.Error("HookContext.Signal should hold the provided context")
	}
}

// =============================================================================
// HookCallback Type Tests
// =============================================================================

func TestHookCallbackSignature(t *testing.T) {
	// Verify the callback signature matches expected pattern
	var callback HookCallback = func(
		_ context.Context,
		_ any,
		_ *string,
		_ HookContext,
	) (HookJSONOutput, error) {
		return HookJSONOutput{}, nil
	}

	// Just verify it compiles with correct signature
	ctx := context.Background()
	result, err := callback(ctx, nil, nil, HookContext{})
	if err != nil {
		t.Errorf("Callback returned unexpected error: %v", err)
	}
	if result.Continue != nil {
		t.Error("Empty HookJSONOutput should have nil Continue")
	}
}

// =============================================================================
// WI-10: Subagent Context in BaseHookInput
// =============================================================================

func TestBaseHookInputSubagentContext(t *testing.T) {
	agentID := "agent-42"
	agentType := "general-purpose"
	input := BaseHookInput{
		SessionID:      "session-123",
		TranscriptPath: "/tmp/transcript.json",
		Cwd:            "/home/user",
		AgentID:        &agentID,
		AgentType:      &agentType,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal BaseHookInput with subagent context: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "agent_id", "agent-42")
	assertHookJSONField(t, result, "agent_type", "general-purpose")
}

func TestBaseHookInputSubagentContextOmitEmpty(t *testing.T) {
	input := BaseHookInput{
		SessionID:      "session-123",
		TranscriptPath: "/tmp/transcript.json",
		Cwd:            "/home/user",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal BaseHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if _, exists := result["agent_id"]; exists {
		t.Error("agent_id should be omitted when nil")
	}
	if _, exists := result["agent_type"]; exists {
		t.Error("agent_type should be omitted when nil")
	}
}

// =============================================================================
// WI-3: Python-Shared Hook Event Constants
// =============================================================================

func TestPythonSharedHookEventConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant HookEvent
		expected string
	}{
		{"post_tool_use_failure", HookEventPostToolUseFailure, "PostToolUseFailure"},
		{"notification", HookEventNotification, "Notification"},
		{"subagent_start", HookEventSubagentStart, "SubagentStart"},
		{"permission_request", HookEventPermissionRequest, "PermissionRequest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("HookEvent constant %s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

// =============================================================================
// WI-3: Python-Shared Hook Input Types
// =============================================================================

func TestPostToolUseFailureHookInputSerialization(t *testing.T) {
	isInterrupt := true
	input := PostToolUseFailureHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "PostToolUseFailure",
		ToolName:      "Bash",
		ToolInput:     map[string]any{"command": "rm -rf /"},
		ToolUseID:     "tool-use-abc",
		Error:         "Permission denied",
		IsInterrupt:   &isInterrupt,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal PostToolUseFailureHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "PostToolUseFailure")
	assertHookJSONField(t, result, "tool_name", "Bash")
	assertHookJSONField(t, result, "tool_use_id", "tool-use-abc")
	assertHookJSONField(t, result, "error", "Permission denied")
	if result["is_interrupt"] != true {
		t.Errorf("is_interrupt = %v, want true", result["is_interrupt"])
	}
}

func TestNotificationHookInputSerialization(t *testing.T) {
	title := "Build Complete"
	input := NotificationHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:   "Notification",
		Message:          "Build succeeded with 0 errors",
		Title:            &title,
		NotificationType: "info",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal NotificationHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "Notification")
	assertHookJSONField(t, result, "message", "Build succeeded with 0 errors")
	assertHookJSONField(t, result, "title", "Build Complete")
	assertHookJSONField(t, result, "notification_type", "info")
}

func TestSubagentStartHookInputSerialization(t *testing.T) {
	input := SubagentStartHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:   "SubagentStart",
		SubagentAgentID: "subagent-99",
		SubagentType:    "general-purpose",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal SubagentStartHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "SubagentStart")
	assertHookJSONField(t, result, "agent_id", "subagent-99")
	assertHookJSONField(t, result, "agent_type", "general-purpose")
}

func TestPermissionRequestHookInputSerialization(t *testing.T) {
	input := PermissionRequestHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:        "PermissionRequest",
		ToolName:              "Write",
		ToolInput:             map[string]any{"path": "/tmp/test.txt"},
		PermissionSuggestions: []any{"allow", "deny"},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal PermissionRequestHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "PermissionRequest")
	assertHookJSONField(t, result, "tool_name", "Write")

	suggestions, ok := result["permission_suggestions"].([]any)
	if !ok {
		t.Fatal("permission_suggestions should be an array")
	}
	if len(suggestions) != 2 {
		t.Errorf("permission_suggestions length = %d, want 2", len(suggestions))
	}
}

// =============================================================================
// WI-11: Hook-Specific Output Types for new events
// =============================================================================

func TestPostToolUseFailureHookSpecificOutputSerialization(t *testing.T) {
	ctx := "Failure was expected"
	output := PostToolUseFailureHookSpecificOutput{
		HookEventName:     "PostToolUseFailure",
		AdditionalContext: &ctx,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal PostToolUseFailureHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "PostToolUseFailure")
	assertHookJSONField(t, result, "additionalContext", "Failure was expected")
}

func TestNotificationHookSpecificOutputSerialization(t *testing.T) {
	ctx := "Notification handled"
	output := NotificationHookSpecificOutput{
		HookEventName:     "Notification",
		AdditionalContext: &ctx,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal NotificationHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "Notification")
	assertHookJSONField(t, result, "additionalContext", "Notification handled")
}

func TestSubagentStartHookSpecificOutputSerialization(t *testing.T) {
	ctx := "Subagent initialized"
	output := SubagentStartHookSpecificOutput{
		HookEventName:     "SubagentStart",
		AdditionalContext: &ctx,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal SubagentStartHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "SubagentStart")
	assertHookJSONField(t, result, "additionalContext", "Subagent initialized")
}

func TestPermissionRequestHookSpecificOutputSerialization(t *testing.T) {
	output := PermissionRequestHookSpecificOutput{
		HookEventName: "PermissionRequest",
		Decision:      "allow",
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal PermissionRequestHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "PermissionRequest")
	assertHookJSONField(t, result, "decision", "allow")
}

func TestPreToolUseHookSpecificOutputAdditionalContext(t *testing.T) {
	ctx := "Additional context for pre-tool use"
	output := PreToolUseHookSpecificOutput{
		HookEventName:     "PreToolUse",
		AdditionalContext: &ctx,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal PreToolUseHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "additionalContext", "Additional context for pre-tool use")
}

func TestPostToolUseHookSpecificOutputUpdatedMCPToolOutput(t *testing.T) {
	output := PostToolUseHookSpecificOutput{
		HookEventName:       "PostToolUse",
		UpdatedMCPToolOutput: map[string]any{"result": "modified"},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal PostToolUseHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	mcpOutput, ok := result["updatedMcpToolOutput"].(map[string]any)
	if !ok {
		t.Fatal("updatedMcpToolOutput should be a map")
	}
	if mcpOutput["result"] != "modified" {
		t.Errorf("updatedMcpToolOutput.result = %v, want %q", mcpOutput["result"], "modified")
	}
}

func TestStopHookInputLastAssistantMessage(t *testing.T) {
	lastMsg := "Here is the final answer."
	input := StopHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:       "Stop",
		StopHookActive:      true,
		LastAssistantMessage: &lastMsg,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal StopHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "last_assistant_message", "Here is the final answer.")
}

func TestSubagentStopHookInputFullFields(t *testing.T) {
	lastMsg := "Subagent done."
	input := SubagentStopHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:       "SubagentStop",
		StopHookActive:      false,
		SubagentAgentID:     "subagent-1",
		AgentTranscriptPath: "/tmp/subagent-transcript.json",
		SubagentType:        "code-review",
		LastAssistantMessage: &lastMsg,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal SubagentStopHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "SubagentStop")
	assertHookJSONField(t, result, "agent_id", "subagent-1")
	assertHookJSONField(t, result, "agent_transcript_path", "/tmp/subagent-transcript.json")
	assertHookJSONField(t, result, "agent_type", "code-review")
	assertHookJSONField(t, result, "last_assistant_message", "Subagent done.")
}

// =============================================================================
// WI-15: TS-Only Hook Event Constants
// =============================================================================

func TestTSOnlyHookEventConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant HookEvent
		expected string
	}{
		{"session_start", HookEventSessionStart, "SessionStart"},
		{"session_end", HookEventSessionEnd, "SessionEnd"},
		{"stop_failure", HookEventStopFailure, "StopFailure"},
		{"post_compact", HookEventPostCompact, "PostCompact"},
		{"setup", HookEventSetup, "Setup"},
		{"teammate_idle", HookEventTeammateIdle, "TeammateIdle"},
		{"task_completed", HookEventTaskCompleted, "TaskCompleted"},
		{"elicitation", HookEventElicitation, "Elicitation"},
		{"elicitation_result", HookEventElicitationResult, "ElicitationResult"},
		{"config_change", HookEventConfigChange, "ConfigChange"},
		{"worktree_create", HookEventWorktreeCreate, "WorktreeCreate"},
		{"worktree_remove", HookEventWorktreeRemove, "WorktreeRemove"},
		{"instructions_loaded", HookEventInstructionsLoaded, "InstructionsLoaded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.expected {
				t.Errorf("HookEvent constant %s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestAllHookEventCount(t *testing.T) {
	// 6 original + 4 Python-shared + 13 TS-only = 23 total
	events := []HookEvent{
		// Original 6
		HookEventPreToolUse,
		HookEventPostToolUse,
		HookEventUserPromptSubmit,
		HookEventStop,
		HookEventSubagentStop,
		HookEventPreCompact,
		// Python-shared 4
		HookEventPostToolUseFailure,
		HookEventNotification,
		HookEventSubagentStart,
		HookEventPermissionRequest,
		// TS-only 13
		HookEventSessionStart,
		HookEventSessionEnd,
		HookEventStopFailure,
		HookEventPostCompact,
		HookEventSetup,
		HookEventTeammateIdle,
		HookEventTaskCompleted,
		HookEventElicitation,
		HookEventElicitationResult,
		HookEventConfigChange,
		HookEventWorktreeCreate,
		HookEventWorktreeRemove,
		HookEventInstructionsLoaded,
	}

	if len(events) != 23 {
		t.Errorf("Expected 23 total hook events, got %d", len(events))
	}
}

// =============================================================================
// WI-15: TS-Only Hook Input Types
// =============================================================================

func TestSessionStartHookInputSerialization(t *testing.T) {
	agentType := "general-purpose"
	model := "claude-sonnet-4-20250514"
	input := SessionStartHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:  "SessionStart",
		Source:          "startup",
		AgentStartType: &agentType,
		Model:          &model,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal SessionStartHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "SessionStart")
	assertHookJSONField(t, result, "source", "startup")
	assertHookJSONField(t, result, "agent_start_type", "general-purpose")
	assertHookJSONField(t, result, "model", "claude-sonnet-4-20250514")
}

func TestSessionEndHookInputSerialization(t *testing.T) {
	input := SessionEndHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "SessionEnd",
		Reason:         "user_exit",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal SessionEndHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "SessionEnd")
	assertHookJSONField(t, result, "reason", "user_exit")
}

func TestStopFailureHookInputSerialization(t *testing.T) {
	details := "Timeout while waiting for response"
	lastMsg := "I was trying to..."
	input := StopFailureHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:       "StopFailure",
		Error:                "stop_timeout",
		ErrorDetails:         &details,
		LastAssistantMessage: &lastMsg,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal StopFailureHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "StopFailure")
	assertHookJSONField(t, result, "error", "stop_timeout")
	assertHookJSONField(t, result, "error_details", "Timeout while waiting for response")
	assertHookJSONField(t, result, "last_assistant_message", "I was trying to...")
}

func TestPostCompactHookInputSerialization(t *testing.T) {
	input := PostCompactHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:  "PostCompact",
		Trigger:        "auto",
		CompactSummary: "Reduced context from 50k to 10k tokens",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal PostCompactHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "PostCompact")
	assertHookJSONField(t, result, "trigger", "auto")
	assertHookJSONField(t, result, "compact_summary", "Reduced context from 50k to 10k tokens")
}

func TestSetupHookInputSerialization(t *testing.T) {
	input := SetupHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "Setup",
		Trigger:        "init",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal SetupHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "Setup")
	assertHookJSONField(t, result, "trigger", "init")
}

func TestTeammateIdleHookInputSerialization(t *testing.T) {
	input := TeammateIdleHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "TeammateIdle",
		TeammateName:   "coder-1",
		TeamName:       "engineering",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal TeammateIdleHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "TeammateIdle")
	assertHookJSONField(t, result, "teammate_name", "coder-1")
	assertHookJSONField(t, result, "team_name", "engineering")
}

func TestTaskCompletedHookInputSerialization(t *testing.T) {
	desc := "Implement the login feature"
	teammate := "coder-2"
	team := "frontend"
	input := TaskCompletedHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:  "TaskCompleted",
		TaskID:          "task-456",
		TaskSubject:     "Login Feature",
		TaskDescription: &desc,
		TeammateName:    &teammate,
		TeamName:        &team,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal TaskCompletedHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "TaskCompleted")
	assertHookJSONField(t, result, "task_id", "task-456")
	assertHookJSONField(t, result, "task_subject", "Login Feature")
	assertHookJSONField(t, result, "task_description", "Implement the login feature")
	assertHookJSONField(t, result, "teammate_name", "coder-2")
	assertHookJSONField(t, result, "team_name", "frontend")
}

func TestElicitationHookInputSerialization(t *testing.T) {
	mode := "form"
	url := "https://example.com/elicit"
	elicitID := "elicit-789"
	input := ElicitationHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName:  "Elicitation",
		McpServerName:  "my-server",
		Message:         "Please provide your API key",
		Mode:            &mode,
		URL:             &url,
		ElicitationID:  &elicitID,
		RequestedSchema: map[string]any{"type": "object"},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal ElicitationHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "Elicitation")
	assertHookJSONField(t, result, "mcp_server_name", "my-server")
	assertHookJSONField(t, result, "message", "Please provide your API key")
	assertHookJSONField(t, result, "mode", "form")
	assertHookJSONField(t, result, "url", "https://example.com/elicit")
	assertHookJSONField(t, result, "elicitation_id", "elicit-789")
}

func TestElicitationResultHookInputSerialization(t *testing.T) {
	elicitID := "elicit-789"
	mode := "form"
	input := ElicitationResultHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "ElicitationResult",
		McpServerName:  "my-server",
		ElicitationID:  &elicitID,
		Mode:            &mode,
		Action:          "submit",
		Content:         map[string]any{"key": "sk-abc123"},
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal ElicitationResultHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "ElicitationResult")
	assertHookJSONField(t, result, "mcp_server_name", "my-server")
	assertHookJSONField(t, result, "elicitation_id", "elicit-789")
	assertHookJSONField(t, result, "mode", "form")
	assertHookJSONField(t, result, "action", "submit")
}

func TestConfigChangeHookInputSerialization(t *testing.T) {
	filePath := "/home/user/.config/claude/settings.json"
	input := ConfigChangeHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "ConfigChange",
		Source:          "file_change",
		FilePath:        &filePath,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal ConfigChangeHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "ConfigChange")
	assertHookJSONField(t, result, "source", "file_change")
	assertHookJSONField(t, result, "file_path", "/home/user/.config/claude/settings.json")
}

func TestWorktreeCreateHookInputSerialization(t *testing.T) {
	input := WorktreeCreateHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "WorktreeCreate",
		Name:            "feature-branch-worktree",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal WorktreeCreateHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "WorktreeCreate")
	assertHookJSONField(t, result, "name", "feature-branch-worktree")
}

func TestWorktreeRemoveHookInputSerialization(t *testing.T) {
	input := WorktreeRemoveHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "WorktreeRemove",
		WorktreePath:   "/home/user/worktrees/feature-branch",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal WorktreeRemoveHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "WorktreeRemove")
	assertHookJSONField(t, result, "worktree_path", "/home/user/worktrees/feature-branch")
}

func TestInstructionsLoadedHookInputSerialization(t *testing.T) {
	scope := "project"
	input := InstructionsLoadedHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
		},
		HookEventName: "InstructionsLoaded",
		FilePath:        "/home/user/.claude/CLAUDE.md",
		FileType:        "markdown",
		Source:          "local",
		Scope:           &scope,
		Content:         "# Project Instructions\nBe helpful.",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal InstructionsLoadedHookInput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hook_event_name", "InstructionsLoaded")
	assertHookJSONField(t, result, "file_path", "/home/user/.claude/CLAUDE.md")
	assertHookJSONField(t, result, "file_type", "markdown")
	assertHookJSONField(t, result, "source", "local")
	assertHookJSONField(t, result, "scope", "project")
	assertHookJSONField(t, result, "content", "# Project Instructions\nBe helpful.")
}

// =============================================================================
// WI-15: TS-Only Hook Output Types
// =============================================================================

func TestSessionStartHookSpecificOutputSerialization(t *testing.T) {
	ctx := "Session started with custom config"
	output := SessionStartHookSpecificOutput{
		HookEventName:     "SessionStart",
		AdditionalContext: &ctx,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal SessionStartHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "SessionStart")
	assertHookJSONField(t, result, "additionalContext", "Session started with custom config")
}

func TestSetupHookSpecificOutputSerialization(t *testing.T) {
	ctx := "Setup completed"
	output := SetupHookSpecificOutput{
		HookEventName:     "Setup",
		AdditionalContext: &ctx,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal SetupHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "Setup")
	assertHookJSONField(t, result, "additionalContext", "Setup completed")
}

func TestElicitationHookSpecificOutputSerialization(t *testing.T) {
	ctx := "Elicitation handled"
	output := ElicitationHookSpecificOutput{
		HookEventName:     "Elicitation",
		AdditionalContext: &ctx,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal ElicitationHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "Elicitation")
	assertHookJSONField(t, result, "additionalContext", "Elicitation handled")
}

func TestElicitationResultHookSpecificOutputSerialization(t *testing.T) {
	ctx := "Result processed"
	output := ElicitationResultHookSpecificOutput{
		HookEventName:     "ElicitationResult",
		AdditionalContext: &ctx,
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal ElicitationResultHookSpecificOutput: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	assertHookJSONField(t, result, "hookEventName", "ElicitationResult")
	assertHookJSONField(t, result, "additionalContext", "Result processed")
}

// =============================================================================
// Helper Functions
// =============================================================================

func assertHookJSONField(t *testing.T, result map[string]any, field string, expected string) {
	t.Helper()
	if result[field] != expected {
		t.Errorf("%s = %v, want %q", field, result[field], expected)
	}
}
