package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Test Functions (primary purpose)
// =============================================================================

// --- Type Tests ---

// TestSDKSessionInfoJSONRoundTrip tests JSON serialization of SDKSessionInfo.
func TestSDKSessionInfoJSONRoundTrip(t *testing.T) {
	size := int64(1234)
	title := "My session"
	prompt := "Hello Claude"
	branch := "main"
	cwd := "/tmp/project"
	tag := "experiment"
	created := int64(1700000000000)

	info := SDKSessionInfo{
		SessionID:    "550e8400-e29b-41d4-a716-446655440000",
		Summary:      "Test summary",
		LastModified: 1700000001000,
		FileSize:     &size,
		CustomTitle:  &title,
		FirstPrompt:  &prompt,
		GitBranch:    &branch,
		Cwd:          &cwd,
		Tag:          &tag,
		CreatedAt:    &created,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded SDKSessionInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.SessionID != info.SessionID {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, info.SessionID)
	}
	if decoded.Summary != info.Summary {
		t.Errorf("Summary = %q, want %q", decoded.Summary, info.Summary)
	}
	if decoded.LastModified != info.LastModified {
		t.Errorf("LastModified = %d, want %d", decoded.LastModified, info.LastModified)
	}
	if decoded.FileSize == nil || *decoded.FileSize != size {
		t.Errorf("FileSize = %v, want %d", decoded.FileSize, size)
	}
	if decoded.CustomTitle == nil || *decoded.CustomTitle != title {
		t.Errorf("CustomTitle = %v, want %q", decoded.CustomTitle, title)
	}
	if decoded.Tag == nil || *decoded.Tag != tag {
		t.Errorf("Tag = %v, want %q", decoded.Tag, tag)
	}
}

// TestSDKSessionInfoJSONOmitsNil tests that nil optional fields are omitted from JSON.
func TestSDKSessionInfoJSONOmitsNil(t *testing.T) {
	info := SDKSessionInfo{
		SessionID:    "test-id",
		Summary:      "summary",
		LastModified: 1000,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	s := string(data)
	optionalFields := []string{"fileSize", "customTitle", "firstPrompt", "gitBranch", "cwd", "tag", "createdAt"}
	for _, field := range optionalFields {
		if strings.Contains(s, field) {
			t.Errorf("Expected %q to be omitted from JSON, but found in: %s", field, s)
		}
	}
}

// TestSessionMessageJSONRoundTrip tests JSON serialization of SessionMessage.
func TestSessionMessageJSONRoundTrip(t *testing.T) {
	msg := SessionMessage{
		Type:      "user",
		UUID:      "abc-123",
		SessionID: "session-456",
		Message: map[string]any{
			"role":    "user",
			"content": "Hello",
		},
		ParentToolUseID: nil,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded SessionMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Type != msg.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, msg.Type)
	}
	if decoded.UUID != msg.UUID {
		t.Errorf("UUID = %q, want %q", decoded.UUID, msg.UUID)
	}
	if decoded.SessionID != msg.SessionID {
		t.Errorf("SessionID = %q, want %q", decoded.SessionID, msg.SessionID)
	}
}

// --- UUID Validation Tests ---

// TestIsValidUUID tests UUID validation.
func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid_lowercase", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid_uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"valid_mixed", "550e8400-E29B-41d4-a716-446655440000", true},
		{"empty", "", false},
		{"too_short", "550e8400", false},
		{"no_dashes", "550e8400e29b41d4a716446655440000", false},
		{"invalid_chars", "550g8400-e29b-41d4-a716-446655440000", false},
		{"extra_dash", "550e8400-e29b-41d4-a716-4466554400-00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidUUID(tt.input)
			if got != tt.want {
				t.Errorf("isValidUUID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- Path Sanitization Tests ---

// TestSanitizePath tests path sanitization.
func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple_unix", "/home/user/project", "-home-user-project"},
		{"alphanumeric", "abc123", "abc123"},
		{"dots_and_slashes", "../relative/path", "---relative-path"},
		{"spaces", "path with spaces", "path-with-spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizePath(tt.input)
			if got != tt.want {
				t.Errorf("sanitizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestSanitizePathLong tests that long paths are truncated with a hash suffix.
func TestSanitizePathLong(t *testing.T) {
	longPath := "/" + strings.Repeat("a", 300)
	result := sanitizePath(longPath)

	if len(result) <= maxSanitizedLength {
		t.Errorf("Expected result longer than %d chars, got %d", maxSanitizedLength, len(result))
	}
	if !strings.Contains(result, "-") {
		t.Error("Expected hash suffix with dash separator")
	}

	// Verify the prefix is the truncated sanitized path.
	sanitized := sanitizeRe.ReplaceAllString(longPath, "-")
	if !strings.HasPrefix(result, sanitized[:maxSanitizedLength]) {
		t.Error("Result should start with truncated sanitized path")
	}
}

// TestSimpleHash tests the hash function.
func TestSimpleHash(t *testing.T) {
	// Verify deterministic output.
	h1 := simpleHash("test")
	h2 := simpleHash("test")
	if h1 != h2 {
		t.Errorf("simpleHash not deterministic: %q != %q", h1, h2)
	}

	// Different inputs should produce different hashes.
	h3 := simpleHash("other")
	if h1 == h3 {
		t.Error("Expected different hashes for different inputs")
	}

	// Empty string.
	h4 := simpleHash("")
	if h4 != "0" {
		t.Errorf("simpleHash('') = %q, want %q", h4, "0")
	}
}

// --- JSON Field Extraction Tests ---

// TestExtractJSONStringField tests JSON string field extraction without full parsing.
func TestExtractJSONStringField(t *testing.T) {
	tests := []struct {
		name string
		text string
		key  string
		want string
	}{
		{"compact", `{"type":"user","cwd":"/tmp"}`, "cwd", "/tmp"},
		{"with_space", `{"type": "user", "cwd": "/tmp"}`, "cwd", "/tmp"},
		{"not_found", `{"type":"user"}`, "cwd", ""},
		{"escaped_value", `{"name":"hello\"world"}`, "name", `hello"world`},
		{"nested", `{"outer":{"inner":"value"}}`, "inner", "value"},
		{"empty_value", `{"key":""}`, "key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSONStringField(tt.text, tt.key)
			if got != tt.want {
				t.Errorf("extractJSONStringField(%q, %q) = %q, want %q", tt.text, tt.key, got, tt.want)
			}
		})
	}
}

// TestExtractLastJSONStringField tests extracting the last occurrence.
func TestExtractLastJSONStringField(t *testing.T) {
	tests := []struct {
		name string
		text string
		key  string
		want string
	}{
		{
			"single_occurrence",
			`{"gitBranch":"main"}`,
			"gitBranch",
			"main",
		},
		{
			"multiple_occurrences",
			`{"gitBranch":"main"}` + "\n" + `{"gitBranch":"feature"}`,
			"gitBranch",
			"feature",
		},
		{
			"not_found",
			`{"type":"user"}`,
			"gitBranch",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractLastJSONStringField(tt.text, tt.key)
			if got != tt.want {
				t.Errorf("extractLastJSONStringField(%q, %q) = %q, want %q", tt.text, tt.key, got, tt.want)
			}
		})
	}
}

// --- First Prompt Extraction Tests ---

// TestExtractFirstPromptFromHead tests first prompt extraction.
func TestExtractFirstPromptFromHead(t *testing.T) {
	tests := []struct {
		name string
		head string
		want string
	}{
		{
			"simple_user_message",
			`{"type":"user","message":{"role":"user","content":"Hello Claude"}}`,
			"Hello Claude",
		},
		{
			"skips_tool_result",
			`{"type":"user","message":{"role":"user","content":"tool_result stuff"},"tool_result":true}` + "\n" +
				`{"type":"user","message":{"role":"user","content":"Real prompt"}}`,
			"Real prompt",
		},
		{
			"skips_meta",
			`{"type":"user","message":{"role":"user","content":"Meta content"},"isMeta":true}` + "\n" +
				`{"type":"user","message":{"role":"user","content":"Actual prompt"}}`,
			"Actual prompt",
		},
		{
			"content_array",
			`{"type":"user","message":{"role":"user","content":[{"type":"text","text":"Array prompt"}]}}`,
			"Array prompt",
		},
		{
			"truncates_long_prompt",
			`{"type":"user","message":{"role":"user","content":"` + strings.Repeat("a", 250) + `"}}`,
			strings.Repeat("a", 200) + "\u2026",
		},
		{
			"empty_content",
			`{"type":"user","message":{"role":"user","content":""}}`,
			"",
		},
		{
			"no_user_messages",
			`{"type":"assistant","message":{"role":"assistant","content":"Hi!"}}`,
			"",
		},
		{
			"skips_session_start_hook",
			`{"type":"user","message":{"role":"user","content":"<session-start-hook>data"}}` + "\n" +
				`{"type":"user","message":{"role":"user","content":"Real message"}}`,
			"Real message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFirstPromptFromHead(tt.head)
			if got != tt.want {
				t.Errorf("extractFirstPromptFromHead() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Lite Session File Tests ---

// TestReadSessionLite tests the lite session file reader.
func TestReadSessionLite(t *testing.T) {
	t.Run("reads_small_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.jsonl")
		content := `{"type":"user","message":{"role":"user","content":"Hello"}}` + "\n"
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		lite, err := readSessionLite(path)
		if err != nil {
			t.Fatalf("readSessionLite error: %v", err)
		}
		if lite == nil {
			t.Fatal("Expected non-nil result")
		}
		if lite.head != content {
			t.Errorf("head = %q, want %q", lite.head, content)
		}
		if lite.tail != lite.head {
			t.Error("Expected tail == head for small files")
		}
		if lite.size != int64(len(content)) {
			t.Errorf("size = %d, want %d", lite.size, len(content))
		}
	})

	t.Run("returns_nil_for_empty_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "empty.jsonl")
		if err := os.WriteFile(path, []byte{}, 0600); err != nil {
			t.Fatal(err)
		}

		lite, err := readSessionLite(path)
		if err != nil {
			t.Fatalf("readSessionLite error: %v", err)
		}
		if lite != nil {
			t.Error("Expected nil for empty file")
		}
	})

	t.Run("returns_error_for_missing_file", func(t *testing.T) {
		_, err := readSessionLite("/nonexistent/file.jsonl")
		if err == nil {
			t.Error("Expected error for missing file")
		}
	})
}

// --- Parse Session Info Tests ---

// TestParseSessionInfoFromLite tests session info parsing from lite data.
func TestParseSessionInfoFromLite(t *testing.T) {
	t.Run("basic_session", func(t *testing.T) {
		head := `{"type":"user","message":{"role":"user","content":"Hello Claude"},"cwd":"/tmp/project","gitBranch":"main","timestamp":"2024-01-15T10:00:00.000Z"}` + "\n" +
			`{"type":"assistant","message":{"role":"assistant","content":"Hi!"}}`
		lite := &liteSessionFile{
			mtime: 1705312800000,
			size:  int64(len(head)),
			head:  head,
			tail:  head,
		}

		info := parseSessionInfoFromLite("550e8400-e29b-41d4-a716-446655440000", lite, nil)
		if info == nil {
			t.Fatal("Expected non-nil info")
		}
		if info.SessionID != "550e8400-e29b-41d4-a716-446655440000" {
			t.Errorf("SessionID = %q", info.SessionID)
		}
		if info.Summary != "Hello Claude" {
			t.Errorf("Summary = %q, want %q", info.Summary, "Hello Claude")
		}
		if info.FirstPrompt == nil || *info.FirstPrompt != "Hello Claude" {
			t.Errorf("FirstPrompt = %v, want %q", info.FirstPrompt, "Hello Claude")
		}
		if info.GitBranch == nil || *info.GitBranch != "main" {
			t.Errorf("GitBranch = %v, want %q", info.GitBranch, "main")
		}
		if info.Cwd == nil || *info.Cwd != "/tmp/project" {
			t.Errorf("Cwd = %v, want %q", info.Cwd, "/tmp/project")
		}
		if info.CreatedAt == nil {
			t.Error("Expected non-nil CreatedAt")
		}
	})

	t.Run("sidechain_returns_nil", func(t *testing.T) {
		head := `{"type":"user","isSidechain":true,"message":{"role":"user","content":"Hello"}}`
		lite := &liteSessionFile{
			mtime: 1705312800000,
			size:  int64(len(head)),
			head:  head,
			tail:  head,
		}

		info := parseSessionInfoFromLite("test-id", lite, nil)
		if info != nil {
			t.Error("Expected nil for sidechain session")
		}
	})

	t.Run("custom_title_wins", func(t *testing.T) {
		head := `{"type":"user","message":{"role":"user","content":"Hello"}}` + "\n"
		tail := head + `{"type":"custom-title","customTitle":"My Custom Title"}` + "\n"
		lite := &liteSessionFile{
			mtime: 1705312800000,
			size:  int64(len(tail)),
			head:  tail,
			tail:  tail,
		}

		info := parseSessionInfoFromLite("test-id", lite, nil)
		if info == nil {
			t.Fatal("Expected non-nil info")
		}
		if info.Summary != "My Custom Title" {
			t.Errorf("Summary = %q, want %q", info.Summary, "My Custom Title")
		}
		if info.CustomTitle == nil || *info.CustomTitle != "My Custom Title" {
			t.Errorf("CustomTitle = %v, want %q", info.CustomTitle, "My Custom Title")
		}
	})

	t.Run("tag_extraction", func(t *testing.T) {
		head := `{"type":"user","message":{"role":"user","content":"Hello"}}` + "\n"
		tail := head + `{"type":"tag","tag":"experiment","sessionId":"test-id"}` + "\n"
		lite := &liteSessionFile{
			mtime: 1705312800000,
			size:  int64(len(tail)),
			head:  tail,
			tail:  tail,
		}

		info := parseSessionInfoFromLite("test-id", lite, nil)
		if info == nil {
			t.Fatal("Expected non-nil info")
		}
		if info.Tag == nil || *info.Tag != "experiment" {
			t.Errorf("Tag = %v, want %q", info.Tag, "experiment")
		}
	})

	t.Run("no_summary_returns_nil", func(t *testing.T) {
		head := `{"type":"system","message":"init"}`
		lite := &liteSessionFile{
			mtime: 1705312800000,
			size:  int64(len(head)),
			head:  head,
			tail:  head,
		}

		info := parseSessionInfoFromLite("test-id", lite, nil)
		if info != nil {
			t.Error("Expected nil for session with no summary")
		}
	})

	t.Run("project_path_fallback_for_cwd", func(t *testing.T) {
		head := `{"type":"user","message":{"role":"user","content":"Hello"}}` + "\n"
		lite := &liteSessionFile{
			mtime: 1705312800000,
			size:  int64(len(head)),
			head:  head,
			tail:  head,
		}

		projectPath := "/my/project"
		info := parseSessionInfoFromLite("test-id", lite, &projectPath)
		if info == nil {
			t.Fatal("Expected non-nil info")
		}
		if info.Cwd == nil || *info.Cwd != "/my/project" {
			t.Errorf("Cwd = %v, want %q", info.Cwd, "/my/project")
		}
	})
}

// --- Transcript Parsing Tests ---

// TestParseTranscriptEntries tests JSONL transcript parsing.
func TestParseTranscriptEntries(t *testing.T) {
	content := `{"type":"user","uuid":"msg-1","message":{"role":"user","content":"Hello"}}
{"type":"assistant","uuid":"msg-2","parentUuid":"msg-1","message":{"role":"assistant","content":"Hi"}}
{"type":"summary","summary":"test"}
{"type":"progress","uuid":"msg-3","parentUuid":"msg-2"}
invalid json line
`
	entries := parseTranscriptEntries(content)

	// Should have 3 entries (user, assistant, progress - summary has no uuid)
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Verify types.
	types := make([]string, len(entries))
	for i, e := range entries {
		types[i], _ = e["type"].(string)
	}
	wantTypes := []string{"user", "assistant", "progress"}
	for i, want := range wantTypes {
		if types[i] != want {
			t.Errorf("Entry %d type = %q, want %q", i, types[i], want)
		}
	}
}

// TestBuildConversationChain tests the conversation chain builder.
func TestBuildConversationChain(t *testing.T) {
	t.Run("simple_chain", func(t *testing.T) {
		entries := []transcriptEntry{
			{"type": "user", "uuid": "msg-1"},
			{"type": "assistant", "uuid": "msg-2", "parentUuid": "msg-1"},
			{"type": "user", "uuid": "msg-3", "parentUuid": "msg-2"},
			{"type": "assistant", "uuid": "msg-4", "parentUuid": "msg-3"},
		}

		chain := buildConversationChain(entries)
		if len(chain) != 4 {
			t.Fatalf("Expected 4 entries in chain, got %d", len(chain))
		}

		// Should be in chronological order.
		for i, e := range chain {
			uuid, _ := e["uuid"].(string)
			want := fmt.Sprintf("msg-%d", i+1)
			if uuid != want {
				t.Errorf("Chain[%d].uuid = %q, want %q", i, uuid, want)
			}
		}
	})

	t.Run("empty_entries", func(t *testing.T) {
		chain := buildConversationChain(nil)
		if chain != nil {
			t.Errorf("Expected nil for empty entries, got %v", chain)
		}
	})

	t.Run("skips_sidechain_for_leaf_pick", func(t *testing.T) {
		entries := []transcriptEntry{
			{"type": "user", "uuid": "msg-1"},
			{"type": "assistant", "uuid": "msg-2", "parentUuid": "msg-1"},
			{"type": "user", "uuid": "msg-3", "parentUuid": "msg-2", "isSidechain": true},
		}

		chain := buildConversationChain(entries)
		// The leaf should be msg-2 (not the sidechain msg-3).
		if len(chain) < 1 {
			t.Fatal("Expected non-empty chain")
		}
		lastUUID, _ := chain[len(chain)-1]["uuid"].(string)
		if lastUUID != "msg-2" {
			t.Errorf("Last chain entry uuid = %q, want %q", lastUUID, "msg-2")
		}
	})
}

// --- Visible Message Tests ---

// TestIsVisibleMessage tests message visibility filtering.
func TestIsVisibleMessage(t *testing.T) {
	tests := []struct {
		name  string
		entry transcriptEntry
		want  bool
	}{
		{"user_visible", transcriptEntry{"type": "user"}, true},
		{"assistant_visible", transcriptEntry{"type": "assistant"}, true},
		{"progress_hidden", transcriptEntry{"type": "progress"}, false},
		{"system_hidden", transcriptEntry{"type": "system"}, false},
		{"meta_hidden", transcriptEntry{"type": "user", "isMeta": true}, false},
		{"sidechain_hidden", transcriptEntry{"type": "assistant", "isSidechain": true}, false},
		{"team_hidden", transcriptEntry{"type": "user", "teamName": "team1"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVisibleMessage(tt.entry)
			if got != tt.want {
				t.Errorf("isVisibleMessage(%v) = %v, want %v", tt.entry, got, tt.want)
			}
		})
	}
}

// --- ToSessionMessage Tests ---

// TestToSessionMessage tests conversion from transcript entry to SessionMessage.
func TestToSessionMessage(t *testing.T) {
	t.Run("user_message", func(t *testing.T) {
		entry := transcriptEntry{
			"type":      "user",
			"uuid":      "msg-1",
			"sessionId": "session-1",
			"message":   map[string]any{"role": "user", "content": "Hello"},
		}

		msg := toSessionMessage(entry)
		if msg.Type != "user" {
			t.Errorf("Type = %q, want %q", msg.Type, "user")
		}
		if msg.UUID != "msg-1" {
			t.Errorf("UUID = %q, want %q", msg.UUID, "msg-1")
		}
		if msg.SessionID != "session-1" {
			t.Errorf("SessionID = %q, want %q", msg.SessionID, "session-1")
		}
		if msg.ParentToolUseID != nil {
			t.Errorf("ParentToolUseID = %v, want nil", msg.ParentToolUseID)
		}
	})

	t.Run("assistant_message", func(t *testing.T) {
		entry := transcriptEntry{
			"type": "assistant",
			"uuid": "msg-2",
		}

		msg := toSessionMessage(entry)
		if msg.Type != "assistant" {
			t.Errorf("Type = %q, want %q", msg.Type, "assistant")
		}
	})
}

// --- End-to-End Tests (with tmp filesystem) ---

// TestListSessionsEndToEnd tests the full ListSessions flow with temp directories.
func TestListSessionsEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	// Create projects dir.
	projectsDir := filepath.Join(tmpDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a project directory.
	projectDir := filepath.Join(projectsDir, sanitizePath("/test/project"))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create session files.
	sid1, _ := createTestSession(t, projectDir, "550e8400-e29b-41d4-a716-446655440000",
		"First session prompt", nil, nil)
	sid2, _ := createTestSession(t, projectDir, "660e8400-e29b-41d4-a716-446655440000",
		"Second session prompt", nil, nil)

	// Touch the second file to be newer.
	now := time.Now()
	if err := os.Chtimes(
		filepath.Join(projectDir, sid2+".jsonl"),
		now, now,
	); err != nil {
		t.Fatal(err)
	}

	t.Run("list_all", func(t *testing.T) {
		sessions, err := ListSessions(nil)
		if err != nil {
			t.Fatalf("ListSessions error: %v", err)
		}
		if len(sessions) != 2 {
			t.Fatalf("Expected 2 sessions, got %d", len(sessions))
		}
		// Should be sorted by LastModified descending.
		if sessions[0].LastModified < sessions[1].LastModified {
			t.Error("Expected sessions sorted by LastModified descending")
		}
	})

	t.Run("list_with_dir", func(t *testing.T) {
		dir := "/test/project"
		sessions, err := ListSessions(&ListSessionsOptions{Dir: &dir})
		if err != nil {
			t.Fatalf("ListSessions error: %v", err)
		}
		if len(sessions) != 2 {
			t.Fatalf("Expected 2 sessions, got %d", len(sessions))
		}
	})

	t.Run("list_with_limit", func(t *testing.T) {
		limit := 1
		sessions, err := ListSessions(&ListSessionsOptions{Limit: &limit})
		if err != nil {
			t.Fatalf("ListSessions error: %v", err)
		}
		if len(sessions) != 1 {
			t.Fatalf("Expected 1 session, got %d", len(sessions))
		}
	})

	t.Run("list_with_offset", func(t *testing.T) {
		offset := 1
		sessions, err := ListSessions(&ListSessionsOptions{Offset: &offset})
		if err != nil {
			t.Fatalf("ListSessions error: %v", err)
		}
		if len(sessions) != 1 {
			t.Fatalf("Expected 1 session, got %d", len(sessions))
		}
	})

	t.Run("list_nonexistent_dir", func(t *testing.T) {
		dir := "/nonexistent/project"
		sessions, err := ListSessions(&ListSessionsOptions{Dir: &dir})
		if err != nil {
			t.Fatalf("ListSessions error: %v", err)
		}
		if len(sessions) != 0 {
			t.Errorf("Expected 0 sessions for nonexistent dir, got %d", len(sessions))
		}
	})

	_ = sid1
}

// TestGetSessionInfoEndToEnd tests GetSessionInfo with temp directories.
func TestGetSessionInfoEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	projectsDir := filepath.Join(tmpDir, "projects")
	projectDir := filepath.Join(projectsDir, sanitizePath("/test/project"))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	title := "My Custom Title"
	sid, _ := createTestSession(t, projectDir, "550e8400-e29b-41d4-a716-446655440000",
		"Hello Claude", &title, nil)

	t.Run("found_no_dir", func(t *testing.T) {
		info, err := GetSessionInfo(sid, nil)
		if err != nil {
			t.Fatalf("GetSessionInfo error: %v", err)
		}
		if info == nil {
			t.Fatal("Expected non-nil info")
		}
		if info.SessionID != sid {
			t.Errorf("SessionID = %q, want %q", info.SessionID, sid)
		}
		if info.CustomTitle == nil || *info.CustomTitle != title {
			t.Errorf("CustomTitle = %v, want %q", info.CustomTitle, title)
		}
	})

	t.Run("found_with_dir", func(t *testing.T) {
		dir := "/test/project"
		info, err := GetSessionInfo(sid, &GetSessionInfoOptions{Dir: &dir})
		if err != nil {
			t.Fatalf("GetSessionInfo error: %v", err)
		}
		if info == nil {
			t.Fatal("Expected non-nil info")
		}
	})

	t.Run("not_found", func(t *testing.T) {
		info, err := GetSessionInfo("00000000-0000-0000-0000-000000000000", nil)
		if err != nil {
			t.Fatalf("GetSessionInfo error: %v", err)
		}
		if info != nil {
			t.Error("Expected nil for missing session")
		}
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		info, err := GetSessionInfo("not-a-uuid", nil)
		if err != nil {
			t.Fatalf("GetSessionInfo error: %v", err)
		}
		if info != nil {
			t.Error("Expected nil for invalid UUID")
		}
	})
}

// TestGetSessionMessagesEndToEnd tests GetSessionMessages with temp directories.
func TestGetSessionMessagesEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	projectsDir := filepath.Join(tmpDir, "projects")
	projectDir := filepath.Join(projectsDir, sanitizePath("/test/project"))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	sid := "550e8400-e29b-41d4-a716-446655440000"
	content := `{"type":"user","uuid":"msg-1","sessionId":"` + sid + `","message":{"role":"user","content":"Hello"}}
{"type":"assistant","uuid":"msg-2","parentUuid":"msg-1","sessionId":"` + sid + `","message":{"role":"assistant","content":"Hi there!"}}
{"type":"user","uuid":"msg-3","parentUuid":"msg-2","sessionId":"` + sid + `","message":{"role":"user","content":"How are you?"}}
{"type":"assistant","uuid":"msg-4","parentUuid":"msg-3","sessionId":"` + sid + `","message":{"role":"assistant","content":"I am well!"}}
`
	if err := os.WriteFile(filepath.Join(projectDir, sid+".jsonl"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	t.Run("all_messages", func(t *testing.T) {
		messages, err := GetSessionMessages(sid, nil)
		if err != nil {
			t.Fatalf("GetSessionMessages error: %v", err)
		}
		if len(messages) != 4 {
			t.Fatalf("Expected 4 messages, got %d", len(messages))
		}
		if messages[0].Type != "user" {
			t.Errorf("First message type = %q, want %q", messages[0].Type, "user")
		}
		if messages[1].Type != "assistant" {
			t.Errorf("Second message type = %q, want %q", messages[1].Type, "assistant")
		}
	})

	t.Run("with_limit", func(t *testing.T) {
		limit := 2
		messages, err := GetSessionMessages(sid, &GetSessionMessagesOptions{Limit: &limit})
		if err != nil {
			t.Fatalf("GetSessionMessages error: %v", err)
		}
		if len(messages) != 2 {
			t.Fatalf("Expected 2 messages, got %d", len(messages))
		}
	})

	t.Run("with_offset", func(t *testing.T) {
		offset := 2
		messages, err := GetSessionMessages(sid, &GetSessionMessagesOptions{Offset: &offset})
		if err != nil {
			t.Fatalf("GetSessionMessages error: %v", err)
		}
		if len(messages) != 2 {
			t.Fatalf("Expected 2 messages, got %d", len(messages))
		}
		if messages[0].UUID != "msg-3" {
			t.Errorf("First message UUID = %q, want %q", messages[0].UUID, "msg-3")
		}
	})

	t.Run("with_limit_and_offset", func(t *testing.T) {
		limit := 1
		offset := 1
		messages, err := GetSessionMessages(sid, &GetSessionMessagesOptions{
			Limit:  &limit,
			Offset: &offset,
		})
		if err != nil {
			t.Fatalf("GetSessionMessages error: %v", err)
		}
		if len(messages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(messages))
		}
		if messages[0].UUID != "msg-2" {
			t.Errorf("Message UUID = %q, want %q", messages[0].UUID, "msg-2")
		}
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		messages, err := GetSessionMessages("not-a-uuid", nil)
		if err != nil {
			t.Fatalf("GetSessionMessages error: %v", err)
		}
		if messages != nil {
			t.Error("Expected nil for invalid UUID")
		}
	})

	t.Run("not_found", func(t *testing.T) {
		messages, err := GetSessionMessages("00000000-0000-0000-0000-000000000000", nil)
		if err != nil {
			t.Fatalf("GetSessionMessages error: %v", err)
		}
		if messages != nil {
			t.Error("Expected nil for missing session")
		}
	})
}

// --- Mutation Tests ---

// TestRenameSession tests session renaming.
func TestRenameSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	projectsDir := filepath.Join(tmpDir, "projects")
	projectDir := filepath.Join(projectsDir, sanitizePath("/test/project"))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	sid, path := createTestSession(t, projectDir, "550e8400-e29b-41d4-a716-446655440000",
		"Hello Claude", nil, nil)

	t.Run("rename_success", func(t *testing.T) {
		err := RenameSession(sid, "New Title", nil)
		if err != nil {
			t.Fatalf("RenameSession error: %v", err)
		}

		// Verify the file was appended.
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		if !strings.Contains(content, `"customTitle":"New Title"`) && !strings.Contains(content, `"customTitle": "New Title"`) {
			t.Errorf("Expected custom title in file, got: %s", content)
		}
	})

	t.Run("rename_trims_whitespace", func(t *testing.T) {
		err := RenameSession(sid, "  Trimmed Title  ", nil)
		if err != nil {
			t.Fatalf("RenameSession error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"customTitle":"Trimmed Title"`) {
			t.Error("Expected trimmed title in file")
		}
	})

	t.Run("rename_empty_title", func(t *testing.T) {
		err := RenameSession(sid, "", nil)
		if err == nil {
			t.Error("Expected error for empty title")
		}
	})

	t.Run("rename_whitespace_title", func(t *testing.T) {
		err := RenameSession(sid, "   ", nil)
		if err == nil {
			t.Error("Expected error for whitespace-only title")
		}
	})

	t.Run("rename_invalid_uuid", func(t *testing.T) {
		err := RenameSession("not-a-uuid", "Title", nil)
		if err == nil {
			t.Error("Expected error for invalid UUID")
		}
	})

	t.Run("rename_with_dir", func(t *testing.T) {
		dir := "/test/project"
		err := RenameSession(sid, "Dir Title", &SessionMutationOptions{Dir: &dir})
		if err != nil {
			t.Fatalf("RenameSession with dir error: %v", err)
		}
	})

	t.Run("rename_nonexistent_session", func(t *testing.T) {
		err := RenameSession("00000000-0000-0000-0000-000000000000", "Title", nil)
		if err == nil {
			t.Error("Expected error for nonexistent session")
		}
	})
}

// TestTagSession tests session tagging.
func TestTagSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	projectsDir := filepath.Join(tmpDir, "projects")
	projectDir := filepath.Join(projectsDir, sanitizePath("/test/project"))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	sid, path := createTestSession(t, projectDir, "550e8400-e29b-41d4-a716-446655440000",
		"Hello Claude", nil, nil)

	t.Run("tag_success", func(t *testing.T) {
		tag := "experiment"
		err := TagSession(sid, &tag, nil)
		if err != nil {
			t.Fatalf("TagSession error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"tag":"experiment"`) && !strings.Contains(string(data), `"tag": "experiment"`) {
			t.Errorf("Expected tag in file, got: %s", string(data))
		}
	})

	t.Run("clear_tag", func(t *testing.T) {
		err := TagSession(sid, nil, nil)
		if err != nil {
			t.Fatalf("TagSession error: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// Clearing should write empty string tag.
		if !strings.Contains(string(data), `"tag":""`) && !strings.Contains(string(data), `"tag": ""`) {
			t.Errorf("Expected empty tag in file, got: %s", string(data))
		}
	})

	t.Run("tag_empty_string", func(t *testing.T) {
		tag := ""
		err := TagSession(sid, &tag, nil)
		if err == nil {
			t.Error("Expected error for empty tag string")
		}
	})

	t.Run("tag_whitespace_only", func(t *testing.T) {
		tag := "   "
		err := TagSession(sid, &tag, nil)
		if err == nil {
			t.Error("Expected error for whitespace-only tag")
		}
	})

	t.Run("tag_invalid_uuid", func(t *testing.T) {
		tag := "test"
		err := TagSession("not-a-uuid", &tag, nil)
		if err == nil {
			t.Error("Expected error for invalid UUID")
		}
	})
}

// --- Fork Tests ---

// TestForkSession tests session forking.
func TestForkSession(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	projectsDir := filepath.Join(tmpDir, "projects")
	projectDir := filepath.Join(projectsDir, sanitizePath("/test/project"))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	sid := "550e8400-e29b-41d4-a716-446655440000"
	content := `{"type":"user","uuid":"msg-1","sessionId":"` + sid + `","message":{"role":"user","content":"Hello"}}
{"type":"assistant","uuid":"msg-2","parentUuid":"msg-1","sessionId":"` + sid + `","message":{"role":"assistant","content":"Hi!"}}
{"type":"user","uuid":"msg-3","parentUuid":"msg-2","sessionId":"` + sid + `","message":{"role":"user","content":"More"}}
{"type":"assistant","uuid":"msg-4","parentUuid":"msg-3","sessionId":"` + sid + `","message":{"role":"assistant","content":"Done"}}
`
	if err := os.WriteFile(filepath.Join(projectDir, sid+".jsonl"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	t.Run("fork_full", func(t *testing.T) {
		result, err := ForkSession(sid, nil)
		if err != nil {
			t.Fatalf("ForkSession error: %v", err)
		}
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		if !isValidUUID(result.SessionID) {
			t.Errorf("Expected valid UUID, got %q", result.SessionID)
		}
		if result.SessionID == sid {
			t.Error("Fork should have a different session ID")
		}

		// Verify the forked file exists.
		forkedPath := filepath.Join(projectDir, result.SessionID+".jsonl")
		if _, err := os.Stat(forkedPath); os.IsNotExist(err) {
			t.Error("Forked session file does not exist")
		}

		// Verify content was copied.
		forkedData, err := os.ReadFile(forkedPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(forkedData) != content {
			t.Error("Forked content should match original")
		}
	})

	t.Run("fork_with_truncation", func(t *testing.T) {
		upTo := "msg-2"
		result, err := ForkSession(sid, &ForkSessionOptions{UpToMessageID: &upTo})
		if err != nil {
			t.Fatalf("ForkSession error: %v", err)
		}

		forkedData, err := os.ReadFile(filepath.Join(projectDir, result.SessionID+".jsonl"))
		if err != nil {
			t.Fatal(err)
		}

		// Should only contain lines up to msg-2.
		lines := strings.Split(strings.TrimSpace(string(forkedData)), "\n")
		if len(lines) != 2 {
			t.Errorf("Expected 2 lines after truncation, got %d", len(lines))
		}
	})

	t.Run("fork_with_title", func(t *testing.T) {
		title := "My Fork"
		result, err := ForkSession(sid, &ForkSessionOptions{Title: &title})
		if err != nil {
			t.Fatalf("ForkSession error: %v", err)
		}

		forkedData, err := os.ReadFile(filepath.Join(projectDir, result.SessionID+".jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(forkedData), `"customTitle":"My Fork"`) {
			t.Error("Expected custom title in forked file")
		}
	})

	t.Run("fork_invalid_uuid", func(t *testing.T) {
		_, err := ForkSession("not-a-uuid", nil)
		if err == nil {
			t.Error("Expected error for invalid UUID")
		}
	})

	t.Run("fork_nonexistent", func(t *testing.T) {
		_, err := ForkSession("00000000-0000-0000-0000-000000000000", nil)
		if err == nil {
			t.Error("Expected error for nonexistent session")
		}
	})
}

// --- TryAppend Tests ---

// TestTryAppend tests the low-level append helper.
func TestTryAppend(t *testing.T) {
	t.Run("appends_to_existing", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.jsonl")
		if err := os.WriteFile(path, []byte("line1\n"), 0600); err != nil {
			t.Fatal(err)
		}

		result := tryAppend(path, []byte("line2\n"))
		if !result {
			t.Error("Expected tryAppend to return true")
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "line1\nline2\n" {
			t.Errorf("File content = %q, want %q", string(data), "line1\nline2\n")
		}
	})

	t.Run("returns_false_for_missing", func(t *testing.T) {
		result := tryAppend("/nonexistent/file.jsonl", []byte("data\n"))
		if result {
			t.Error("Expected tryAppend to return false for missing file")
		}
	})

	t.Run("returns_false_for_empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "empty.jsonl")
		if err := os.WriteFile(path, []byte{}, 0600); err != nil {
			t.Fatal(err)
		}

		result := tryAppend(path, []byte("data\n"))
		if result {
			t.Error("Expected tryAppend to return false for empty file")
		}
	})
}

// --- Truncation Tests ---

// TestTruncateAtMessage tests JSONL truncation at a message UUID.
func TestTruncateAtMessage(t *testing.T) {
	content := `{"type":"user","uuid":"msg-1","message":"hello"}
{"type":"assistant","uuid":"msg-2","parentUuid":"msg-1","message":"hi"}
{"type":"user","uuid":"msg-3","parentUuid":"msg-2","message":"more"}
{"type":"assistant","uuid":"msg-4","parentUuid":"msg-3","message":"done"}
`

	t.Run("truncate_at_msg_2", func(t *testing.T) {
		result := truncateAtMessage(content, "msg-2")
		lines := strings.Split(strings.TrimSpace(result), "\n")
		if len(lines) != 2 {
			t.Errorf("Expected 2 lines, got %d: %v", len(lines), lines)
		}
	})

	t.Run("truncate_at_msg_1", func(t *testing.T) {
		result := truncateAtMessage(content, "msg-1")
		lines := strings.Split(strings.TrimSpace(result), "\n")
		if len(lines) != 1 {
			t.Errorf("Expected 1 line, got %d: %v", len(lines), lines)
		}
	})

	t.Run("truncate_at_nonexistent", func(t *testing.T) {
		result := truncateAtMessage(content, "msg-999")
		// Should return all content since the message was not found.
		if result != content {
			t.Error("Expected full content when message ID not found")
		}
	})
}

// --- Pagination Tests ---

// TestApplySortAndPagination tests the sorting and pagination helper.
func TestApplySortAndPagination(t *testing.T) {
	sessions := []SDKSessionInfo{
		{SessionID: "a", LastModified: 100},
		{SessionID: "b", LastModified: 300},
		{SessionID: "c", LastModified: 200},
	}

	t.Run("sort_only", func(t *testing.T) {
		result := applySortAndPagination(cloneSessions(sessions), nil, nil)
		if len(result) != 3 {
			t.Fatalf("Expected 3 sessions, got %d", len(result))
		}
		if result[0].SessionID != "b" || result[1].SessionID != "c" || result[2].SessionID != "a" {
			t.Errorf("Wrong sort order: %v", sessionIDs(result))
		}
	})

	t.Run("with_limit", func(t *testing.T) {
		limit := 2
		result := applySortAndPagination(cloneSessions(sessions), &limit, nil)
		if len(result) != 2 {
			t.Fatalf("Expected 2 sessions, got %d", len(result))
		}
		if result[0].SessionID != "b" {
			t.Errorf("First session = %q, want %q", result[0].SessionID, "b")
		}
	})

	t.Run("with_offset", func(t *testing.T) {
		offset := 1
		result := applySortAndPagination(cloneSessions(sessions), nil, &offset)
		if len(result) != 2 {
			t.Fatalf("Expected 2 sessions, got %d", len(result))
		}
		if result[0].SessionID != "c" {
			t.Errorf("First session = %q, want %q", result[0].SessionID, "c")
		}
	})

	t.Run("with_limit_and_offset", func(t *testing.T) {
		limit := 1
		offset := 1
		result := applySortAndPagination(cloneSessions(sessions), &limit, &offset)
		if len(result) != 1 {
			t.Fatalf("Expected 1 session, got %d", len(result))
		}
		if result[0].SessionID != "c" {
			t.Errorf("Session = %q, want %q", result[0].SessionID, "c")
		}
	})

	t.Run("offset_beyond_length", func(t *testing.T) {
		offset := 10
		result := applySortAndPagination(cloneSessions(sessions), nil, &offset)
		if result != nil {
			t.Errorf("Expected nil for offset beyond length, got %v", result)
		}
	})
}

// --- Deduplication Tests ---

// TestDeduplicateBySessionID tests session deduplication.
func TestDeduplicateBySessionID(t *testing.T) {
	sessions := []SDKSessionInfo{
		{SessionID: "a", LastModified: 100, Summary: "old"},
		{SessionID: "a", LastModified: 200, Summary: "new"},
		{SessionID: "b", LastModified: 150, Summary: "only"},
	}

	result := deduplicateBySessionID(sessions)
	if len(result) != 2 {
		t.Fatalf("Expected 2 sessions, got %d", len(result))
	}

	// Find session "a" and verify the newer one was kept.
	for _, s := range result {
		if s.SessionID == "a" {
			if s.Summary != "new" {
				t.Errorf("Expected newer session 'a', got Summary = %q", s.Summary)
			}
			if s.LastModified != 200 {
				t.Errorf("Expected LastModified 200, got %d", s.LastModified)
			}
		}
	}
}

// --- Generate Session ID Tests ---

// TestGenerateSessionID tests UUID generation.
func TestGenerateSessionID(t *testing.T) {
	id := generateSessionID()
	if !isValidUUID(id) {
		t.Errorf("generateSessionID() = %q, not a valid UUID", id)
	}

	// Ensure uniqueness.
	id2 := generateSessionID()
	if id == id2 {
		t.Error("Expected unique session IDs")
	}
}

// --- FirstNonEmpty Tests ---

// TestFirstNonEmpty tests the firstNonEmpty helper.
func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{"first_non_empty", []string{"", "", "hello", "world"}, "hello"},
		{"first_is_value", []string{"first", "second"}, "first"},
		{"all_empty", []string{"", "", ""}, ""},
		{"single_value", []string{"only"}, "only"},
		{"no_values", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstNonEmpty(tt.values...)
			if got != tt.want {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}

// --- Config Dir Tests ---

// TestGetClaudeConfigDir tests config directory resolution.
func TestGetClaudeConfigDir(t *testing.T) {
	t.Run("uses_env_var", func(t *testing.T) {
		t.Setenv("CLAUDE_CONFIG_DIR", "/custom/config")
		got := getClaudeConfigDir()
		if got != "/custom/config" {
			t.Errorf("getClaudeConfigDir() = %q, want %q", got, "/custom/config")
		}
	})

	t.Run("defaults_to_home", func(t *testing.T) {
		t.Setenv("CLAUDE_CONFIG_DIR", "")
		got := getClaudeConfigDir()
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Cannot determine home directory")
		}
		want := filepath.Join(home, ".claude")
		if got != want {
			t.Errorf("getClaudeConfigDir() = %q, want %q", got, want)
		}
	})
}

// =============================================================================
// Helper Functions (utilities)
// =============================================================================

// createTestSession creates a test session JSONL file.
// Returns (sessionID, filePath).
func createTestSession(t *testing.T, projectDir, sessionID, prompt string, customTitle, tag *string) (string, string) {
	t.Helper()

	path := filepath.Join(projectDir, sessionID+".jsonl")
	var lines []string

	userEntry := map[string]any{
		"type":    "user",
		"uuid":    "user-msg-1",
		"message": map[string]any{"role": "user", "content": prompt},
	}
	userJSON, _ := json.Marshal(userEntry)
	lines = append(lines, string(userJSON))

	assistantEntry := map[string]any{
		"type":       "assistant",
		"uuid":       "asst-msg-1",
		"parentUuid": "user-msg-1",
		"message":    map[string]any{"role": "assistant", "content": "Hi there!"},
	}
	assistantJSON, _ := json.Marshal(assistantEntry)
	lines = append(lines, string(assistantJSON))

	if customTitle != nil {
		titleEntry := map[string]string{
			"type":        "custom-title",
			"customTitle": *customTitle,
			"sessionId":   sessionID,
		}
		titleJSON, _ := json.Marshal(titleEntry)
		lines = append(lines, string(titleJSON))
	}

	if tag != nil {
		tagEntry := map[string]string{
			"type":      "tag",
			"tag":       *tag,
			"sessionId": sessionID,
		}
		tagJSON, _ := json.Marshal(tagEntry)
		lines = append(lines, string(tagJSON))
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	return sessionID, path
}

// cloneSessions creates a copy of a session slice.
func cloneSessions(sessions []SDKSessionInfo) []SDKSessionInfo {
	result := make([]SDKSessionInfo, len(sessions))
	copy(result, sessions)
	return result
}

// sessionIDs extracts session IDs from a slice for debug output.
func sessionIDs(sessions []SDKSessionInfo) []string {
	ids := make([]string, len(sessions))
	for i, s := range sessions {
		ids[i] = s.SessionID
	}
	return ids
}
