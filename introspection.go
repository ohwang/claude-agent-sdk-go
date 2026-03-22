package claudecode

// SlashCommand represents a supported slash command in the CLI.
type SlashCommand struct {
	// Name is the command name (without leading slash).
	Name string `json:"name"`
	// Description is a human-readable description of the command.
	Description string `json:"description"`
	// ArgumentHint describes the expected arguments.
	ArgumentHint string `json:"argumentHint"`
}

// ModelInfo describes an available AI model.
type ModelInfo struct {
	// Value is the model identifier used in API calls.
	Value string `json:"value"`
	// DisplayName is a human-readable model name.
	DisplayName string `json:"displayName"`
	// Description is a human-readable description of the model.
	Description string `json:"description"`
	// SupportsEffort indicates whether the model supports effort levels.
	SupportsEffort *bool `json:"supportsEffort,omitempty"`
	// SupportedEffortLevels lists the available effort levels for this model.
	SupportedEffortLevels []string `json:"supportedEffortLevels,omitempty"`
	// SupportsAdaptiveThinking indicates whether the model supports adaptive thinking.
	SupportsAdaptiveThinking *bool `json:"supportsAdaptiveThinking,omitempty"`
}

// AgentInfo describes an available agent.
type AgentInfo struct {
	// Name is the agent identifier.
	Name string `json:"name"`
	// Description is a human-readable description of the agent.
	Description string `json:"description"`
	// Model is the optional model used by this agent.
	Model *string `json:"model,omitempty"`
}

// AccountInfo contains information about the authenticated account.
type AccountInfo struct {
	// Email is the account email address.
	Email *string `json:"email,omitempty"`
	// Organization is the organization name.
	Organization *string `json:"organization,omitempty"`
	// SubscriptionType indicates the subscription tier.
	SubscriptionType *string `json:"subscriptionType,omitempty"`
	// TokenSource describes where the auth token was obtained from.
	TokenSource *string `json:"tokenSource,omitempty"`
	// ApiKeySource describes the API key source.
	ApiKeySource *string `json:"apiKeySource,omitempty"`
	// ApiProvider indicates the API provider being used.
	ApiProvider *string `json:"apiProvider,omitempty"`
}
