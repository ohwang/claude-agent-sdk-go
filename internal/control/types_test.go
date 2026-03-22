package control

import "testing"

// TestPermissionUpdateDestinationConstants tests all PermissionUpdateDestination values.
func TestPermissionUpdateDestinationConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant PermissionUpdateDestination
		expected string
	}{
		{"user_settings", PermissionDestinationUserSettings, "userSettings"},
		{"project_settings", PermissionDestinationProjectSettings, "projectSettings"},
		{"local_settings", PermissionDestinationLocalSettings, "localSettings"},
		{"session", PermissionDestinationSession, "session"},
		{"cli_arg", PermissionDestinationCLIArg, "cliArg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPermissionUpdateDestination(t, tt.constant, tt.expected)
		})
	}
}

// TestPermissionBehaviorConstants tests all PermissionBehavior values.
func TestPermissionBehaviorConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant PermissionBehavior
		expected string
	}{
		{"allow", PermissionBehaviorAllow, "allow"},
		{"deny", PermissionBehaviorDeny, "deny"},
		{"ask", PermissionBehaviorAsk, "ask"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertPermissionBehavior(t, tt.constant, tt.expected)
		})
	}
}

// TestPermissionUpdateDestinationType tests the type identity of PermissionUpdateDestination.
func TestPermissionUpdateDestinationType(t *testing.T) {
	// Verify the type is based on string
	var dest PermissionUpdateDestination = "customValue"
	if string(dest) != "customValue" {
		t.Errorf("Expected string conversion to work, got %q", string(dest))
	}
}

// TestPermissionBehaviorType tests the type identity of PermissionBehavior.
func TestPermissionBehaviorType(t *testing.T) {
	// Verify the type is based on string
	var behavior PermissionBehavior = "customBehavior"
	if string(behavior) != "customBehavior" {
		t.Errorf("Expected string conversion to work, got %q", string(behavior))
	}
}

// TestPermissionUpdateWithDestination tests PermissionUpdate struct with Destination field.
func TestPermissionUpdateWithDestination(t *testing.T) {
	dest := string(PermissionDestinationSession)
	behavior := string(PermissionBehaviorAllow)

	update := PermissionUpdate{
		Type:        PermissionUpdateTypeAddRules,
		Behavior:    &behavior,
		Destination: &dest,
		Rules: []PermissionRuleValue{
			{ToolName: "Bash"},
		},
	}

	if update.Type != PermissionUpdateTypeAddRules {
		t.Errorf("Expected Type = %q, got %q", PermissionUpdateTypeAddRules, update.Type)
	}
	if update.Destination == nil || *update.Destination != "session" {
		t.Errorf("Expected Destination = %q, got %v", "session", update.Destination)
	}
	if update.Behavior == nil || *update.Behavior != "allow" {
		t.Errorf("Expected Behavior = %q, got %v", "allow", update.Behavior)
	}
	if len(update.Rules) != 1 || update.Rules[0].ToolName != "Bash" {
		t.Errorf("Expected single rule with ToolName 'Bash', got %v", update.Rules)
	}
}

// Assertion helpers

func assertPermissionUpdateDestination(t *testing.T, actual PermissionUpdateDestination, expected string) {
	t.Helper()
	if string(actual) != expected {
		t.Errorf("Expected PermissionUpdateDestination = %q, got %q", expected, string(actual))
	}
}

func assertPermissionBehavior(t *testing.T, actual PermissionBehavior, expected string) {
	t.Helper()
	if string(actual) != expected {
		t.Errorf("Expected PermissionBehavior = %q, got %q", expected, string(actual))
	}
}
