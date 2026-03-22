package shared

import (
	"encoding/json"
	"testing"
)

// --- WI-6: Rate Limit Event Tests ---

func TestRateLimitEvent_Type(t *testing.T) {
	event := &RateLimitEvent{
		RateLimitInfo: RateLimitInfo{
			Status: RateLimitStatusAllowed,
		},
		UUID:      "uuid-1",
		SessionID: "session-1",
	}
	if got := event.Type(); got != MessageTypeRateLimitEvent {
		t.Errorf("Type() = %q, want %q", got, MessageTypeRateLimitEvent)
	}
}

func TestRateLimitEvent_JSONRoundTrip(t *testing.T) {
	resetsAt := int64(1711100000)
	rateLimitType := RateLimitTypeFiveHour
	utilization := 0.85
	overageStatus := RateLimitStatusRejected
	overageResetsAt := int64(1711200000)
	overageReason := "payment_required"
	isUsingOverage := false
	threshold := 0.8

	event := &RateLimitEvent{
		MessageType: MessageTypeRateLimitEvent,
		RateLimitInfo: RateLimitInfo{
			Status:                RateLimitStatusAllowedWarning,
			ResetsAt:              &resetsAt,
			RateLimitType:         &rateLimitType,
			Utilization:           &utilization,
			OverageStatus:         &overageStatus,
			OverageResetsAt:       &overageResetsAt,
			OverageDisabledReason: &overageReason,
			IsUsingOverage:        &isUsingOverage,
			SurpassedThreshold:    &threshold,
		},
		UUID:      "uuid-rl",
		SessionID: "session-rl",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded RateLimitEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.RateLimitInfo.Status != RateLimitStatusAllowedWarning {
		t.Errorf("Status = %q, want %q", decoded.RateLimitInfo.Status, RateLimitStatusAllowedWarning)
	}
	if decoded.UUID != "uuid-rl" {
		t.Errorf("UUID = %q, want %q", decoded.UUID, "uuid-rl")
	}
	if decoded.SessionID != "session-rl" {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, "session-rl")
	}
	if decoded.RateLimitInfo.ResetsAt == nil || *decoded.RateLimitInfo.ResetsAt != resetsAt {
		t.Errorf("ResetsAt = %v, want %v", decoded.RateLimitInfo.ResetsAt, resetsAt)
	}
	if decoded.RateLimitInfo.RateLimitType == nil || *decoded.RateLimitInfo.RateLimitType != RateLimitTypeFiveHour {
		t.Errorf("RateLimitType = %v, want %q", decoded.RateLimitInfo.RateLimitType, RateLimitTypeFiveHour)
	}
	if decoded.RateLimitInfo.Utilization == nil || *decoded.RateLimitInfo.Utilization != utilization {
		t.Errorf("Utilization = %v, want %v", decoded.RateLimitInfo.Utilization, utilization)
	}
	if decoded.RateLimitInfo.OverageStatus == nil || *decoded.RateLimitInfo.OverageStatus != RateLimitStatusRejected {
		t.Errorf("OverageStatus = %v, want %q", decoded.RateLimitInfo.OverageStatus, RateLimitStatusRejected)
	}
	if decoded.RateLimitInfo.IsUsingOverage == nil || *decoded.RateLimitInfo.IsUsingOverage != false {
		t.Errorf("IsUsingOverage = %v, want false", decoded.RateLimitInfo.IsUsingOverage)
	}
	if decoded.RateLimitInfo.SurpassedThreshold == nil || *decoded.RateLimitInfo.SurpassedThreshold != threshold {
		t.Errorf("SurpassedThreshold = %v, want %v", decoded.RateLimitInfo.SurpassedThreshold, threshold)
	}
}

func TestRateLimitEvent_MinimalJSON(t *testing.T) {
	input := `{"type":"rate_limit_event","rate_limit_info":{"status":"allowed"},"uuid":"u1","session_id":"s1"}`

	var event RateLimitEvent
	if err := json.Unmarshal([]byte(input), &event); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if event.RateLimitInfo.Status != RateLimitStatusAllowed {
		t.Errorf("Status = %q, want %q", event.RateLimitInfo.Status, RateLimitStatusAllowed)
	}
	if event.RateLimitInfo.ResetsAt != nil {
		t.Error("ResetsAt should be nil for minimal event")
	}
	if event.RateLimitInfo.RateLimitType != nil {
		t.Error("RateLimitType should be nil for minimal event")
	}
}

func TestRateLimitStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		val  RateLimitStatus
		want string
	}{
		{"allowed", RateLimitStatusAllowed, "allowed"},
		{"allowed_warning", RateLimitStatusAllowedWarning, "allowed_warning"},
		{"rejected", RateLimitStatusRejected, "rejected"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.val) != tt.want {
				t.Errorf("got %q, want %q", tt.val, tt.want)
			}
		})
	}
}

func TestRateLimitTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		val  RateLimitType
		want string
	}{
		{"five_hour", RateLimitTypeFiveHour, "five_hour"},
		{"seven_day", RateLimitTypeSevenDay, "seven_day"},
		{"seven_day_opus", RateLimitTypeSevenDayOpus, "seven_day_opus"},
		{"seven_day_sonnet", RateLimitTypeSevenDaySonnet, "seven_day_sonnet"},
		{"overage", RateLimitTypeOverage, "overage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.val) != tt.want {
				t.Errorf("got %q, want %q", tt.val, tt.want)
			}
		})
	}
}

// --- WI-5: Task Management Message Types Tests ---

func TestTaskNotificationStatusConstants(t *testing.T) {
	tests := []struct {
		name string
		val  TaskNotificationStatus
		want string
	}{
		{"completed", TaskStatusCompleted, "completed"},
		{"failed", TaskStatusFailed, "failed"},
		{"stopped", TaskStatusStopped, "stopped"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.val) != tt.want {
				t.Errorf("got %q, want %q", tt.val, tt.want)
			}
		})
	}
}

func TestTaskStartedMessage_Type(t *testing.T) {
	msg := &TaskStartedMessage{
		Subtype:     "task_started",
		TaskID:      "task-1",
		Description: "Running analysis",
		UUID:        "uuid-ts",
		SessionID:   "session-ts",
	}
	if got := msg.Type(); got != MessageTypeSystem {
		t.Errorf("Type() = %q, want %q", got, MessageTypeSystem)
	}
}

func TestTaskStartedMessage_JSONRoundTrip(t *testing.T) {
	toolUseID := "tool-1"
	taskType := "agent"
	prompt := "Analyze this code"

	msg := &TaskStartedMessage{
		MessageType: MessageTypeSystem,
		Subtype:     "task_started",
		TaskID:      "task-1",
		ToolUseID:   &toolUseID,
		Description: "Running analysis",
		TaskType:    &taskType,
		Prompt:      &prompt,
		UUID:        "uuid-ts",
		SessionID:   "session-ts",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded TaskStartedMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Subtype != "task_started" {
		t.Errorf("Subtype = %q, want %q", decoded.Subtype, "task_started")
	}
	if decoded.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want %q", decoded.TaskID, "task-1")
	}
	if decoded.ToolUseID == nil || *decoded.ToolUseID != "tool-1" {
		t.Errorf("ToolUseID = %v, want %q", decoded.ToolUseID, "tool-1")
	}
	if decoded.TaskType == nil || *decoded.TaskType != "agent" {
		t.Errorf("TaskType = %v, want %q", decoded.TaskType, "agent")
	}
	if decoded.Prompt == nil || *decoded.Prompt != "Analyze this code" {
		t.Errorf("Prompt = %v, want %q", decoded.Prompt, "Analyze this code")
	}
}

func TestTaskProgressMessage_Type(t *testing.T) {
	msg := &TaskProgressMessage{
		Subtype:     "task_progress",
		TaskID:      "task-1",
		Description: "Processing...",
		Usage:       TaskUsage{TotalTokens: 100, ToolUses: 2, DurationMs: 5000},
		UUID:        "uuid-tp",
		SessionID:   "session-tp",
	}
	if got := msg.Type(); got != MessageTypeSystem {
		t.Errorf("Type() = %q, want %q", got, MessageTypeSystem)
	}
}

func TestTaskProgressMessage_JSONRoundTrip(t *testing.T) {
	lastTool := "Read"
	summary := "Read 3 files so far"

	msg := &TaskProgressMessage{
		MessageType:  MessageTypeSystem,
		Subtype:      "task_progress",
		TaskID:       "task-2",
		Description:  "Analyzing codebase",
		Usage:        TaskUsage{TotalTokens: 500, ToolUses: 5, DurationMs: 12000},
		LastToolName: &lastTool,
		Summary:      &summary,
		UUID:         "uuid-tp2",
		SessionID:    "session-tp2",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded TaskProgressMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Usage.TotalTokens != 500 {
		t.Errorf("Usage.TotalTokens = %d, want 500", decoded.Usage.TotalTokens)
	}
	if decoded.Usage.ToolUses != 5 {
		t.Errorf("Usage.ToolUses = %d, want 5", decoded.Usage.ToolUses)
	}
	if decoded.Usage.DurationMs != 12000 {
		t.Errorf("Usage.DurationMs = %d, want 12000", decoded.Usage.DurationMs)
	}
	if decoded.LastToolName == nil || *decoded.LastToolName != "Read" {
		t.Errorf("LastToolName = %v, want %q", decoded.LastToolName, "Read")
	}
	if decoded.Summary == nil || *decoded.Summary != "Read 3 files so far" {
		t.Errorf("Summary = %v, want %q", decoded.Summary, "Read 3 files so far")
	}
}

func TestTaskNotificationMessage_Type(t *testing.T) {
	msg := &TaskNotificationMessage{
		Subtype:    "task_notification",
		TaskID:     "task-1",
		Status:     TaskStatusCompleted,
		OutputFile: "/tmp/output.txt",
		Summary:    "Task completed successfully",
		UUID:       "uuid-tn",
		SessionID:  "session-tn",
	}
	if got := msg.Type(); got != MessageTypeSystem {
		t.Errorf("Type() = %q, want %q", got, MessageTypeSystem)
	}
}

func TestTaskNotificationMessage_JSONRoundTrip(t *testing.T) {
	usage := &TaskUsage{TotalTokens: 1000, ToolUses: 10, DurationMs: 30000}

	msg := &TaskNotificationMessage{
		MessageType: MessageTypeSystem,
		Subtype:     "task_notification",
		TaskID:      "task-3",
		Status:      TaskStatusFailed,
		OutputFile:  "/tmp/error.log",
		Summary:     "Task failed due to error",
		Usage:       usage,
		UUID:        "uuid-tn2",
		SessionID:   "session-tn2",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded TaskNotificationMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Status != TaskStatusFailed {
		t.Errorf("Status = %q, want %q", decoded.Status, TaskStatusFailed)
	}
	if decoded.OutputFile != "/tmp/error.log" {
		t.Errorf("OutputFile = %q, want %q", decoded.OutputFile, "/tmp/error.log")
	}
	if decoded.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if decoded.Usage.TotalTokens != 1000 {
		t.Errorf("Usage.TotalTokens = %d, want 1000", decoded.Usage.TotalTokens)
	}
}

func TestTaskUsage_JSONRoundTrip(t *testing.T) {
	usage := TaskUsage{
		TotalTokens: 42,
		ToolUses:    7,
		DurationMs:  9999,
	}

	data, err := json.Marshal(usage)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded TaskUsage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded != usage {
		t.Errorf("decoded = %+v, want %+v", decoded, usage)
	}
}

// --- WI-8: Additional Message Types Tests ---

func TestStatusMessage_Type(t *testing.T) {
	msg := &StatusMessage{
		Subtype:   "status",
		UUID:      "uuid-s",
		SessionID: "session-s",
	}
	if got := msg.Type(); got != MessageTypeSystem {
		t.Errorf("Type() = %q, want %q", got, MessageTypeSystem)
	}
}

func TestStatusMessage_JSONRoundTrip(t *testing.T) {
	status := "compacting"
	permMode := "auto"

	msg := &StatusMessage{
		MessageType:    MessageTypeSystem,
		Subtype:        "status",
		Status:         &status,
		PermissionMode: &permMode,
		UUID:           "uuid-s2",
		SessionID:      "session-s2",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded StatusMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Status == nil || *decoded.Status != "compacting" {
		t.Errorf("Status = %v, want %q", decoded.Status, "compacting")
	}
	if decoded.PermissionMode == nil || *decoded.PermissionMode != "auto" {
		t.Errorf("PermissionMode = %v, want %q", decoded.PermissionMode, "auto")
	}
}

func TestStatusMessage_NullStatus(t *testing.T) {
	input := `{"type":"system","subtype":"status","status":null,"uuid":"u1","session_id":"s1"}`

	var msg StatusMessage
	if err := json.Unmarshal([]byte(input), &msg); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if msg.Status != nil {
		t.Errorf("Status should be nil for null status, got %v", msg.Status)
	}
}

func TestAPIRetryMessage_Type(t *testing.T) {
	msg := &APIRetryMessage{
		Subtype:      "api_retry",
		Attempt:      1,
		MaxRetries:   3,
		RetryDelayMs: 1000,
		Error:        "rate limited",
		UUID:         "uuid-ar",
		SessionID:    "session-ar",
	}
	if got := msg.Type(); got != MessageTypeSystem {
		t.Errorf("Type() = %q, want %q", got, MessageTypeSystem)
	}
}

func TestAPIRetryMessage_JSONRoundTrip(t *testing.T) {
	errorStatus := 429

	msg := &APIRetryMessage{
		MessageType:  MessageTypeSystem,
		Subtype:      "api_retry",
		Attempt:      2,
		MaxRetries:   5,
		RetryDelayMs: 2000,
		ErrorStatus:  &errorStatus,
		Error:        "Too many requests",
		UUID:         "uuid-ar2",
		SessionID:    "session-ar2",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded APIRetryMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Attempt != 2 {
		t.Errorf("Attempt = %d, want 2", decoded.Attempt)
	}
	if decoded.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", decoded.MaxRetries)
	}
	if decoded.RetryDelayMs != 2000 {
		t.Errorf("RetryDelayMs = %d, want 2000", decoded.RetryDelayMs)
	}
	if decoded.ErrorStatus == nil || *decoded.ErrorStatus != 429 {
		t.Errorf("ErrorStatus = %v, want 429", decoded.ErrorStatus)
	}
	if decoded.Error != "Too many requests" {
		t.Errorf("Error = %q, want %q", decoded.Error, "Too many requests")
	}
}

func TestToolProgressMessage_Type(t *testing.T) {
	msg := &ToolProgressMessage{
		ToolUseID:          "tool-1",
		ToolName:           "Read",
		ElapsedTimeSeconds: 3.5,
		UUID:               "uuid-tpm",
		SessionID:          "session-tpm",
	}
	if got := msg.Type(); got != MessageTypeToolProgress {
		t.Errorf("Type() = %q, want %q", got, MessageTypeToolProgress)
	}
}

func TestToolProgressMessage_JSONRoundTrip(t *testing.T) {
	parentID := "parent-tool-1"
	taskID := "task-42"

	msg := &ToolProgressMessage{
		MessageType:        MessageTypeToolProgress,
		ToolUseID:          "tool-99",
		ToolName:           "Bash",
		ParentToolUseID:    &parentID,
		ElapsedTimeSeconds: 12.75,
		TaskID:             &taskID,
		UUID:               "uuid-tpm2",
		SessionID:          "session-tpm2",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ToolProgressMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ToolUseID != "tool-99" {
		t.Errorf("ToolUseID = %q, want %q", decoded.ToolUseID, "tool-99")
	}
	if decoded.ToolName != "Bash" {
		t.Errorf("ToolName = %q, want %q", decoded.ToolName, "Bash")
	}
	if decoded.ParentToolUseID == nil || *decoded.ParentToolUseID != "parent-tool-1" {
		t.Errorf("ParentToolUseID = %v, want %q", decoded.ParentToolUseID, "parent-tool-1")
	}
	if decoded.ElapsedTimeSeconds != 12.75 {
		t.Errorf("ElapsedTimeSeconds = %f, want 12.75", decoded.ElapsedTimeSeconds)
	}
	if decoded.TaskID == nil || *decoded.TaskID != "task-42" {
		t.Errorf("TaskID = %v, want %q", decoded.TaskID, "task-42")
	}
}

func TestToolUseSummaryMessage_Type(t *testing.T) {
	msg := &ToolUseSummaryMessage{
		Summary:             "Edited 3 files",
		PrecedingToolUseIDs: []string{"tool-1", "tool-2", "tool-3"},
		UUID:                "uuid-tus",
		SessionID:           "session-tus",
	}
	if got := msg.Type(); got != MessageTypeToolUseSummary {
		t.Errorf("Type() = %q, want %q", got, MessageTypeToolUseSummary)
	}
}

func TestToolUseSummaryMessage_JSONRoundTrip(t *testing.T) {
	msg := &ToolUseSummaryMessage{
		MessageType:         MessageTypeToolUseSummary,
		Summary:             "Read and analyzed 5 files",
		PrecedingToolUseIDs: []string{"t1", "t2", "t3", "t4", "t5"},
		UUID:                "uuid-tus2",
		SessionID:           "session-tus2",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ToolUseSummaryMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Summary != "Read and analyzed 5 files" {
		t.Errorf("Summary = %q, want %q", decoded.Summary, "Read and analyzed 5 files")
	}
	if len(decoded.PrecedingToolUseIDs) != 5 {
		t.Errorf("PrecedingToolUseIDs length = %d, want 5", len(decoded.PrecedingToolUseIDs))
	}
	if decoded.PrecedingToolUseIDs[0] != "t1" {
		t.Errorf("PrecedingToolUseIDs[0] = %q, want %q", decoded.PrecedingToolUseIDs[0], "t1")
	}
}

func TestAuthStatusMessage_Type(t *testing.T) {
	msg := &AuthStatusMessage{
		IsAuthenticating: true,
		Output:           []string{"Authenticating..."},
		UUID:             "uuid-as",
		SessionID:        "session-as",
	}
	if got := msg.Type(); got != MessageTypeAuthStatus {
		t.Errorf("Type() = %q, want %q", got, MessageTypeAuthStatus)
	}
}

func TestAuthStatusMessage_JSONRoundTrip(t *testing.T) {
	authErr := "token expired"

	msg := &AuthStatusMessage{
		MessageType:      MessageTypeAuthStatus,
		IsAuthenticating: false,
		Output:           []string{"Login required", "Opening browser..."},
		Error:            &authErr,
		UUID:             "uuid-as2",
		SessionID:        "session-as2",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded AuthStatusMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.IsAuthenticating != false {
		t.Error("IsAuthenticating should be false")
	}
	if len(decoded.Output) != 2 {
		t.Errorf("Output length = %d, want 2", len(decoded.Output))
	}
	if decoded.Output[0] != "Login required" {
		t.Errorf("Output[0] = %q, want %q", decoded.Output[0], "Login required")
	}
	if decoded.Error == nil || *decoded.Error != "token expired" {
		t.Errorf("Error = %v, want %q", decoded.Error, "token expired")
	}
}

func TestPromptSuggestionMessage_Type(t *testing.T) {
	msg := &PromptSuggestionMessage{
		Suggestion: "Try asking about the architecture",
		UUID:       "uuid-ps",
		SessionID:  "session-ps",
	}
	if got := msg.Type(); got != MessageTypePromptSuggestion {
		t.Errorf("Type() = %q, want %q", got, MessageTypePromptSuggestion)
	}
}

func TestPromptSuggestionMessage_JSONRoundTrip(t *testing.T) {
	msg := &PromptSuggestionMessage{
		MessageType: MessageTypePromptSuggestion,
		Suggestion:  "Explain the error in detail",
		UUID:        "uuid-ps2",
		SessionID:   "session-ps2",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded PromptSuggestionMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Suggestion != "Explain the error in detail" {
		t.Errorf("Suggestion = %q, want %q", decoded.Suggestion, "Explain the error in detail")
	}
	if decoded.UUID != "uuid-ps2" {
		t.Errorf("UUID = %q, want %q", decoded.UUID, "uuid-ps2")
	}
}

// --- ResultMessage new fields tests ---

func TestResultMessage_NewFields_JSONRoundTrip(t *testing.T) {
	stopReason := "end_turn"
	result := "Task completed"
	costUSD := 0.05

	msg := &ResultMessage{
		MessageType:   MessageTypeResult,
		Subtype:       "result",
		DurationMs:    5000,
		DurationAPIMs: 4500,
		IsError:       false,
		NumTurns:      3,
		SessionID:     "session-rm",
		TotalCostUSD:  &costUSD,
		Result:        &result,
		StopReason:    &stopReason,
		ModelUsage: map[string]any{
			"input_tokens":  float64(100),
			"output_tokens": float64(200),
		},
		PermissionDenials: []map[string]any{
			{"tool": "Bash", "reason": "denied by policy"},
		},
		Errors: []string{"warning: large file skipped"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ResultMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.StopReason == nil || *decoded.StopReason != "end_turn" {
		t.Errorf("StopReason = %v, want %q", decoded.StopReason, "end_turn")
	}
	if decoded.ModelUsage == nil {
		t.Fatal("ModelUsage should not be nil")
	}
	if decoded.ModelUsage["input_tokens"] != float64(100) {
		t.Errorf("ModelUsage[input_tokens] = %v, want 100", decoded.ModelUsage["input_tokens"])
	}
	if len(decoded.PermissionDenials) != 1 {
		t.Fatalf("PermissionDenials length = %d, want 1", len(decoded.PermissionDenials))
	}
	if decoded.PermissionDenials[0]["tool"] != "Bash" {
		t.Errorf("PermissionDenials[0][tool] = %v, want %q", decoded.PermissionDenials[0]["tool"], "Bash")
	}
	if len(decoded.Errors) != 1 || decoded.Errors[0] != "warning: large file skipped" {
		t.Errorf("Errors = %v, want [\"warning: large file skipped\"]", decoded.Errors)
	}
}

func TestResultMessage_NewFields_Omitempty(t *testing.T) {
	msg := &ResultMessage{
		MessageType:   MessageTypeResult,
		Subtype:       "result",
		DurationMs:    1000,
		DurationAPIMs: 900,
		IsError:       false,
		NumTurns:      1,
		SessionID:     "session-rm2",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	for _, field := range []string{"stop_reason", "model_usage", "permission_denials", "errors"} {
		if _, exists := raw[field]; exists {
			t.Errorf("Expected field %q to be omitted when nil/empty", field)
		}
	}
}

// --- AssistantMessageError new constant test ---

func TestAssistantMessageErrorMaxOutputTokens(t *testing.T) {
	if string(AssistantMessageErrorMaxOutputTokens) != "max_output_tokens" {
		t.Errorf("AssistantMessageErrorMaxOutputTokens = %q, want %q",
			AssistantMessageErrorMaxOutputTokens, "max_output_tokens")
	}
}

// --- Message type constant tests ---

func TestNewMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want string
	}{
		{"rate_limit_event", MessageTypeRateLimitEvent, "rate_limit_event"},
		{"tool_progress", MessageTypeToolProgress, "tool_progress"},
		{"tool_use_summary", MessageTypeToolUseSummary, "tool_use_summary"},
		{"auth_status", MessageTypeAuthStatus, "auth_status"},
		{"prompt_suggestion", MessageTypePromptSuggestion, "prompt_suggestion"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val != tt.want {
				t.Errorf("got %q, want %q", tt.val, tt.want)
			}
		})
	}
}

// --- Interface compliance tests ---

func TestNewMessageTypes_InterfaceCompliance(t *testing.T) {
	// All new message types must implement the Message interface
	messages := []Message{
		&RateLimitEvent{RateLimitInfo: RateLimitInfo{Status: RateLimitStatusAllowed}},
		&TaskStartedMessage{Subtype: "task_started", TaskID: "t1"},
		&TaskProgressMessage{Subtype: "task_progress", TaskID: "t1", Usage: TaskUsage{}},
		&TaskNotificationMessage{Subtype: "task_notification", TaskID: "t1", Status: TaskStatusCompleted},
		&StatusMessage{Subtype: "status"},
		&APIRetryMessage{Subtype: "api_retry", Error: "err"},
		&ToolProgressMessage{ToolUseID: "t1", ToolName: "Read"},
		&ToolUseSummaryMessage{Summary: "sum"},
		&AuthStatusMessage{Output: []string{}},
		&PromptSuggestionMessage{Suggestion: "try this"},
	}

	expectedTypes := []string{
		MessageTypeRateLimitEvent,
		MessageTypeSystem,
		MessageTypeSystem,
		MessageTypeSystem,
		MessageTypeSystem,
		MessageTypeSystem,
		MessageTypeToolProgress,
		MessageTypeToolUseSummary,
		MessageTypeAuthStatus,
		MessageTypePromptSuggestion,
	}

	for i, msg := range messages {
		if got := msg.Type(); got != expectedTypes[i] {
			t.Errorf("message[%d] (%T).Type() = %q, want %q", i, msg, got, expectedTypes[i])
		}
	}
}
