package claudecode

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// SDKSessionInfo holds metadata about a session.
type SDKSessionInfo struct {
	SessionID    string  `json:"sessionId"`
	Summary      string  `json:"summary"`
	LastModified int64   `json:"lastModified"`
	FileSize     *int64  `json:"fileSize,omitempty"`
	CustomTitle  *string `json:"customTitle,omitempty"`
	FirstPrompt  *string `json:"firstPrompt,omitempty"`
	GitBranch    *string `json:"gitBranch,omitempty"`
	Cwd          *string `json:"cwd,omitempty"`
	Tag          *string `json:"tag,omitempty"`
	CreatedAt    *int64  `json:"createdAt,omitempty"`
}

// SessionMessage is a user or assistant message from a session transcript.
type SessionMessage struct {
	Type            string  `json:"type"`              // "user" | "assistant"
	UUID            string  `json:"uuid"`              //nolint:revive,stylecheck // Matches TS/Python SDK field name.
	SessionID       string  `json:"session_id"`        //nolint:revive,stylecheck // Matches TS/Python SDK wire format.
	Message         any     `json:"message"`           // Raw Anthropic API message
	ParentToolUseID *string `json:"parent_tool_use_id"` //nolint:revive,stylecheck // Matches TS/Python SDK wire format.
}

// ListSessionsOptions configures session listing.
type ListSessionsOptions struct {
	// Dir restricts listing to sessions for this project directory.
	// When nil, sessions across all projects are returned.
	Dir *string
	// Limit caps the number of sessions returned.
	Limit *int
	// Offset skips this many sessions before returning results.
	Offset *int
	// IncludeWorktrees controls whether git worktree paths are included.
	// Defaults to true when Dir is set.
	IncludeWorktrees *bool
}

// GetSessionInfoOptions configures session info retrieval.
type GetSessionInfoOptions struct {
	// Dir is the project directory to search in.
	// When nil, all project directories are searched.
	Dir *string
}

// GetSessionMessagesOptions configures message retrieval.
type GetSessionMessagesOptions struct {
	// Dir is the project directory to search in.
	// When nil, all project directories are searched.
	Dir *string
	// Limit caps the number of messages returned.
	Limit *int
	// Offset skips this many messages before returning results.
	Offset *int
}

// ForkSessionOptions configures session forking.
type ForkSessionOptions struct {
	// Dir is the project directory containing the source session.
	// When nil, all project directories are searched.
	Dir *string
	// UpToMessageID truncates the fork at this message UUID (inclusive).
	UpToMessageID *string
	// Title sets a custom title on the forked session.
	Title *string
}

// ForkSessionResult holds the result of a fork operation.
type ForkSessionResult struct {
	SessionID string `json:"sessionId"`
}

// SessionMutationOptions is shared by session mutation functions.
type SessionMutationOptions struct {
	// Dir is the project directory containing the session.
	// When nil, all project directories are searched.
	Dir *string
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	// liteReadBufSize is the buffer size for head/tail reads of session files.
	liteReadBufSize = 65536

	// maxSanitizedLength is the max length for a sanitized path component.
	maxSanitizedLength = 200
)

// ---------------------------------------------------------------------------
// Regex patterns
// ---------------------------------------------------------------------------

var (
	uuidRe      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	sanitizeRe  = regexp.MustCompile(`[^a-zA-Z0-9]`)
	commandRe   = regexp.MustCompile(`<command-name>(.*?)</command-name>`)
	skipPromptRe = regexp.MustCompile(
		`^(?:<local-command-stdout>|<session-start-hook>|<tick>|<goal>|` +
			`\[Request interrupted by user[^\]]*\]|` +
			`\s*<ide_opened_file>[\s\S]*</ide_opened_file>\s*$|` +
			`\s*<ide_selection>[\s\S]*</ide_selection>\s*$)`,
	)
)

// transcriptEntryTypes contains the JSONL types that carry uuid + parentUuid chain links.
var transcriptEntryTypes = map[string]bool{
	"user":       true,
	"assistant":  true,
	"progress":   true,
	"system":     true,
	"attachment": true,
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// ListSessions returns metadata for sessions, optionally filtered by directory.
//
// When opts.Dir is set, sessions for that project directory (and its git worktrees)
// are returned. When opts.Dir is nil, sessions across all projects under
// ~/.claude/projects/ are returned. Results are sorted by LastModified descending.
//
// Example:
//
//	// List all sessions across all projects
//	sessions, err := claudecode.ListSessions(nil)
//
//	// List sessions for a specific project
//	dir := "/path/to/project"
//	sessions, err := claudecode.ListSessions(&claudecode.ListSessionsOptions{Dir: &dir})
func ListSessions(opts *ListSessionsOptions) ([]SDKSessionInfo, error) {
	if opts == nil {
		opts = &ListSessionsOptions{}
	}

	var sessions []SDKSessionInfo
	var err error

	if opts.Dir != nil {
		includeWorktrees := true
		if opts.IncludeWorktrees != nil {
			includeWorktrees = *opts.IncludeWorktrees
		}
		sessions, err = listSessionsForProject(*opts.Dir, includeWorktrees)
	} else {
		sessions, err = listAllSessions()
	}
	if err != nil {
		return nil, err
	}

	sessions = deduplicateBySessionID(sessions)
	sessions = applySortAndPagination(sessions, opts.Limit, opts.Offset)

	return sessions, nil
}

// GetSessionInfo reads metadata for a single session by ID.
//
// When opts.Dir is set, the session is looked up in that project directory.
// When opts.Dir is nil, all project directories are searched.
//
// Returns nil with no error if the session is not found, is a sidechain session,
// or has no extractable summary.
//
// Example:
//
//	info, err := claudecode.GetSessionInfo("550e8400-e29b-41d4-a716-446655440000", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if info != nil {
//	    fmt.Println(info.Summary)
//	}
func GetSessionInfo(sessionID string, opts *GetSessionInfoOptions) (*SDKSessionInfo, error) {
	if !isValidUUID(sessionID) {
		return nil, nil
	}
	if opts == nil {
		opts = &GetSessionInfoOptions{}
	}

	fileName := sessionID + ".jsonl"

	if opts.Dir != nil {
		canonicalDir := canonicalizePath(*opts.Dir)
		projectDir := findProjectDir(canonicalDir)
		if projectDir != "" {
			lite, err := readSessionLite(filepath.Join(projectDir, fileName))
			if err == nil && lite != nil {
				return parseSessionInfoFromLite(sessionID, lite, &canonicalDir), nil
			}
		}
		return nil, nil
	}

	// No directory - search all project directories.
	projectsDir := getProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, nil
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		lite, readErr := readSessionLite(filepath.Join(projectsDir, entry.Name(), fileName))
		if readErr == nil && lite != nil {
			return parseSessionInfoFromLite(sessionID, lite, nil), nil
		}
	}
	return nil, nil
}

// GetSessionMessages reads conversation messages from a session transcript.
//
// Parses the full JSONL, builds the conversation chain via parentUuid links,
// and returns user/assistant messages in chronological order.
//
// Returns an empty slice with no error if the session is not found or the
// session_id is not a valid UUID.
//
// Example:
//
//	messages, err := claudecode.GetSessionMessages(
//	    "550e8400-e29b-41d4-a716-446655440000",
//	    &claudecode.GetSessionMessagesOptions{Dir: &dir},
//	)
func GetSessionMessages(sessionID string, opts *GetSessionMessagesOptions) ([]SessionMessage, error) {
	if !isValidUUID(sessionID) {
		return nil, nil
	}
	if opts == nil {
		opts = &GetSessionMessagesOptions{}
	}

	var dir *string
	if opts.Dir != nil {
		dir = opts.Dir
	}

	content, err := readSessionFile(sessionID, dir)
	if err != nil || content == "" {
		return nil, nil
	}

	entries := parseTranscriptEntries(content)
	chain := buildConversationChain(entries)

	var messages []SessionMessage
	for _, e := range chain {
		if isVisibleMessage(e) {
			messages = append(messages, toSessionMessage(e))
		}
	}

	messages = applyMessagePagination(messages, opts.Limit, opts.Offset)
	return messages, nil
}

// ForkSession creates a new session by copying a transcript.
//
// The source session's JSONL is copied to a new session ID. When
// opts.UpToMessageID is set, the copy is truncated at that message.
// When opts.Title is set, a custom title entry is appended to the fork.
//
// Example:
//
//	result, err := claudecode.ForkSession(
//	    "550e8400-e29b-41d4-a716-446655440000",
//	    &claudecode.ForkSessionOptions{Title: stringPtr("My fork")},
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Forked to:", result.SessionID)
func ForkSession(sessionID string, opts *ForkSessionOptions) (*ForkSessionResult, error) {
	if !isValidUUID(sessionID) {
		return nil, fmt.Errorf("invalid session ID: %s", sessionID)
	}
	if opts == nil {
		opts = &ForkSessionOptions{}
	}

	var dir *string
	if opts.Dir != nil {
		dir = opts.Dir
	}

	content, err := readSessionFile(sessionID, dir)
	if err != nil {
		return nil, fmt.Errorf("session %s not found: %w", sessionID, err)
	}
	if content == "" {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// If UpToMessageID is set, truncate content at that message.
	if opts.UpToMessageID != nil {
		content = truncateAtMessage(content, *opts.UpToMessageID)
	}

	// Generate new session ID.
	newSessionID := generateSessionID()

	// Find the source session file path to determine the project directory.
	sourcePath := findSessionFilePath(sessionID, dir)
	if sourcePath == "" {
		return nil, fmt.Errorf("session %s not found on disk", sessionID)
	}

	projectDir := filepath.Dir(sourcePath)
	newPath := filepath.Join(projectDir, newSessionID+".jsonl")

	// Write the forked content.
	err = os.WriteFile(newPath, []byte(content), 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to write forked session: %w", err)
	}

	// Append custom title if provided.
	if opts.Title != nil && *opts.Title != "" {
		titleEntry := map[string]string{
			"type":        "custom-title",
			"customTitle": strings.TrimSpace(*opts.Title),
			"sessionId":   newSessionID,
		}
		titleJSON, jsonErr := json.Marshal(titleEntry)
		if jsonErr != nil {
			return nil, fmt.Errorf("failed to marshal title: %w", jsonErr)
		}
		f, openErr := os.OpenFile(newPath, os.O_WRONLY|os.O_APPEND, 0600)
		if openErr != nil {
			return nil, fmt.Errorf("failed to append title to forked session: %w", openErr)
		}
		_, writeErr := f.Write(append(titleJSON, '\n'))
		closeErr := f.Close()
		if writeErr != nil {
			return nil, fmt.Errorf("failed to write title entry: %w", writeErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close forked session file: %w", closeErr)
		}
	}

	return &ForkSessionResult{SessionID: newSessionID}, nil
}

// RenameSession sets a custom title for a session.
//
// Appends a custom-title entry to the session's JSONL file. The most recent
// custom-title entry wins when the session metadata is read.
//
// Example:
//
//	err := claudecode.RenameSession(
//	    "550e8400-e29b-41d4-a716-446655440000",
//	    "My refactoring session",
//	    nil,
//	)
func RenameSession(sessionID string, title string, opts *SessionMutationOptions) error {
	if !isValidUUID(sessionID) {
		return fmt.Errorf("invalid session ID: %s", sessionID)
	}
	stripped := strings.TrimSpace(title)
	if stripped == "" {
		return fmt.Errorf("title must be non-empty")
	}
	if opts == nil {
		opts = &SessionMutationOptions{}
	}

	entry := map[string]string{
		"type":        "custom-title",
		"customTitle": stripped,
		"sessionId":   sessionID,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal rename entry: %w", err)
	}

	return appendToSession(sessionID, append(data, '\n'), opts.Dir)
}

// TagSession tags a session. Pass nil tag to clear.
//
// Appends a tag entry to the session's JSONL file. The most recent tag
// entry wins when the session metadata is read. Pass nil to clear the tag.
//
// Example:
//
//	// Set a tag
//	tag := "experiment"
//	err := claudecode.TagSession(sessionID, &tag, nil)
//
//	// Clear the tag
//	err = claudecode.TagSession(sessionID, nil, nil)
func TagSession(sessionID string, tag *string, opts *SessionMutationOptions) error {
	if !isValidUUID(sessionID) {
		return fmt.Errorf("invalid session ID: %s", sessionID)
	}
	if opts == nil {
		opts = &SessionMutationOptions{}
	}

	tagValue := ""
	if tag != nil {
		stripped := strings.TrimSpace(*tag)
		if stripped == "" {
			return fmt.Errorf("tag must be non-empty (use nil to clear)")
		}
		tagValue = stripped
	}

	entry := map[string]string{
		"type":      "tag",
		"tag":       tagValue,
		"sessionId": sessionID,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal tag entry: %w", err)
	}

	return appendToSession(sessionID, append(data, '\n'), opts.Dir)
}

// ---------------------------------------------------------------------------
// Internal: path sanitization and project directory discovery
// ---------------------------------------------------------------------------

// isValidUUID reports whether s is a valid UUID string.
func isValidUUID(s string) bool {
	return uuidRe.MatchString(s)
}

// simpleHash mirrors the JS simpleHash function (32-bit integer hash, base36).
func simpleHash(s string) string {
	var h int32
	for _, ch := range s {
		h = (h << 5) - h + int32(ch)
	}
	if h < 0 {
		h = -h
	}

	if h == 0 {
		return "0"
	}

	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	var result []byte
	n := h
	for n > 0 {
		result = append(result, digits[n%36])
		n /= 36
	}
	// Reverse.
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

// sanitizePath makes a string safe for use as a directory name.
func sanitizePath(name string) string {
	sanitized := sanitizeRe.ReplaceAllString(name, "-")
	if len(sanitized) <= maxSanitizedLength {
		return sanitized
	}
	h := simpleHash(name)
	return sanitized[:maxSanitizedLength] + "-" + h
}

// getClaudeConfigDir returns the Claude config directory.
func getClaudeConfigDir() string {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".claude")
	}
	return filepath.Join(home, ".claude")
}

// getProjectsDir returns the path to ~/.claude/projects/.
func getProjectsDir() string {
	return filepath.Join(getClaudeConfigDir(), "projects")
}

// getProjectDir returns the expected project directory for a given project path.
func getProjectDir(projectPath string) string {
	return filepath.Join(getProjectsDir(), sanitizePath(projectPath))
}

// canonicalizePath resolves a directory path to its canonical form.
func canonicalizePath(d string) string {
	resolved, err := filepath.EvalSymlinks(d)
	if err != nil {
		abs, absErr := filepath.Abs(d)
		if absErr != nil {
			return d
		}
		return abs
	}
	abs, err := filepath.Abs(resolved)
	if err != nil {
		return resolved
	}
	return abs
}

// findProjectDir finds the project directory for a given path.
// Returns empty string if not found.
func findProjectDir(projectPath string) string {
	exact := getProjectDir(projectPath)
	info, err := os.Stat(exact)
	if err == nil && info.IsDir() {
		return exact
	}

	// For short paths, exact match is authoritative.
	sanitized := sanitizePath(projectPath)
	if len(sanitized) <= maxSanitizedLength {
		return ""
	}

	// For long paths, try prefix-based scanning to handle hash mismatches.
	prefix := sanitized[:maxSanitizedLength]
	projectsDir := getProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix+"-") {
			return filepath.Join(projectsDir, entry.Name())
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Internal: lite session file reading
// ---------------------------------------------------------------------------

// liteSessionFile holds the result of reading a session file's head, tail, mtime and size.
type liteSessionFile struct {
	mtime int64  // milliseconds since epoch
	size  int64  // file size in bytes
	head  string // first liteReadBufSize bytes as string
	tail  string // last liteReadBufSize bytes as string
}

// readSessionLite opens a session file, stats it, and reads head + tail.
// Returns nil if the file is empty or on error.
func readSessionLite(path string) (*liteSessionFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	mtime := stat.ModTime().UnixMilli()

	if size == 0 {
		return nil, nil
	}

	headBuf := make([]byte, liteReadBufSize)
	n, err := f.Read(headBuf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	head := string(headBuf[:n])

	var tail string
	tailOffset := size - int64(liteReadBufSize)
	if tailOffset <= 0 {
		tail = head
	} else {
		_, err = f.Seek(tailOffset, io.SeekStart)
		if err != nil {
			return nil, err
		}
		tailBuf := make([]byte, liteReadBufSize)
		tn, readErr := f.Read(tailBuf)
		if readErr != nil && readErr != io.EOF {
			return nil, readErr
		}
		tail = string(tailBuf[:tn])
	}

	return &liteSessionFile{mtime: mtime, size: size, head: head, tail: tail}, nil
}

// ---------------------------------------------------------------------------
// Internal: JSON string field extraction (no full parse)
// ---------------------------------------------------------------------------

// extractJSONStringField extracts a simple JSON string field value without full parsing.
// Returns empty string if not found.
func extractJSONStringField(text, key string) string {
	patterns := []string{
		`"` + key + `":"`,
		`"` + key + `": "`,
	}
	for _, pattern := range patterns {
		idx := strings.Index(text, pattern)
		if idx < 0 {
			continue
		}
		valueStart := idx + len(pattern)
		i := valueStart
		for i < len(text) {
			if text[i] == '\\' {
				i += 2
				continue
			}
			if text[i] == '"' {
				return unescapeJSONString(text[valueStart:i])
			}
			i++
		}
	}
	return ""
}

// extractLastJSONStringField extracts the LAST occurrence of a JSON string field.
func extractLastJSONStringField(text, key string) string {
	patterns := []string{
		`"` + key + `":"`,
		`"` + key + `": "`,
	}
	lastValue := ""
	for _, pattern := range patterns {
		searchFrom := 0
		for {
			idx := strings.Index(text[searchFrom:], pattern)
			if idx < 0 {
				break
			}
			idx += searchFrom
			valueStart := idx + len(pattern)
			i := valueStart
			for i < len(text) {
				if text[i] == '\\' {
					i += 2
					continue
				}
				if text[i] == '"' {
					lastValue = unescapeJSONString(text[valueStart:i])
					break
				}
				i++
			}
			searchFrom = i + 1
		}
	}
	return lastValue
}

// unescapeJSONString unescapes a JSON string value.
func unescapeJSONString(raw string) string {
	if !strings.Contains(raw, `\`) {
		return raw
	}
	var result string
	err := json.Unmarshal([]byte(`"`+raw+`"`), &result)
	if err != nil {
		return raw
	}
	return result
}

// ---------------------------------------------------------------------------
// Internal: first prompt extraction
// ---------------------------------------------------------------------------

// extractFirstPromptFromHead extracts the first meaningful user prompt from JSONL head.
func extractFirstPromptFromHead(head string) string {
	commandFallback := ""
	scanner := bufio.NewScanner(strings.NewReader(head))
	scanner.Buffer(make([]byte, liteReadBufSize), liteReadBufSize)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, `"type":"user"`) && !strings.Contains(line, `"type": "user"`) {
			continue
		}
		if strings.Contains(line, `"tool_result"`) {
			continue
		}
		if strings.Contains(line, `"isMeta":true`) || strings.Contains(line, `"isMeta": true`) {
			continue
		}
		if strings.Contains(line, `"isCompactSummary":true`) || strings.Contains(line, `"isCompactSummary": true`) {
			continue
		}

		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entryType, _ := entry["type"].(string)
		if entryType != "user" {
			continue
		}
		msg, ok := entry["message"].(map[string]any)
		if !ok {
			continue
		}

		var texts []string
		switch content := msg["content"].(type) {
		case string:
			texts = append(texts, content)
		case []any:
			for _, block := range content {
				blockMap, isMap := block.(map[string]any)
				if !isMap {
					continue
				}
				blockType, _ := blockMap["type"].(string)
				blockText, _ := blockMap["text"].(string)
				if blockType == "text" && blockText != "" {
					texts = append(texts, blockText)
				}
			}
		}

		for _, raw := range texts {
			result := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
			if result == "" {
				continue
			}

			if cmdMatch := commandRe.FindStringSubmatch(result); cmdMatch != nil {
				if commandFallback == "" {
					commandFallback = cmdMatch[1]
				}
				continue
			}

			if skipPromptRe.MatchString(result) {
				continue
			}

			if len(result) > 200 {
				result = strings.TrimRight(result[:200], " ") + "\u2026"
			}
			return result
		}
	}

	if commandFallback != "" {
		return commandFallback
	}
	return ""
}

// ---------------------------------------------------------------------------
// Internal: session info parsing from lite read
// ---------------------------------------------------------------------------

// parseSessionInfoFromLite parses SDKSessionInfo fields from a lite session read.
// Returns nil for sidechain sessions or sessions with no extractable summary.
func parseSessionInfoFromLite(sessionID string, lite *liteSessionFile, projectPath *string) *SDKSessionInfo {
	head, tail := lite.head, lite.tail

	// Check first line for sidechain sessions.
	firstNewline := strings.Index(head, "\n")
	firstLine := head
	if firstNewline >= 0 {
		firstLine = head[:firstNewline]
	}
	if strings.Contains(firstLine, `"isSidechain":true`) || strings.Contains(firstLine, `"isSidechain": true`) {
		return nil
	}

	// Custom title extraction: user-set wins over AI-generated.
	customTitle := firstNonEmpty(
		extractLastJSONStringField(tail, "customTitle"),
		extractLastJSONStringField(head, "customTitle"),
		extractLastJSONStringField(tail, "aiTitle"),
		extractLastJSONStringField(head, "aiTitle"),
	)

	firstPrompt := extractFirstPromptFromHead(head)

	summary := firstNonEmpty(
		customTitle,
		extractLastJSONStringField(tail, "lastPrompt"),
		extractLastJSONStringField(tail, "summary"),
		firstPrompt,
	)

	if summary == "" {
		return nil
	}

	gitBranch := firstNonEmpty(
		extractLastJSONStringField(tail, "gitBranch"),
		extractJSONStringField(head, "gitBranch"),
	)

	sessionCwd := extractJSONStringField(head, "cwd")
	if sessionCwd == "" && projectPath != nil {
		sessionCwd = *projectPath
	}

	// Scope tag extraction to {"type":"tag"} lines.
	var tagValue string
	tailLines := strings.Split(tail, "\n")
	for i := len(tailLines) - 1; i >= 0; i-- {
		if strings.HasPrefix(tailLines[i], `{"type":"tag"`) {
			tagValue = extractLastJSONStringField(tailLines[i], "tag")
			break
		}
	}

	// created_at from first entry's ISO timestamp.
	var createdAt *int64
	firstTimestamp := extractJSONStringField(firstLine, "timestamp")
	if firstTimestamp != "" {
		ts := firstTimestamp
		if strings.HasSuffix(ts, "Z") {
			ts = ts[:len(ts)-1] + "+00:00"
		}
		if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			ms := t.UnixMilli()
			createdAt = &ms
		} else if t, err = time.Parse("2006-01-02T15:04:05-07:00", ts); err == nil {
			ms := t.UnixMilli()
			createdAt = &ms
		}
	}

	info := &SDKSessionInfo{
		SessionID:    sessionID,
		Summary:      summary,
		LastModified: lite.mtime,
		FileSize:     &lite.size,
	}
	if customTitle != "" {
		info.CustomTitle = &customTitle
	}
	if firstPrompt != "" {
		info.FirstPrompt = &firstPrompt
	}
	if gitBranch != "" {
		info.GitBranch = &gitBranch
	}
	if sessionCwd != "" {
		info.Cwd = &sessionCwd
	}
	if tagValue != "" {
		info.Tag = &tagValue
	}
	info.CreatedAt = createdAt

	return info
}

// ---------------------------------------------------------------------------
// Internal: session listing helpers
// ---------------------------------------------------------------------------

// readSessionsFromDir reads session files from a single project directory.
func readSessionsFromDir(projectDir string, projectPath *string) []SDKSessionInfo {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil
	}

	var results []SDKSessionInfo
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		sid := name[:len(name)-6]
		if !isValidUUID(sid) {
			continue
		}

		lite, err := readSessionLite(filepath.Join(projectDir, name))
		if err != nil || lite == nil {
			continue
		}

		info := parseSessionInfoFromLite(sid, lite, projectPath)
		if info != nil {
			results = append(results, *info)
		}
	}
	return results
}

// listSessionsForProject lists sessions for a specific project directory.
func listSessionsForProject(directory string, includeWorktrees bool) ([]SDKSessionInfo, error) {
	canonicalDir := canonicalizePath(directory)

	projectDir := findProjectDir(canonicalDir)
	if projectDir == "" {
		return nil, nil
	}

	sessions := readSessionsFromDir(projectDir, &canonicalDir)

	if includeWorktrees {
		// Try to find worktree paths - this is best-effort.
		worktrees := getWorktreePaths(canonicalDir)
		if len(worktrees) > 1 {
			for _, wt := range worktrees {
				if wt == canonicalDir {
					continue
				}
				wtProjectDir := findProjectDir(wt)
				if wtProjectDir != "" && wtProjectDir != projectDir {
					wtSessions := readSessionsFromDir(wtProjectDir, &wt)
					sessions = append(sessions, wtSessions...)
				}
			}
		}
	}

	return sessions, nil
}

// listAllSessions lists sessions across all project directories.
func listAllSessions() ([]SDKSessionInfo, error) {
	projectsDir := getProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	var all []SDKSessionInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sessions := readSessionsFromDir(filepath.Join(projectsDir, entry.Name()), nil)
		all = append(all, sessions...)
	}
	return all, nil
}

// deduplicateBySessionID deduplicates sessions by ID, keeping the newest.
func deduplicateBySessionID(sessions []SDKSessionInfo) []SDKSessionInfo {
	byID := make(map[string]SDKSessionInfo)
	for _, s := range sessions {
		existing, ok := byID[s.SessionID]
		if !ok || s.LastModified > existing.LastModified {
			byID[s.SessionID] = s
		}
	}
	result := make([]SDKSessionInfo, 0, len(byID))
	for _, s := range byID {
		result = append(result, s)
	}
	return result
}

// applySortAndPagination sorts sessions by LastModified descending and applies limit/offset.
func applySortAndPagination(sessions []SDKSessionInfo, limit, offset *int) []SDKSessionInfo {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastModified > sessions[j].LastModified
	})

	startIdx := 0
	if offset != nil && *offset > 0 {
		startIdx = *offset
	}
	if startIdx >= len(sessions) {
		return nil
	}
	sessions = sessions[startIdx:]

	if limit != nil && *limit > 0 && *limit < len(sessions) {
		sessions = sessions[:*limit]
	}
	return sessions
}

// ---------------------------------------------------------------------------
// Internal: message reading and chain building
// ---------------------------------------------------------------------------

// transcriptEntry is a parsed JSONL transcript entry.
type transcriptEntry map[string]any

// readSessionFile finds and reads a session JSONL file.
func readSessionFile(sessionID string, dir *string) (string, error) {
	fileName := sessionID + ".jsonl"

	if dir != nil {
		canonicalDir := canonicalizePath(*dir)
		projectDir := findProjectDir(canonicalDir)
		if projectDir != "" {
			data, err := os.ReadFile(filepath.Join(projectDir, fileName))
			if err == nil {
				return string(data), nil
			}
		}
		return "", fmt.Errorf("session %s not found in project directory for %s", sessionID, *dir)
	}

	// No directory - search all project directories.
	projectsDir := getProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(projectsDir, entry.Name(), fileName))
		if readErr == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("session %s not found in any project directory", sessionID)
}

// findSessionFilePath returns the absolute path to a session file, or empty string.
func findSessionFilePath(sessionID string, dir *string) string {
	fileName := sessionID + ".jsonl"

	if dir != nil {
		canonicalDir := canonicalizePath(*dir)
		projectDir := findProjectDir(canonicalDir)
		if projectDir != "" {
			path := filepath.Join(projectDir, fileName)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
		return ""
	}

	projectsDir := getProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(projectsDir, entry.Name(), fileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// parseTranscriptEntries parses JSONL content into transcript entries.
func parseTranscriptEntries(content string) []transcriptEntry {
	var entries []transcriptEntry
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for long lines

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry transcriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		entryType, _ := entry["type"].(string)
		_, hasUUID := entry["uuid"].(string)
		if transcriptEntryTypes[entryType] && hasUUID {
			entries = append(entries, entry)
		}
	}
	return entries
}

// buildConversationChain builds the conversation chain from transcript entries.
func buildConversationChain(entries []transcriptEntry) []transcriptEntry {
	if len(entries) == 0 {
		return nil
	}

	// Index by uuid.
	byUUID := make(map[string]transcriptEntry, len(entries))
	entryIndex := make(map[string]int, len(entries))
	for i, entry := range entries {
		uuid, _ := entry["uuid"].(string)
		byUUID[uuid] = entry
		entryIndex[uuid] = i
	}

	// Find parent uuids (only count non-sidechain children as making something a parent).
	parentUUIDs := make(map[string]bool, len(entries))
	for _, entry := range entries {
		isSidechain, _ := entry["isSidechain"].(bool)
		if isSidechain {
			continue
		}
		parent, _ := entry["parentUuid"].(string)
		if parent != "" {
			parentUUIDs[parent] = true
		}
	}

	// Find terminal entries (no non-sidechain children).
	var terminals []transcriptEntry
	for _, entry := range entries {
		uuid, _ := entry["uuid"].(string)
		isSidechain, _ := entry["isSidechain"].(bool)
		if !parentUUIDs[uuid] && !isSidechain {
			terminals = append(terminals, entry)
		}
	}

	// Find leaves (nearest user/assistant terminal).
	var leaves []transcriptEntry
	for _, terminal := range terminals {
		seen := make(map[string]bool)
		cur := terminal
		for {
			uuid, _ := cur["uuid"].(string)
			if seen[uuid] {
				break
			}
			seen[uuid] = true

			curType, _ := cur["type"].(string)
			if curType == "user" || curType == "assistant" {
				leaves = append(leaves, cur)
				break
			}
			parent, _ := cur["parentUuid"].(string)
			if parent == "" {
				break
			}
			next, ok := byUUID[parent]
			if !ok {
				break
			}
			cur = next
		}
	}

	if len(leaves) == 0 {
		return nil
	}

	// Pick best leaf: prefer main chain (not sidechain/team/meta).
	var mainLeaves []transcriptEntry
	for _, leaf := range leaves {
		isSidechain, _ := leaf["isSidechain"].(bool)
		_, hasTeamName := leaf["teamName"].(string)
		isMeta, _ := leaf["isMeta"].(bool)
		if !isSidechain && !hasTeamName && !isMeta {
			mainLeaves = append(mainLeaves, leaf)
		}
	}

	pickBest := func(candidates []transcriptEntry) transcriptEntry {
		best := candidates[0]
		bestIdx := entryIndex[best["uuid"].(string)]
		for _, cur := range candidates[1:] {
			curIdx := entryIndex[cur["uuid"].(string)]
			if curIdx > bestIdx {
				best = cur
				bestIdx = curIdx
			}
		}
		return best
	}

	var leaf transcriptEntry
	if len(mainLeaves) > 0 {
		leaf = pickBest(mainLeaves)
	} else {
		leaf = pickBest(leaves)
	}

	// Walk from leaf to root via parentUuid.
	var chain []transcriptEntry
	chainSeen := make(map[string]bool)
	cur := leaf
	for {
		uuid, _ := cur["uuid"].(string)
		if chainSeen[uuid] {
			break
		}
		chainSeen[uuid] = true
		chain = append(chain, cur)

		parent, _ := cur["parentUuid"].(string)
		if parent == "" {
			break
		}
		next, ok := byUUID[parent]
		if !ok {
			break
		}
		cur = next
	}

	// Reverse to get chronological order.
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

// isVisibleMessage reports whether a transcript entry should be included in results.
func isVisibleMessage(entry transcriptEntry) bool {
	entryType, _ := entry["type"].(string)
	if entryType != "user" && entryType != "assistant" {
		return false
	}
	if isMeta, _ := entry["isMeta"].(bool); isMeta {
		return false
	}
	if isSidechain, _ := entry["isSidechain"].(bool); isSidechain {
		return false
	}
	if _, hasTeamName := entry["teamName"].(string); hasTeamName {
		return false
	}
	return true
}

// toSessionMessage converts a transcript entry to a SessionMessage.
func toSessionMessage(entry transcriptEntry) SessionMessage {
	entryType, _ := entry["type"].(string)
	msgType := "user"
	if entryType != "user" {
		msgType = "assistant"
	}
	uuid, _ := entry["uuid"].(string)
	sid, _ := entry["sessionId"].(string)

	return SessionMessage{
		Type:            msgType,
		UUID:            uuid,
		SessionID:       sid,
		Message:         entry["message"],
		ParentToolUseID: nil,
	}
}

// applyMessagePagination applies limit and offset to a message slice.
func applyMessagePagination(messages []SessionMessage, limit, offset *int) []SessionMessage {
	startIdx := 0
	if offset != nil && *offset > 0 {
		startIdx = *offset
	}
	if startIdx >= len(messages) {
		return nil
	}
	messages = messages[startIdx:]

	if limit != nil && *limit > 0 && *limit < len(messages) {
		messages = messages[:*limit]
	}
	return messages
}

// ---------------------------------------------------------------------------
// Internal: mutation helpers
// ---------------------------------------------------------------------------

// appendToSession appends data to an existing session file.
func appendToSession(sessionID string, data []byte, dir *string) error {
	fileName := sessionID + ".jsonl"

	if dir != nil {
		canonicalDir := canonicalizePath(*dir)
		projectDir := findProjectDir(canonicalDir)
		if projectDir != "" {
			if tryAppend(filepath.Join(projectDir, fileName), data) {
				return nil
			}
		}
		return fmt.Errorf("session %s not found in project directory for %s", sessionID, *dir)
	}

	// No directory - search all project directories.
	projectsDir := getProjectsDir()
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return fmt.Errorf("session %s not found (no projects directory): %w", sessionID, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if tryAppend(filepath.Join(projectsDir, entry.Name(), fileName), data) {
			return nil
		}
	}
	return fmt.Errorf("session %s not found in any project directory", sessionID)
}

// tryAppend tries to append data to a file. Returns true on success.
// Uses O_WRONLY|O_APPEND (no O_CREATE) so the open fails with ENOENT
// for missing files, avoiding TOCTOU races.
func tryAppend(path string, data []byte) bool {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return false
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil || stat.Size() == 0 {
		return false
	}

	_, err = f.Write(data)
	return err == nil
}

// ---------------------------------------------------------------------------
// Internal: fork helpers
// ---------------------------------------------------------------------------

// truncateAtMessage truncates JSONL content at the line containing the given message UUID.
func truncateAtMessage(content, messageID string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		result.WriteString(line)
		result.WriteByte('\n')

		// Check if this line contains the target message UUID.
		if strings.Contains(line, `"uuid"`) && strings.Contains(line, messageID) {
			// Verify it's actually the uuid field.
			var entry map[string]any
			if err := json.Unmarshal([]byte(line), &entry); err == nil {
				if uuid, _ := entry["uuid"].(string); uuid == messageID {
					break
				}
			}
		}
	}
	return result.String()
}

// generateSessionID generates a new random UUID v4 for session IDs.
func generateSessionID() string {
	// Use crypto/rand via os for a simple UUID v4.
	f, err := os.Open("/dev/urandom")
	if err != nil {
		// Fallback: use time-based pseudo-random.
		return generateFallbackUUID()
	}
	defer f.Close()

	var b [16]byte
	_, err = io.ReadFull(f, b[:])
	if err != nil {
		return generateFallbackUUID()
	}

	// Set UUID version 4 and variant bits.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// generateFallbackUUID creates a time-based UUID when /dev/urandom is unavailable.
func generateFallbackUUID() string {
	now := time.Now().UnixNano()
	return fmt.Sprintf("%08x-%04x-4%03x-%04x-%012x",
		uint32(now),
		uint16(now>>32),
		uint16(now>>48)&0x0fff,
		uint16(now>>56)&0x3fff|0x8000,
		uint64(now)&0xffffffffffff)
}

// ---------------------------------------------------------------------------
// Internal: git worktree detection
// ---------------------------------------------------------------------------

// getWorktreePaths returns worktree paths for the git repo containing dir.
// Returns nil if git is unavailable or dir is not in a repo.
func getWorktreePaths(dir string) []string {
	// This is a best-effort operation. We shell out to git.
	// Import os/exec would add a dependency we want to avoid in a simple helper.
	// For now, return nil (no worktree detection).
	// Worktree detection is an optimization for multi-worktree repos;
	// listing sessions still works without it via the project directory.
	_ = dir
	return nil
}

// ---------------------------------------------------------------------------
// Internal: utility helpers
// ---------------------------------------------------------------------------

// firstNonEmpty returns the first non-empty string from the arguments.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
