// Package control provides hook types for lifecycle event handling.
// This file implements feature parity with the Python and TypeScript SDK hooks systems.
package control

import "context"

// =============================================================================
// Hook Event Types
// =============================================================================

// HookEvent represents lifecycle events that can trigger hooks.
type HookEvent string

const (
	// HookEventPreToolUse is triggered before a tool is executed.
	HookEventPreToolUse HookEvent = "PreToolUse"
	// HookEventPostToolUse is triggered after a tool is executed.
	HookEventPostToolUse HookEvent = "PostToolUse"
	// HookEventUserPromptSubmit is triggered when a user submits a prompt.
	HookEventUserPromptSubmit HookEvent = "UserPromptSubmit"
	// HookEventStop is triggered when the session is stopping.
	HookEventStop HookEvent = "Stop"
	// HookEventSubagentStop is triggered when a subagent is stopping.
	HookEventSubagentStop HookEvent = "SubagentStop"
	// HookEventPreCompact is triggered before context compaction.
	HookEventPreCompact HookEvent = "PreCompact"

	// Python-shared hook events (WI-3)

	// HookEventPostToolUseFailure is triggered after a tool use fails.
	HookEventPostToolUseFailure HookEvent = "PostToolUseFailure"
	// HookEventNotification is triggered when a notification is emitted.
	HookEventNotification HookEvent = "Notification"
	// HookEventSubagentStart is triggered when a subagent starts.
	HookEventSubagentStart HookEvent = "SubagentStart"
	// HookEventPermissionRequest is triggered when a permission request is made.
	HookEventPermissionRequest HookEvent = "PermissionRequest"

	// TS-only hook events (WI-15)

	// HookEventSessionStart is triggered when a session starts.
	HookEventSessionStart HookEvent = "SessionStart"
	// HookEventSessionEnd is triggered when a session ends.
	HookEventSessionEnd HookEvent = "SessionEnd"
	// HookEventStopFailure is triggered when a stop attempt fails.
	HookEventStopFailure HookEvent = "StopFailure"
	// HookEventPostCompact is triggered after context compaction.
	HookEventPostCompact HookEvent = "PostCompact"
	// HookEventSetup is triggered during setup.
	HookEventSetup HookEvent = "Setup"
	// HookEventTeammateIdle is triggered when a teammate becomes idle.
	HookEventTeammateIdle HookEvent = "TeammateIdle"
	// HookEventTaskCompleted is triggered when a task is completed.
	HookEventTaskCompleted HookEvent = "TaskCompleted"
	// HookEventElicitation is triggered for an elicitation request.
	HookEventElicitation HookEvent = "Elicitation"
	// HookEventElicitationResult is triggered when an elicitation result is received.
	HookEventElicitationResult HookEvent = "ElicitationResult"
	// HookEventConfigChange is triggered when configuration changes.
	HookEventConfigChange HookEvent = "ConfigChange"
	// HookEventWorktreeCreate is triggered when a worktree is created.
	HookEventWorktreeCreate HookEvent = "WorktreeCreate"
	// HookEventWorktreeRemove is triggered when a worktree is removed.
	HookEventWorktreeRemove HookEvent = "WorktreeRemove"
	// HookEventInstructionsLoaded is triggered when instructions are loaded.
	HookEventInstructionsLoaded HookEvent = "InstructionsLoaded"
)

// =============================================================================
// Hook Input Types
// =============================================================================

// BaseHookInput contains common fields present across all hook events.
type BaseHookInput struct {
	// SessionID is the unique identifier for the session.
	SessionID string `json:"session_id"`
	// TranscriptPath is the path to the transcript file.
	TranscriptPath string `json:"transcript_path"`
	// Cwd is the current working directory.
	Cwd string `json:"cwd"`
	// PermissionMode is the current permission mode (optional).
	PermissionMode string `json:"permission_mode,omitempty"`
	// AgentID identifies the agent (nil for main thread). Present when the hook fires from within a subagent.
	AgentID *string `json:"agent_id,omitempty"`
	// AgentType distinguishes main-thread from subagent calls (nil for main thread).
	// Present in subagent context (e.g., "general-purpose").
	AgentType *string `json:"agent_type,omitempty"`
}

// PreToolUseHookInput is the input for PreToolUse hook events.
type PreToolUseHookInput struct {
	BaseHookInput
	// HookEventName is always "PreToolUse".
	HookEventName string `json:"hook_event_name"`
	// ToolName is the name of the tool being executed.
	ToolName string `json:"tool_name"`
	// ToolInput contains the tool's input parameters.
	ToolInput map[string]any `json:"tool_input"`
}

// PostToolUseHookInput is the input for PostToolUse hook events.
type PostToolUseHookInput struct {
	BaseHookInput
	// HookEventName is always "PostToolUse".
	HookEventName string `json:"hook_event_name"`
	// ToolName is the name of the tool that was executed.
	ToolName string `json:"tool_name"`
	// ToolInput contains the tool's input parameters.
	ToolInput map[string]any `json:"tool_input"`
	// ToolResponse contains the tool's output.
	ToolResponse any `json:"tool_response"`
}

// UserPromptSubmitHookInput is the input for UserPromptSubmit hook events.
type UserPromptSubmitHookInput struct {
	BaseHookInput
	// HookEventName is always "UserPromptSubmit".
	HookEventName string `json:"hook_event_name"`
	// Prompt is the user's submitted prompt.
	Prompt string `json:"prompt"`
}

// StopHookInput is the input for Stop hook events.
type StopHookInput struct {
	BaseHookInput
	// HookEventName is always "Stop".
	HookEventName string `json:"hook_event_name"`
	// StopHookActive indicates if the stop hook is currently active.
	StopHookActive bool `json:"stop_hook_active"`
	// LastAssistantMessage is the last message from the assistant (optional).
	LastAssistantMessage *string `json:"last_assistant_message,omitempty"`
}

// SubagentStopHookInput is the input for SubagentStop hook events.
type SubagentStopHookInput struct {
	BaseHookInput
	// HookEventName is always "SubagentStop".
	HookEventName string `json:"hook_event_name"`
	// StopHookActive indicates if the stop hook is currently active.
	StopHookActive bool `json:"stop_hook_active"`
	// AgentID is the identifier of the subagent.
	SubagentAgentID string `json:"agent_id"`
	// AgentTranscriptPath is the path to the subagent's transcript.
	AgentTranscriptPath string `json:"agent_transcript_path"`
	// SubagentType is the type of the subagent.
	SubagentType string `json:"agent_type"`
	// LastAssistantMessage is the last message from the assistant (optional).
	LastAssistantMessage *string `json:"last_assistant_message,omitempty"`
}

// PreCompactHookInput is the input for PreCompact hook events.
type PreCompactHookInput struct {
	BaseHookInput
	// HookEventName is always "PreCompact".
	HookEventName string `json:"hook_event_name"`
	// Trigger is either "manual" or "auto".
	Trigger string `json:"trigger"`
	// CustomInstructions contains custom compaction instructions (optional).
	CustomInstructions *string `json:"custom_instructions,omitempty"`
}

// =============================================================================
// Python-Shared Hook Input Types (WI-3)
// =============================================================================

// PostToolUseFailureHookInput is the input for PostToolUseFailure hook events.
type PostToolUseFailureHookInput struct {
	BaseHookInput
	// HookEventName is always "PostToolUseFailure".
	HookEventName string `json:"hook_event_name"`
	// ToolName is the name of the tool that failed.
	ToolName string `json:"tool_name"`
	// ToolInput contains the tool's input parameters.
	ToolInput map[string]any `json:"tool_input"`
	// ToolUseID is the unique identifier for this tool use.
	ToolUseID string `json:"tool_use_id"`
	// Error is the error message from the failed tool use.
	Error string `json:"error"`
	// IsInterrupt indicates whether the failure was due to an interrupt (optional).
	IsInterrupt *bool `json:"is_interrupt,omitempty"`
}

// NotificationHookInput is the input for Notification hook events.
type NotificationHookInput struct {
	BaseHookInput
	// HookEventName is always "Notification".
	HookEventName string `json:"hook_event_name"`
	// Message is the notification message.
	Message string `json:"message"`
	// Title is the optional notification title.
	Title *string `json:"title,omitempty"`
	// NotificationType is the type of notification.
	NotificationType string `json:"notification_type"`
}

// SubagentStartHookInput is the input for SubagentStart hook events.
type SubagentStartHookInput struct {
	BaseHookInput
	// HookEventName is always "SubagentStart".
	HookEventName string `json:"hook_event_name"`
	// SubagentAgentID is the identifier of the subagent being started.
	SubagentAgentID string `json:"agent_id"`
	// SubagentType is the type of the subagent being started.
	SubagentType string `json:"agent_type"`
}

// PermissionRequestHookInput is the input for PermissionRequest hook events.
type PermissionRequestHookInput struct {
	BaseHookInput
	// HookEventName is always "PermissionRequest".
	HookEventName string `json:"hook_event_name"`
	// ToolName is the name of the tool requesting permission.
	ToolName string `json:"tool_name"`
	// ToolInput contains the tool's input parameters.
	ToolInput any `json:"tool_input"`
	// PermissionSuggestions contains suggested permission decisions (optional).
	PermissionSuggestions []any `json:"permission_suggestions,omitempty"`
}

// =============================================================================
// TS-Only Hook Input Types (WI-15)
// =============================================================================

// SessionStartHookInput is the input for SessionStart hook events.
type SessionStartHookInput struct {
	BaseHookInput
	// HookEventName is always "SessionStart".
	HookEventName string `json:"hook_event_name"`
	// Source is the reason for session start ("startup", "resume", "clear", "compact").
	Source string `json:"source"`
	// AgentStartType is the type of agent starting the session (optional).
	AgentStartType *string `json:"agent_start_type,omitempty"`
	// Model is the model being used (optional).
	Model *string `json:"model,omitempty"`
}

// SessionEndHookInput is the input for SessionEnd hook events.
type SessionEndHookInput struct {
	BaseHookInput
	// HookEventName is always "SessionEnd".
	HookEventName string `json:"hook_event_name"`
	// Reason is why the session ended.
	Reason string `json:"reason"`
}

// StopFailureHookInput is the input for StopFailure hook events.
type StopFailureHookInput struct {
	BaseHookInput
	// HookEventName is always "StopFailure".
	HookEventName string `json:"hook_event_name"`
	// Error is the error message.
	Error string `json:"error"`
	// ErrorDetails provides additional error details (optional).
	ErrorDetails *string `json:"error_details,omitempty"`
	// LastAssistantMessage is the last message from the assistant (optional).
	LastAssistantMessage *string `json:"last_assistant_message,omitempty"`
}

// PostCompactHookInput is the input for PostCompact hook events.
type PostCompactHookInput struct {
	BaseHookInput
	// HookEventName is always "PostCompact".
	HookEventName string `json:"hook_event_name"`
	// Trigger is what triggered the compaction.
	Trigger string `json:"trigger"`
	// CompactSummary is the summary of the compaction.
	CompactSummary string `json:"compact_summary"`
}

// SetupHookInput is the input for Setup hook events.
type SetupHookInput struct {
	BaseHookInput
	// HookEventName is always "Setup".
	HookEventName string `json:"hook_event_name"`
	// Trigger is what triggered the setup ("init" or "maintenance").
	Trigger string `json:"trigger"`
}

// TeammateIdleHookInput is the input for TeammateIdle hook events.
type TeammateIdleHookInput struct {
	BaseHookInput
	// HookEventName is always "TeammateIdle".
	HookEventName string `json:"hook_event_name"`
	// TeammateName is the name of the idle teammate.
	TeammateName string `json:"teammate_name"`
	// TeamName is the name of the team.
	TeamName string `json:"team_name"`
}

// TaskCompletedHookInput is the input for TaskCompleted hook events.
type TaskCompletedHookInput struct {
	BaseHookInput
	// HookEventName is always "TaskCompleted".
	HookEventName string `json:"hook_event_name"`
	// TaskID is the identifier of the completed task.
	TaskID string `json:"task_id"`
	// TaskSubject is the subject of the completed task.
	TaskSubject string `json:"task_subject"`
	// TaskDescription is the description of the completed task (optional).
	TaskDescription *string `json:"task_description,omitempty"`
	// TeammateName is the name of the teammate that completed the task (optional).
	TeammateName *string `json:"teammate_name,omitempty"`
	// TeamName is the name of the team (optional).
	TeamName *string `json:"team_name,omitempty"`
}

// ElicitationHookInput is the input for Elicitation hook events.
type ElicitationHookInput struct {
	BaseHookInput
	// HookEventName is always "Elicitation".
	HookEventName string `json:"hook_event_name"`
	// McpServerName is the name of the MCP server.
	McpServerName string `json:"mcp_server_name"`
	// Message is the elicitation message.
	Message string `json:"message"`
	// Mode is the elicitation mode (optional).
	Mode *string `json:"mode,omitempty"`
	// URL is the elicitation URL (optional).
	URL *string `json:"url,omitempty"`
	// ElicitationID is the unique identifier for this elicitation (optional).
	ElicitationID *string `json:"elicitation_id,omitempty"`
	// RequestedSchema is the schema requested for the elicitation.
	RequestedSchema map[string]any `json:"requested_schema"`
}

// ElicitationResultHookInput is the input for ElicitationResult hook events.
type ElicitationResultHookInput struct {
	BaseHookInput
	// HookEventName is always "ElicitationResult".
	HookEventName string `json:"hook_event_name"`
	// McpServerName is the name of the MCP server.
	McpServerName string `json:"mcp_server_name"`
	// ElicitationID is the unique identifier for this elicitation (optional).
	ElicitationID *string `json:"elicitation_id,omitempty"`
	// Mode is the elicitation mode (optional).
	Mode *string `json:"mode,omitempty"`
	// Action is the action taken on the elicitation.
	Action string `json:"action"`
	// Content is the elicitation result content.
	Content map[string]any `json:"content"`
}

// ConfigChangeHookInput is the input for ConfigChange hook events.
type ConfigChangeHookInput struct {
	BaseHookInput
	// HookEventName is always "ConfigChange".
	HookEventName string `json:"hook_event_name"`
	// Source is where the config change originated.
	Source string `json:"source"`
	// FilePath is the path of the changed config file (optional).
	FilePath *string `json:"file_path,omitempty"`
}

// WorktreeCreateHookInput is the input for WorktreeCreate hook events.
type WorktreeCreateHookInput struct {
	BaseHookInput
	// HookEventName is always "WorktreeCreate".
	HookEventName string `json:"hook_event_name"`
	// Name is the name of the worktree being created.
	Name string `json:"name"`
}

// WorktreeRemoveHookInput is the input for WorktreeRemove hook events.
type WorktreeRemoveHookInput struct {
	BaseHookInput
	// HookEventName is always "WorktreeRemove".
	HookEventName string `json:"hook_event_name"`
	// WorktreePath is the path of the worktree being removed.
	WorktreePath string `json:"worktree_path"`
}

// InstructionsLoadedHookInput is the input for InstructionsLoaded hook events.
type InstructionsLoadedHookInput struct {
	BaseHookInput
	// HookEventName is always "InstructionsLoaded".
	HookEventName string `json:"hook_event_name"`
	// FilePath is the path of the instructions file.
	FilePath string `json:"file_path"`
	// FileType is the type of the instructions file.
	FileType string `json:"file_type"`
	// Source is where the instructions were loaded from.
	Source string `json:"source"`
	// Scope is the scope of the instructions (optional).
	Scope *string `json:"scope,omitempty"`
	// Content is the instructions content.
	Content string `json:"content"`
}

// =============================================================================
// Hook-Specific Output Types
// =============================================================================

// PreToolUseHookSpecificOutput contains PreToolUse-specific output fields.
type PreToolUseHookSpecificOutput struct {
	// HookEventName is always "PreToolUse".
	HookEventName string `json:"hookEventName"`
	// PermissionDecision is "allow", "deny", or "ask".
	PermissionDecision *string `json:"permissionDecision,omitempty"`
	// PermissionDecisionReason explains the decision.
	PermissionDecisionReason *string `json:"permissionDecisionReason,omitempty"`
	// UpdatedInput contains modified tool input (optional).
	UpdatedInput map[string]any `json:"updatedInput,omitempty"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// PostToolUseHookSpecificOutput contains PostToolUse-specific output fields.
type PostToolUseHookSpecificOutput struct {
	// HookEventName is always "PostToolUse".
	HookEventName string `json:"hookEventName"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
	// UpdatedMCPToolOutput contains modified MCP tool output (optional).
	UpdatedMCPToolOutput any `json:"updatedMcpToolOutput,omitempty"`
}

// UserPromptSubmitHookSpecificOutput contains UserPromptSubmit-specific output fields.
type UserPromptSubmitHookSpecificOutput struct {
	// HookEventName is always "UserPromptSubmit".
	HookEventName string `json:"hookEventName"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// PostToolUseFailureHookSpecificOutput contains PostToolUseFailure-specific output fields.
type PostToolUseFailureHookSpecificOutput struct {
	// HookEventName is always "PostToolUseFailure".
	HookEventName string `json:"hookEventName"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// NotificationHookSpecificOutput contains Notification-specific output fields.
type NotificationHookSpecificOutput struct {
	// HookEventName is always "Notification".
	HookEventName string `json:"hookEventName"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// SubagentStartHookSpecificOutput contains SubagentStart-specific output fields.
type SubagentStartHookSpecificOutput struct {
	// HookEventName is always "SubagentStart".
	HookEventName string `json:"hookEventName"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// PermissionRequestHookSpecificOutput contains PermissionRequest-specific output fields.
type PermissionRequestHookSpecificOutput struct {
	// HookEventName is always "PermissionRequest".
	HookEventName string `json:"hookEventName"`
	// Decision is the permission decision.
	Decision any `json:"decision"`
}

// SessionStartHookSpecificOutput contains SessionStart-specific output fields.
type SessionStartHookSpecificOutput struct {
	// HookEventName is always "SessionStart".
	HookEventName string `json:"hookEventName"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// SetupHookSpecificOutput contains Setup-specific output fields.
type SetupHookSpecificOutput struct {
	// HookEventName is always "Setup".
	HookEventName string `json:"hookEventName"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// ElicitationHookSpecificOutput contains Elicitation-specific output fields.
type ElicitationHookSpecificOutput struct {
	// HookEventName is always "Elicitation".
	HookEventName string `json:"hookEventName"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// ElicitationResultHookSpecificOutput contains ElicitationResult-specific output fields.
type ElicitationResultHookSpecificOutput struct {
	// HookEventName is always "ElicitationResult".
	HookEventName string `json:"hookEventName"`
	// AdditionalContext provides extra context for Claude.
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// =============================================================================
// Hook Output Types
// =============================================================================

// HookJSONOutput is the synchronous hook output structure.
type HookJSONOutput struct {
	// Continue indicates whether Claude should proceed (default: true).
	// Python SDK uses continue_ to avoid keyword conflict.
	Continue *bool `json:"continue,omitempty"`
	// SuppressOutput hides stdout from transcript mode.
	SuppressOutput *bool `json:"suppressOutput,omitempty"`
	// StopReason is the message shown when Continue is false.
	StopReason *string `json:"stopReason,omitempty"`

	// Decision can be "block" to indicate blocking behavior.
	Decision *string `json:"decision,omitempty"`
	// SystemMessage is a warning message displayed to the user.
	SystemMessage *string `json:"systemMessage,omitempty"`
	// Reason is feedback for Claude about the decision.
	Reason *string `json:"reason,omitempty"`

	// HookSpecificOutput contains event-specific output fields.
	HookSpecificOutput any `json:"hookSpecificOutput,omitempty"`
}

// AsyncHookJSONOutput indicates the hook will respond asynchronously.
type AsyncHookJSONOutput struct {
	// Async must be true for async hook output.
	// Python SDK uses async_ to avoid keyword conflict.
	Async bool `json:"async"`
	// AsyncTimeout is the timeout in milliseconds for the async operation.
	AsyncTimeout int `json:"asyncTimeout,omitempty"`
}

// =============================================================================
// Hook Context
// =============================================================================

// HookContext provides context information for hook callbacks.
type HookContext struct {
	// Signal is reserved for future abort signal support.
	// Currently always holds the parent context for cancellation.
	Signal context.Context `json:"-"`
}

// =============================================================================
// Hook Callback Type
// =============================================================================

// HookCallback is the function signature for hook callbacks.
// Go idiom: context.Context as first parameter, (result, error) return.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - input: Hook input (PreToolUseHookInput, PostToolUseHookInput, etc.)
//   - toolUseID: Optional tool use identifier (only for tool-related hooks)
//   - hookCtx: Hook context with signal support
//
// Returns:
//   - HookJSONOutput: The hook's response
//   - error: Non-nil if the callback encounters an error
type HookCallback func(
	ctx context.Context,
	input any,
	toolUseID *string,
	hookCtx HookContext,
) (HookJSONOutput, error)

// =============================================================================
// Hook Matcher
// =============================================================================

// HookMatcher defines which hooks to trigger for a given pattern.
type HookMatcher struct {
	// Matcher is a tool name pattern (e.g., "Bash", "Write|Edit|MultiEdit").
	// Empty string matches all tools (Python SDK: None).
	Matcher string `json:"matcher"`

	// Hooks are the callbacks to execute when the pattern matches.
	// Not serialized to JSON.
	Hooks []HookCallback `json:"-"`

	// Timeout is the maximum time in seconds for all hooks in this matcher.
	// Default is 60 seconds (Python SDK default).
	Timeout *float64 `json:"timeout,omitempty"`
}

// =============================================================================
// Hook Registration Types (for initialize request)
// =============================================================================

// HookMatcherConfig is the serializable format for the initialize request.
// This is what gets sent to the CLI during initialization.
type HookMatcherConfig struct {
	// Matcher is a tool name pattern.
	Matcher string `json:"matcher"`
	// HookCallbackIDs are the generated callback IDs for this matcher.
	HookCallbackIDs []string `json:"hookCallbackIds"`
	// Timeout is the maximum time in seconds.
	Timeout *float64 `json:"timeout,omitempty"`
}

// HookRegistration represents a hook registration for initialization.
type HookRegistration struct {
	// CallbackID is the unique identifier for this callback.
	CallbackID string `json:"callback_id"`
	// Matcher is the tool name pattern.
	Matcher string `json:"matcher"`
	// Timeout is the maximum time in seconds.
	Timeout *float64 `json:"timeout,omitempty"`
}
