package shared

// McpServerConnectionStatus represents the connection status of an MCP server.
type McpServerConnectionStatus string

const (
	// McpServerStatusConnected indicates the server is connected and operational.
	McpServerStatusConnected McpServerConnectionStatus = "connected"
	// McpServerStatusFailed indicates the server connection has failed.
	McpServerStatusFailed McpServerConnectionStatus = "failed"
	// McpServerStatusNeedsAuth indicates the server requires authentication.
	McpServerStatusNeedsAuth McpServerConnectionStatus = "needs-auth"
	// McpServerStatusPending indicates the server connection is in progress.
	McpServerStatusPending McpServerConnectionStatus = "pending"
	// McpServerStatusDisabled indicates the server has been disabled.
	McpServerStatusDisabled McpServerConnectionStatus = "disabled"
)

// McpServerInfo contains basic information about an MCP server.
type McpServerInfo struct {
	// Name is the server's display name.
	Name string `json:"name"`
	// Version is the server's version string.
	Version string `json:"version"`
}

// McpToolInfo describes a tool exposed by an MCP server in status responses.
type McpToolInfo struct {
	// Name is the tool's name.
	Name string `json:"name"`
	// Description is an optional human-readable description.
	Description *string `json:"description,omitempty"`
	// Annotations provides optional metadata about the tool's behavior.
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

// McpServerStatusEntry represents the status of a single MCP server.
type McpServerStatusEntry struct {
	// Name is the server's identifier.
	Name string `json:"name"`
	// Status is the server's connection status.
	Status McpServerConnectionStatus `json:"status"`
	// ServerInfo contains optional server metadata.
	ServerInfo *McpServerInfo `json:"serverInfo,omitempty"`
	// Error contains an error message if the server is in a failed state.
	Error *string `json:"error,omitempty"`
	// Config contains optional server configuration details.
	Config map[string]any `json:"config,omitempty"`
	// Scope indicates the configuration scope (e.g., "project", "user").
	Scope *string `json:"scope,omitempty"`
	// Tools lists the tools provided by this server.
	Tools []McpToolInfo `json:"tools,omitempty"`
}

// McpSetServersResult contains the result of a SetMcpServers operation.
type McpSetServersResult struct {
	// Added lists the names of servers that were added.
	Added []string `json:"added"`
	// Removed lists the names of servers that were removed.
	Removed []string `json:"removed"`
}
