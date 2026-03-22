package claudecode

import (
	"context"
	"testing"
)

// TestElicitationRequestFields tests ElicitationRequest struct field assignments.
func TestElicitationRequestFields(t *testing.T) {
	mode := "form"
	url := "https://example.com/auth"
	elicitationID := "elicit-123"

	req := ElicitationRequest{
		ServerName:    "test-server",
		Message:       "Please provide credentials",
		Mode:          &mode,
		URL:           &url,
		ElicitationID: &elicitationID,
		RequestedSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"username": map[string]any{"type": "string"},
			},
		},
	}

	assertElicitationServerName(t, req, "test-server")
	assertElicitationMessage(t, req, "Please provide credentials")
	assertElicitationMode(t, req, "form")
	assertElicitationURL(t, req, "https://example.com/auth")
	assertElicitationID(t, req, "elicit-123")
	assertElicitationSchemaNotNil(t, req)
}

// TestElicitationRequestMinimalFields tests ElicitationRequest with only required fields.
func TestElicitationRequestMinimalFields(t *testing.T) {
	req := ElicitationRequest{
		ServerName: "minimal-server",
		Message:    "Confirm action",
	}

	assertElicitationServerName(t, req, "minimal-server")
	assertElicitationMessage(t, req, "Confirm action")
	assertElicitationModeNil(t, req)
	assertElicitationURLNil(t, req)
	assertElicitationIDNil(t, req)
	assertElicitationSchemaNil(t, req)
}

// TestElicitationResultActions tests ElicitationResult with different actions.
func TestElicitationResultActions(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		content map[string]any
	}{
		{
			name:   "accept_with_content",
			action: "accept",
			content: map[string]any{
				"username": "admin",
				"password": "secret",
			},
		},
		{
			name:    "decline",
			action:  "decline",
			content: nil,
		},
		{
			name:    "cancel",
			action:  "cancel",
			content: nil,
		},
		{
			name:    "accept_without_content",
			action:  "accept",
			content: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ElicitationResult{
				Action:  tt.action,
				Content: tt.content,
			}
			assertElicitationResultAction(t, result, tt.action)
			if tt.content != nil {
				assertElicitationResultContentNotNil(t, result)
			} else {
				assertElicitationResultContentNil(t, result)
			}
		})
	}
}

// TestOnElicitationCallbackType tests that OnElicitation callback can be used correctly.
func TestOnElicitationCallbackType(t *testing.T) {
	called := false
	var callback OnElicitation = func(ctx context.Context, req ElicitationRequest) (*ElicitationResult, error) {
		called = true
		return &ElicitationResult{Action: "accept"}, nil
	}

	result, err := callback(context.Background(), ElicitationRequest{
		ServerName: "test",
		Message:    "test",
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !called {
		t.Error("Expected callback to be called")
	}
	assertElicitationResultAction(t, *result, "accept")
}

// TestWithOnElicitationOption tests the WithOnElicitation functional option.
func TestWithOnElicitationOption(t *testing.T) {
	t.Run("set_callback", func(t *testing.T) {
		var callback OnElicitation = func(ctx context.Context, req ElicitationRequest) (*ElicitationResult, error) {
			return &ElicitationResult{Action: "decline"}, nil
		}
		options := NewOptions(WithOnElicitation(callback))
		if options.OnElicitation == nil {
			t.Error("Expected OnElicitation to be set, got nil")
		}
	})

	t.Run("nil_callback", func(t *testing.T) {
		options := NewOptions(WithOnElicitation(nil))
		if options.OnElicitation != nil {
			t.Error("Expected OnElicitation to be nil")
		}
	})
}

// Assertion helpers

func assertElicitationServerName(t *testing.T, req ElicitationRequest, expected string) {
	t.Helper()
	if req.ServerName != expected {
		t.Errorf("Expected ServerName = %q, got %q", expected, req.ServerName)
	}
}

func assertElicitationMessage(t *testing.T, req ElicitationRequest, expected string) {
	t.Helper()
	if req.Message != expected {
		t.Errorf("Expected Message = %q, got %q", expected, req.Message)
	}
}

func assertElicitationMode(t *testing.T, req ElicitationRequest, expected string) {
	t.Helper()
	if req.Mode == nil {
		t.Error("Expected Mode to be set, got nil")
		return
	}
	if *req.Mode != expected {
		t.Errorf("Expected Mode = %q, got %q", expected, *req.Mode)
	}
}

func assertElicitationModeNil(t *testing.T, req ElicitationRequest) {
	t.Helper()
	if req.Mode != nil {
		t.Errorf("Expected Mode to be nil, got %q", *req.Mode)
	}
}

func assertElicitationURL(t *testing.T, req ElicitationRequest, expected string) {
	t.Helper()
	if req.URL == nil {
		t.Error("Expected URL to be set, got nil")
		return
	}
	if *req.URL != expected {
		t.Errorf("Expected URL = %q, got %q", expected, *req.URL)
	}
}

func assertElicitationURLNil(t *testing.T, req ElicitationRequest) {
	t.Helper()
	if req.URL != nil {
		t.Errorf("Expected URL to be nil, got %q", *req.URL)
	}
}

func assertElicitationID(t *testing.T, req ElicitationRequest, expected string) {
	t.Helper()
	if req.ElicitationID == nil {
		t.Error("Expected ElicitationID to be set, got nil")
		return
	}
	if *req.ElicitationID != expected {
		t.Errorf("Expected ElicitationID = %q, got %q", expected, *req.ElicitationID)
	}
}

func assertElicitationIDNil(t *testing.T, req ElicitationRequest) {
	t.Helper()
	if req.ElicitationID != nil {
		t.Errorf("Expected ElicitationID to be nil, got %q", *req.ElicitationID)
	}
}

func assertElicitationSchemaNotNil(t *testing.T, req ElicitationRequest) {
	t.Helper()
	if req.RequestedSchema == nil {
		t.Error("Expected RequestedSchema to be set, got nil")
	}
}

func assertElicitationSchemaNil(t *testing.T, req ElicitationRequest) {
	t.Helper()
	if req.RequestedSchema != nil {
		t.Error("Expected RequestedSchema to be nil")
	}
}

func assertElicitationResultAction(t *testing.T, result ElicitationResult, expected string) {
	t.Helper()
	if result.Action != expected {
		t.Errorf("Expected Action = %q, got %q", expected, result.Action)
	}
}

func assertElicitationResultContentNotNil(t *testing.T, result ElicitationResult) {
	t.Helper()
	if result.Content == nil {
		t.Error("Expected Content to be set, got nil")
	}
}

func assertElicitationResultContentNil(t *testing.T, result ElicitationResult) {
	t.Helper()
	if result.Content != nil {
		t.Error("Expected Content to be nil")
	}
}
