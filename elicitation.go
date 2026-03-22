package claudecode

import "context"

// ElicitationRequest represents an MCP server requesting user input.
// This is sent when an MCP server needs information from the user,
// such as form input or URL-based authentication.
type ElicitationRequest struct {
	// ServerName is the name of the MCP server making the request.
	ServerName string `json:"serverName"`
	// Message is the prompt message to display to the user.
	Message string `json:"message"`
	// Mode is the elicitation mode ("form" or "url").
	Mode *string `json:"mode,omitempty"`
	// URL is the URL to present for URL-mode elicitation.
	URL *string `json:"url,omitempty"`
	// ElicitationID is the unique identifier for this elicitation request.
	ElicitationID *string `json:"elicitationId,omitempty"`
	// RequestedSchema is the JSON schema for form-mode input validation.
	RequestedSchema map[string]any `json:"requestedSchema,omitempty"`
}

// ElicitationResult represents the response to an elicitation request.
type ElicitationResult struct {
	// Action is the user's response action: "accept", "decline", or "cancel".
	Action string `json:"action"`
	// Content contains the form data when Action is "accept".
	Content map[string]any `json:"content,omitempty"`
}

// OnElicitation is the callback type for handling elicitation requests.
// The callback is invoked when an MCP server requests user input.
// It must be thread-safe as it may be invoked concurrently.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - request: The elicitation request with server name, message, and schema
//
// Returns:
//   - *ElicitationResult: The user's response (accept/decline/cancel with optional content)
//   - error: Non-nil if the callback encounters an error
type OnElicitation func(ctx context.Context, request ElicitationRequest) (*ElicitationResult, error)
