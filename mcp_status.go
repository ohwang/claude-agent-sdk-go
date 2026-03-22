package claudecode

import "github.com/severity1/claude-agent-sdk-go/internal/shared"

// MCP server connection status types.
type (
	// McpServerConnectionStatus represents the connection status of an MCP server.
	McpServerConnectionStatus = shared.McpServerConnectionStatus
	// McpServerInfo contains basic information about an MCP server.
	McpServerInfo = shared.McpServerInfo
	// ToolAnnotations provides metadata annotations for an MCP tool.
	ToolAnnotations = shared.ToolAnnotations
	// McpToolInfo describes a tool exposed by an MCP server in status responses.
	McpToolInfo = shared.McpToolInfo
	// McpServerStatusEntry represents the status of a single MCP server.
	McpServerStatusEntry = shared.McpServerStatusEntry
	// McpSetServersResult contains the result of a SetMcpServers operation.
	McpSetServersResult = shared.McpSetServersResult
)

// MCP server connection status constants.
const (
	// McpServerStatusConnected indicates the server is connected and operational.
	McpServerStatusConnected = shared.McpServerStatusConnected
	// McpServerStatusFailed indicates the server connection has failed.
	McpServerStatusFailed = shared.McpServerStatusFailed
	// McpServerStatusNeedsAuth indicates the server requires authentication.
	McpServerStatusNeedsAuth = shared.McpServerStatusNeedsAuth
	// McpServerStatusPending indicates the server connection is in progress.
	McpServerStatusPending = shared.McpServerStatusPending
	// McpServerStatusDisabled indicates the server has been disabled.
	McpServerStatusDisabled = shared.McpServerStatusDisabled
)
