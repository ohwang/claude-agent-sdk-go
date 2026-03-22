// Package control hook callback handling and registration.
// This file contains hook lifecycle event processing for PreToolUse, PostToolUse, etc.
package control

import (
	"context"
	"encoding/json"
	"fmt"
)

// handleHookCallbackRequest processes a hook callback request from CLI.
// Follows the same pattern as handleCanUseToolRequest with panic recovery.
func (p *Protocol) handleHookCallbackRequest(ctx context.Context, requestID string, request map[string]any) error {
	// Parse callback ID
	callbackID, _ := request["callback_id"].(string)
	if callbackID == "" {
		return p.sendErrorResponse(ctx, requestID, "missing callback_id")
	}

	// Parse hook event name from input
	inputData, _ := request["input"].(map[string]any)
	if inputData == nil {
		inputData = make(map[string]any)
	}

	eventName, _ := inputData["hook_event_name"].(string)
	event := HookEvent(eventName)

	// Parse input based on event type
	input := p.parseHookInput(event, inputData)

	// Parse tool_use_id if present
	var toolUseID *string
	if id, ok := request["tool_use_id"].(string); ok {
		toolUseID = &id
	}

	// Get callback (thread-safe read)
	p.hookCallbacksMu.RLock()
	callback, exists := p.hookCallbacks[callbackID]
	p.hookCallbacksMu.RUnlock()

	if !exists {
		return p.sendErrorResponse(ctx, requestID, fmt.Sprintf("callback not found: %s", callbackID))
	}

	// Create hook context
	hookCtx := HookContext{Signal: ctx}

	// Invoke callback with panic recovery (matches permission callback pattern)
	var result HookJSONOutput
	var callbackErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				callbackErr = fmt.Errorf("hook callback panicked: %v", r)
			}
		}()
		result, callbackErr = callback(ctx, input, toolUseID, hookCtx)
	}()

	if callbackErr != nil {
		return p.sendErrorResponse(ctx, requestID, fmt.Sprintf("callback error: %v", callbackErr))
	}

	return p.sendHookResponse(ctx, requestID, result)
}

// parseHookInput creates the appropriate typed input based on event type.
// Returns the strongly-typed input struct for the callback.
func (p *Protocol) parseHookInput(event HookEvent, inputData map[string]any) any {
	// Parse base fields (WI-10: includes AgentID and AgentType for subagent context)
	base := BaseHookInput{
		SessionID:      getString(inputData, "session_id"),
		TranscriptPath: getString(inputData, "transcript_path"),
		Cwd:            getString(inputData, "cwd"),
		PermissionMode: getString(inputData, "permission_mode"),
		AgentID:        getStringPtr(inputData, "agent_id"),
		AgentType:      getStringPtr(inputData, "agent_type"),
	}

	switch event {
	case HookEventPreToolUse:
		return &PreToolUseHookInput{
			BaseHookInput: base,
			HookEventName: "PreToolUse",
			ToolName:      getString(inputData, "tool_name"),
			ToolInput:     getMap(inputData, "tool_input"),
		}
	case HookEventPostToolUse:
		return &PostToolUseHookInput{
			BaseHookInput: base,
			HookEventName: "PostToolUse",
			ToolName:      getString(inputData, "tool_name"),
			ToolInput:     getMap(inputData, "tool_input"),
			ToolResponse:  inputData["tool_response"],
		}
	case HookEventUserPromptSubmit:
		return &UserPromptSubmitHookInput{
			BaseHookInput: base,
			HookEventName: "UserPromptSubmit",
			Prompt:        getString(inputData, "prompt"),
		}
	case HookEventStop:
		return &StopHookInput{
			BaseHookInput:        base,
			HookEventName:        "Stop",
			StopHookActive:       getBool(inputData, "stop_hook_active"),
			LastAssistantMessage: getStringPtr(inputData, "last_assistant_message"),
		}
	case HookEventSubagentStop:
		return &SubagentStopHookInput{
			BaseHookInput:        base,
			HookEventName:        "SubagentStop",
			StopHookActive:       getBool(inputData, "stop_hook_active"),
			SubagentAgentID:      getString(inputData, "agent_id"),
			AgentTranscriptPath:  getString(inputData, "agent_transcript_path"),
			SubagentType:         getString(inputData, "agent_type"),
			LastAssistantMessage: getStringPtr(inputData, "last_assistant_message"),
		}
	case HookEventPreCompact:
		return &PreCompactHookInput{
			BaseHookInput:      base,
			HookEventName:      "PreCompact",
			Trigger:            getString(inputData, "trigger"),
			CustomInstructions: getStringPtr(inputData, "custom_instructions"),
		}

	// Python-shared hook events (WI-3)
	case HookEventPostToolUseFailure:
		return &PostToolUseFailureHookInput{
			BaseHookInput: base,
			HookEventName: "PostToolUseFailure",
			ToolName:      getString(inputData, "tool_name"),
			ToolInput:     getMap(inputData, "tool_input"),
			ToolUseID:     getString(inputData, "tool_use_id"),
			Error:         getString(inputData, "error"),
			IsInterrupt:   getBoolPtr(inputData, "is_interrupt"),
		}
	case HookEventNotification:
		return &NotificationHookInput{
			BaseHookInput:    base,
			HookEventName:   "Notification",
			Message:          getString(inputData, "message"),
			Title:            getStringPtr(inputData, "title"),
			NotificationType: getString(inputData, "notification_type"),
		}
	case HookEventSubagentStart:
		return &SubagentStartHookInput{
			BaseHookInput:   base,
			HookEventName:   "SubagentStart",
			SubagentAgentID: getString(inputData, "agent_id"),
			SubagentType:    getString(inputData, "agent_type"),
		}
	case HookEventPermissionRequest:
		return &PermissionRequestHookInput{
			BaseHookInput:         base,
			HookEventName:         "PermissionRequest",
			ToolName:              getString(inputData, "tool_name"),
			ToolInput:             inputData["tool_input"],
			PermissionSuggestions: getSlice(inputData, "permission_suggestions"),
		}

	// TS-only hook events (WI-15)
	case HookEventSessionStart:
		return &SessionStartHookInput{
			BaseHookInput:  base,
			HookEventName:  "SessionStart",
			Source:          getString(inputData, "source"),
			AgentStartType: getStringPtr(inputData, "agent_start_type"),
			Model:           getStringPtr(inputData, "model"),
		}
	case HookEventSessionEnd:
		return &SessionEndHookInput{
			BaseHookInput: base,
			HookEventName: "SessionEnd",
			Reason:         getString(inputData, "reason"),
		}
	case HookEventStopFailure:
		return &StopFailureHookInput{
			BaseHookInput:        base,
			HookEventName:        "StopFailure",
			Error:                getString(inputData, "error"),
			ErrorDetails:         getStringPtr(inputData, "error_details"),
			LastAssistantMessage: getStringPtr(inputData, "last_assistant_message"),
		}
	case HookEventPostCompact:
		return &PostCompactHookInput{
			BaseHookInput:  base,
			HookEventName:  "PostCompact",
			Trigger:         getString(inputData, "trigger"),
			CompactSummary:  getString(inputData, "compact_summary"),
		}
	case HookEventSetup:
		return &SetupHookInput{
			BaseHookInput: base,
			HookEventName: "Setup",
			Trigger:        getString(inputData, "trigger"),
		}
	case HookEventTeammateIdle:
		return &TeammateIdleHookInput{
			BaseHookInput: base,
			HookEventName: "TeammateIdle",
			TeammateName:   getString(inputData, "teammate_name"),
			TeamName:       getString(inputData, "team_name"),
		}
	case HookEventTaskCompleted:
		return &TaskCompletedHookInput{
			BaseHookInput:   base,
			HookEventName:   "TaskCompleted",
			TaskID:           getString(inputData, "task_id"),
			TaskSubject:     getString(inputData, "task_subject"),
			TaskDescription: getStringPtr(inputData, "task_description"),
			TeammateName:    getStringPtr(inputData, "teammate_name"),
			TeamName:        getStringPtr(inputData, "team_name"),
		}
	case HookEventElicitation:
		return &ElicitationHookInput{
			BaseHookInput:   base,
			HookEventName:   "Elicitation",
			McpServerName:   getString(inputData, "mcp_server_name"),
			Message:          getString(inputData, "message"),
			Mode:             getStringPtr(inputData, "mode"),
			URL:              getStringPtr(inputData, "url"),
			ElicitationID:   getStringPtr(inputData, "elicitation_id"),
			RequestedSchema: getMap(inputData, "requested_schema"),
		}
	case HookEventElicitationResult:
		return &ElicitationResultHookInput{
			BaseHookInput: base,
			HookEventName: "ElicitationResult",
			McpServerName: getString(inputData, "mcp_server_name"),
			ElicitationID: getStringPtr(inputData, "elicitation_id"),
			Mode:           getStringPtr(inputData, "mode"),
			Action:         getString(inputData, "action"),
			Content:        getMap(inputData, "content"),
		}
	case HookEventConfigChange:
		return &ConfigChangeHookInput{
			BaseHookInput: base,
			HookEventName: "ConfigChange",
			Source:          getString(inputData, "source"),
			FilePath:        getStringPtr(inputData, "file_path"),
		}
	case HookEventWorktreeCreate:
		return &WorktreeCreateHookInput{
			BaseHookInput: base,
			HookEventName: "WorktreeCreate",
			Name:            getString(inputData, "name"),
		}
	case HookEventWorktreeRemove:
		return &WorktreeRemoveHookInput{
			BaseHookInput: base,
			HookEventName: "WorktreeRemove",
			WorktreePath:   getString(inputData, "worktree_path"),
		}
	case HookEventInstructionsLoaded:
		return &InstructionsLoadedHookInput{
			BaseHookInput: base,
			HookEventName: "InstructionsLoaded",
			FilePath:        getString(inputData, "file_path"),
			FileType:        getString(inputData, "file_type"),
			Source:          getString(inputData, "source"),
			Scope:           getStringPtr(inputData, "scope"),
			Content:         getString(inputData, "content"),
		}
	default:
		// Forward compatibility - return raw input for unknown events
		return inputData
	}
}

// sendHookResponse sends a hook callback response back to CLI.
func (p *Protocol) sendHookResponse(ctx context.Context, requestID string, result HookJSONOutput) error {
	// Build response data from HookJSONOutput
	responseData := make(map[string]any)

	if result.Continue != nil {
		responseData["continue"] = *result.Continue
	}
	if result.SuppressOutput != nil {
		responseData["suppressOutput"] = *result.SuppressOutput
	}
	if result.StopReason != nil {
		responseData["stopReason"] = *result.StopReason
	}
	if result.Decision != nil {
		responseData["decision"] = *result.Decision
	}
	if result.SystemMessage != nil {
		responseData["systemMessage"] = *result.SystemMessage
	}
	if result.Reason != nil {
		responseData["reason"] = *result.Reason
	}
	if result.HookSpecificOutput != nil {
		responseData["hookSpecificOutput"] = result.HookSpecificOutput
	}

	response := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: Response{
			Subtype:   ResponseSubtypeSuccess,
			RequestID: requestID,
			Response:  responseData,
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal hook response: %w", err)
	}

	return p.transport.Write(ctx, append(data, '\n'))
}

// generateHookRegistrations creates hook registrations for initialization.
// This builds the hooks config to send to CLI during initialize.
func (p *Protocol) generateHookRegistrations() []HookRegistration {
	var registrations []HookRegistration

	if p.hooks == nil {
		return registrations
	}

	// Initialize callback map if needed
	p.hookCallbacksMu.Lock()
	if p.hookCallbacks == nil {
		p.hookCallbacks = make(map[string]HookCallback)
	}

	for _, matchers := range p.hooks {
		for _, matcher := range matchers {
			for _, callback := range matcher.Hooks {
				// Generate callback ID matching Python SDK format
				callbackID := fmt.Sprintf("hook_%d", p.nextHookCallback)
				p.nextHookCallback++

				// Store callback for later lookup
				p.hookCallbacks[callbackID] = callback

				registrations = append(registrations, HookRegistration{
					CallbackID: callbackID,
					Matcher:    matcher.Matcher,
					Timeout:    matcher.Timeout,
				})
			}
		}
	}
	p.hookCallbacksMu.Unlock()

	return registrations
}

// buildHooksConfig creates the hooks config for the initialize request.
// Format: {"PreToolUse": [{"matcher": "Bash", "hookCallbackIds": ["hook_0"]}], ...}
// This matches the Python SDK's format exactly for CLI compatibility.
func (p *Protocol) buildHooksConfig() map[string][]HookMatcherConfig {
	if p.hooks == nil {
		return nil
	}

	config := make(map[string][]HookMatcherConfig)

	// Initialize callback map if needed
	p.hookCallbacksMu.Lock()
	if p.hookCallbacks == nil {
		p.hookCallbacks = make(map[string]HookCallback)
	}

	for event, matchers := range p.hooks {
		eventName := string(event)
		var matcherConfigs []HookMatcherConfig

		for _, matcher := range matchers {
			// Generate callback IDs for each callback in this matcher
			var callbackIDs []string
			for _, callback := range matcher.Hooks {
				callbackID := fmt.Sprintf("hook_%d", p.nextHookCallback)
				p.nextHookCallback++

				// Store callback for later lookup
				p.hookCallbacks[callbackID] = callback
				callbackIDs = append(callbackIDs, callbackID)
			}

			matcherConfigs = append(matcherConfigs, HookMatcherConfig{
				Matcher:         matcher.Matcher,
				HookCallbackIDs: callbackIDs,
				Timeout:         matcher.Timeout,
			})
		}

		if len(matcherConfigs) > 0 {
			config[eventName] = matcherConfigs
		}
	}
	p.hookCallbacksMu.Unlock()

	return config
}

// Helper functions for parsing hook input fields

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringPtr(m map[string]any, key string) *string {
	if v, ok := m[key].(string); ok {
		return &v
	}
	return nil
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getBoolPtr(m map[string]any, key string) *bool {
	if v, ok := m[key].(bool); ok {
		return &v
	}
	return nil
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	return make(map[string]any)
}

func getSlice(m map[string]any, key string) []any {
	if v, ok := m[key].([]any); ok {
		return v
	}
	return nil
}
