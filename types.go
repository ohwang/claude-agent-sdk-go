package claudecode

import (
	"context"

	"github.com/severity1/claude-agent-sdk-go/internal/control"
	"github.com/severity1/claude-agent-sdk-go/internal/shared"
)

// Message represents any message type in the conversation.
type Message = shared.Message

// ContentBlock represents a content block within a message.
type ContentBlock = shared.ContentBlock

// UserMessage represents a message from the user.
type UserMessage = shared.UserMessage

// AssistantMessage represents a message from the assistant.
type AssistantMessage = shared.AssistantMessage

// AssistantMessageError represents error types in assistant messages.
type AssistantMessageError = shared.AssistantMessageError

// SystemMessage represents a system prompt message.
type SystemMessage = shared.SystemMessage

// ResultMessage represents a result or status message.
type ResultMessage = shared.ResultMessage

// TextBlock represents a text content block.
type TextBlock = shared.TextBlock

// ThinkingBlock represents a thinking content block.
type ThinkingBlock = shared.ThinkingBlock

// ToolUseBlock represents a tool usage content block.
type ToolUseBlock = shared.ToolUseBlock

// ToolResultBlock represents a tool result content block.
type ToolResultBlock = shared.ToolResultBlock

// RateLimitStatus represents the status of a rate limit check.
type RateLimitStatus = shared.RateLimitStatus

// RateLimitType represents the type of rate limit being applied.
type RateLimitType = shared.RateLimitType

// RateLimitInfo contains detailed rate limit status information.
type RateLimitInfo = shared.RateLimitInfo

// RateLimitEvent represents a rate limit status event from the CLI.
type RateLimitEvent = shared.RateLimitEvent

// TaskNotificationStatus represents the status of a completed task.
type TaskNotificationStatus = shared.TaskNotificationStatus

// TaskUsage contains resource usage information for a task.
type TaskUsage = shared.TaskUsage

// TaskStartedMessage represents a task_started system message.
type TaskStartedMessage = shared.TaskStartedMessage

// TaskProgressMessage represents a task_progress system message.
type TaskProgressMessage = shared.TaskProgressMessage

// TaskNotificationMessage represents a task_notification system message.
type TaskNotificationMessage = shared.TaskNotificationMessage

// StatusMessage represents a status system message.
type StatusMessage = shared.StatusMessage

// APIRetryMessage represents an api_retry system message.
type APIRetryMessage = shared.APIRetryMessage

// ToolProgressMessage represents a tool progress event.
type ToolProgressMessage = shared.ToolProgressMessage

// ToolUseSummaryMessage represents a tool use summary event.
type ToolUseSummaryMessage = shared.ToolUseSummaryMessage

// AuthStatusMessage represents an authentication status event.
type AuthStatusMessage = shared.AuthStatusMessage

// PromptSuggestionMessage represents a prompt suggestion event.
type PromptSuggestionMessage = shared.PromptSuggestionMessage

// StreamMessage represents a message in the streaming protocol.
type StreamMessage = shared.StreamMessage

// MessageIterator provides iteration over messages.
type MessageIterator = shared.MessageIterator

// StreamValidator tracks tool requests and results to detect incomplete streams.
type StreamValidator = shared.StreamValidator

// StreamIssue represents a validation issue found in the stream.
type StreamIssue = shared.StreamIssue

// StreamStats provides statistics about the message stream.
type StreamStats = shared.StreamStats

// Re-export message type constants
const (
	MessageTypeUser      = shared.MessageTypeUser
	MessageTypeAssistant = shared.MessageTypeAssistant
	MessageTypeSystem    = shared.MessageTypeSystem
	MessageTypeResult    = shared.MessageTypeResult

	// Control protocol message types
	MessageTypeControlRequest  = shared.MessageTypeControlRequest
	MessageTypeControlResponse = shared.MessageTypeControlResponse

	// Partial message streaming type
	MessageTypeStreamEvent = shared.MessageTypeStreamEvent

	// Additional message types
	MessageTypeRateLimitEvent   = shared.MessageTypeRateLimitEvent
	MessageTypeToolProgress     = shared.MessageTypeToolProgress
	MessageTypeToolUseSummary   = shared.MessageTypeToolUseSummary
	MessageTypeAuthStatus       = shared.MessageTypeAuthStatus
	MessageTypePromptSuggestion = shared.MessageTypePromptSuggestion
)

// Re-export content block type constants
const (
	ContentBlockTypeText       = shared.ContentBlockTypeText
	ContentBlockTypeThinking   = shared.ContentBlockTypeThinking
	ContentBlockTypeToolUse    = shared.ContentBlockTypeToolUse
	ContentBlockTypeToolResult = shared.ContentBlockTypeToolResult
)

// Re-export stream event type constants for Event["type"] discrimination.
const (
	StreamEventTypeContentBlockStart = shared.StreamEventTypeContentBlockStart
	StreamEventTypeContentBlockDelta = shared.StreamEventTypeContentBlockDelta
	StreamEventTypeContentBlockStop  = shared.StreamEventTypeContentBlockStop
	StreamEventTypeMessageStart      = shared.StreamEventTypeMessageStart
	StreamEventTypeMessageDelta      = shared.StreamEventTypeMessageDelta
	StreamEventTypeMessageStop       = shared.StreamEventTypeMessageStop
)

// Re-export AssistantMessageError constants
const (
	AssistantMessageErrorAuthFailed      = shared.AssistantMessageErrorAuthFailed
	AssistantMessageErrorBilling         = shared.AssistantMessageErrorBilling
	AssistantMessageErrorRateLimit       = shared.AssistantMessageErrorRateLimit
	AssistantMessageErrorInvalidRequest  = shared.AssistantMessageErrorInvalidRequest
	AssistantMessageErrorServer          = shared.AssistantMessageErrorServer
	AssistantMessageErrorUnknown         = shared.AssistantMessageErrorUnknown
	AssistantMessageErrorMaxOutputTokens = shared.AssistantMessageErrorMaxOutputTokens
)

// Re-export RateLimitStatus constants
const (
	RateLimitStatusAllowed        = shared.RateLimitStatusAllowed
	RateLimitStatusAllowedWarning = shared.RateLimitStatusAllowedWarning
	RateLimitStatusRejected       = shared.RateLimitStatusRejected
)

// Re-export RateLimitType constants
const (
	RateLimitTypeFiveHour       = shared.RateLimitTypeFiveHour
	RateLimitTypeSevenDay       = shared.RateLimitTypeSevenDay
	RateLimitTypeSevenDayOpus   = shared.RateLimitTypeSevenDayOpus
	RateLimitTypeSevenDaySonnet = shared.RateLimitTypeSevenDaySonnet
	RateLimitTypeOverage        = shared.RateLimitTypeOverage
)

// Re-export TaskNotificationStatus constants
const (
	TaskStatusCompleted = shared.TaskStatusCompleted
	TaskStatusFailed    = shared.TaskStatusFailed
	TaskStatusStopped   = shared.TaskStatusStopped
)

// AgentModel represents the model to use for an agent.
type AgentModel = shared.AgentModel

// AgentDefinition defines a programmatic subagent.
type AgentDefinition = shared.AgentDefinition

// MCP type aliases for backward compatibility.
// Implementation code lives in the mcp/ sub-package.
type (
	// McpToolResult represents the result of a tool call.
	McpToolResult = shared.McpToolResult
	// McpContent represents content returned by a tool.
	McpContent = shared.McpContent
	// McpToolDefinition describes a tool exposed by an MCP server.
	McpToolDefinition = shared.McpToolDefinition
	// McpSdkServerConfig configures an in-process SDK MCP server.
	McpSdkServerConfig = shared.McpSdkServerConfig
)

// McpServerTypeSdk represents an in-process SDK MCP server.
const McpServerTypeSdk = shared.McpServerTypeSdk

// Re-export agent model constants
const (
	AgentModelSonnet  = shared.AgentModelSonnet
	AgentModelOpus    = shared.AgentModelOpus
	AgentModelHaiku   = shared.AgentModelHaiku
	AgentModelInherit = shared.AgentModelInherit
)

// Transport abstracts the communication layer with Claude Code CLI.
// This interface stays in main package because it's used by client code.
type Transport interface {
	Connect(ctx context.Context) error
	SendMessage(ctx context.Context, message StreamMessage) error
	ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)
	Interrupt(ctx context.Context) error
	// SetModel changes the AI model during streaming session.
	SetModel(ctx context.Context, model *string) error
	// SetPermissionMode changes the permission mode during streaming session.
	SetPermissionMode(ctx context.Context, mode string) error
	// RewindFiles reverts tracked files to their state at a specific user message.
	// Requires file checkpointing to be enabled and control protocol initialized.
	RewindFiles(ctx context.Context, userMessageID string) error
	// GetMcpStatus returns the status of all configured MCP servers.
	GetMcpStatus(ctx context.Context) ([]McpServerStatusEntry, error)
	// ReconnectMcpServer reconnects a disconnected MCP server by name.
	ReconnectMcpServer(ctx context.Context, serverName string) error
	// ToggleMcpServer enables or disables an MCP server by name.
	ToggleMcpServer(ctx context.Context, serverName string, enabled bool) error
	// SetMcpServers dynamically replaces the set of MCP servers.
	SetMcpServers(ctx context.Context, servers map[string]any) (map[string]any, error)
	// StopTask stops a running background task by ID.
	StopTask(ctx context.Context, taskID string) error
	Close() error
	GetValidator() *StreamValidator
}

// RawControlMessage wraps raw control protocol messages for passthrough.
type RawControlMessage = shared.RawControlMessage

// StreamEvent represents a partial message update during streaming.
type StreamEvent = shared.StreamEvent

// Control protocol types for SDK-CLI bidirectional communication.

// SDKControlRequest represents a control request sent to the CLI.
type SDKControlRequest = control.SDKControlRequest

// SDKControlResponse represents a control response received from the CLI.
type SDKControlResponse = control.SDKControlResponse

// ControlResponse is the inner response structure.
type ControlResponse = control.Response

// InitializeRequest for control protocol handshake.
type InitializeRequest = control.InitializeRequest

// InitializeResponse from CLI with supported capabilities.
type InitializeResponse = control.InitializeResponse

// InterruptRequest to interrupt current operation via control protocol.
type InterruptRequest = control.InterruptRequest

// SetPermissionModeRequest to change permission mode via control protocol.
type SetPermissionModeRequest = control.SetPermissionModeRequest

// SetModelRequest to change AI model via control protocol.
type SetModelRequest = control.SetModelRequest

// McpStatusRequest to get MCP server status via control protocol.
type McpStatusRequest = control.McpStatusRequest

// McpReconnectRequest to reconnect an MCP server via control protocol.
type McpReconnectRequest = control.McpReconnectRequest

// McpToggleRequest to toggle an MCP server via control protocol.
type McpToggleRequest = control.McpToggleRequest

// McpSetServersRequest to dynamically replace MCP servers via control protocol.
type McpSetServersRequest = control.McpSetServersRequest

// StopTaskRequest to stop a running task via control protocol.
type StopTaskRequest = control.StopTaskRequest

// ControlProtocol manages bidirectional control communication with CLI.
type ControlProtocol = control.Protocol

// Re-export control protocol subtype constants
const (
	// Control request subtypes
	SubtypeInterrupt         = control.SubtypeInterrupt
	SubtypeCanUseTool        = control.SubtypeCanUseTool
	SubtypeInitialize        = control.SubtypeInitialize
	SubtypeSetPermissionMode = control.SubtypeSetPermissionMode
	SubtypeSetModel          = control.SubtypeSetModel
	SubtypeHookCallback      = control.SubtypeHookCallback
	SubtypeMcpMessage        = control.SubtypeMcpMessage
	SubtypeMcpStatus         = control.SubtypeMcpStatus
	SubtypeMcpReconnect      = control.SubtypeMcpReconnect
	SubtypeMcpToggle         = control.SubtypeMcpToggle
	SubtypeMcpSetServers     = control.SubtypeMcpSetServers
	SubtypeStopTask          = control.SubtypeStopTask

	// Control response subtypes
	ResponseSubtypeSuccess = control.ResponseSubtypeSuccess
	ResponseSubtypeError   = control.ResponseSubtypeError
)
