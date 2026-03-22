package claudecode

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreateSession(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	assertSessionTransportConnected(t, transport)

	if err := session.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Close is idempotent
	if err := session.Close(); err != nil {
		t.Errorf("Second Close failed: %v", err)
	}
}

func TestSession_SingleTurn(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	transport.enqueueTurnMessages(
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Hello there!"}}},
		&ResultMessage{SessionID: "sess-123", Result: strPtr("done")},
	)

	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	// Send a message
	if err := session.Send(ctx, "Hi"); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Stream the turn
	iter := session.Stream(ctx)
	var messages []Message
	for {
		msg, err := iter.Next(ctx)
		if errors.Is(err, ErrNoMoreMessages) || errors.Is(err, ErrNoTurnInProgress) {
			break
		}
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
		messages = append(messages, msg)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// First message should be assistant
	if _, ok := messages[0].(*AssistantMessage); !ok {
		t.Errorf("Expected AssistantMessage, got %T", messages[0])
	}

	// Second should be result
	result, ok := messages[1].(*ResultMessage)
	if !ok {
		t.Errorf("Expected ResultMessage, got %T", messages[1])
	} else if result.SessionID != "sess-123" {
		t.Errorf("Expected session_id 'sess-123', got %q", result.SessionID)
	}

	// SessionID should be extracted
	if session.SessionID() != "sess-123" {
		t.Errorf("Expected SessionID 'sess-123', got %q", session.SessionID())
	}
}

func TestSession_MultiTurn(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	// Enqueue messages for two turns
	transport.enqueueTurnMessages(
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Turn 1 reply"}}},
		&ResultMessage{SessionID: "sess-multi", Result: strPtr("turn1")},
	)
	transport.enqueueTurnMessages(
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "Turn 2 reply"}}},
		&ResultMessage{SessionID: "sess-multi", Result: strPtr("turn2")},
	)

	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	// --- Turn 1 ---
	if err := session.Send(ctx, "First message"); err != nil {
		t.Fatalf("Send (turn 1) failed: %v", err)
	}
	iter := session.Stream(ctx)
	consumeAllMessages(t, ctx, iter)

	if session.SessionID() != "sess-multi" {
		t.Errorf("Expected SessionID 'sess-multi' after turn 1, got %q", session.SessionID())
	}
	assertLastResultText(t, session, "turn1")

	// --- Turn 2 ---
	if err := session.Send(ctx, "Second message"); err != nil {
		t.Fatalf("Send (turn 2) failed: %v", err)
	}
	iter = session.Stream(ctx)
	consumeAllMessages(t, ctx, iter)

	// SessionID persists
	if session.SessionID() != "sess-multi" {
		t.Errorf("Expected SessionID 'sess-multi' after turn 2, got %q", session.SessionID())
	}
	assertLastResultText(t, session, "turn2")
}

func TestSession_SessionIDExtraction(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	transport.enqueueTurnMessages(
		&ResultMessage{SessionID: "from-result-msg", Result: strPtr("ok")},
	)

	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	if err := session.Send(ctx, "test"); err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	iter := session.Stream(ctx)
	consumeAllMessages(t, ctx, iter)

	if session.SessionID() != "from-result-msg" {
		t.Errorf("Expected SessionID 'from-result-msg', got %q", session.SessionID())
	}
}

func TestSession_SystemMessageSessionID(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	transport.enqueueTurnMessages(
		&SystemMessage{
			Subtype: "init",
			Data:    map[string]any{"session_id": "from-sys-msg", "tools": []string{}},
		},
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "reply"}}},
		&ResultMessage{SessionID: "from-sys-msg", Result: strPtr("ok")},
	)

	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	if err := session.Send(ctx, "test"); err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	iter := session.Stream(ctx)

	// Read the first message (SystemMessage)
	msg, err := iter.Next(ctx)
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if _, ok := msg.(*SystemMessage); !ok {
		t.Fatalf("Expected SystemMessage, got %T", msg)
	}

	// After reading SystemMessage, sessionID should be extracted early
	if session.SessionID() != "from-sys-msg" {
		t.Errorf("Expected early SessionID 'from-sys-msg', got %q", session.SessionID())
	}

	// Consume remaining messages
	consumeAllMessages(t, ctx, iter)
}

func TestSession_SendWhileTurnActive(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	// Don't enqueue a ResultMessage — the turn never completes
	transport.enqueueTurnMessages(
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "still going"}}},
	)

	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	if err := session.Send(ctx, "first"); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Read one message so the turn is active but not complete
	iter := session.Stream(ctx)
	_, err = iter.Next(ctx)
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}

	// Try to send again with a short timeout — should fail with context expiration
	// since the turn is still active and no drain is happening
	shortCtx, shortCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer shortCancel()

	err = session.Send(shortCtx, "second")
	if err == nil {
		t.Fatal("Expected error from Send while turn active, got nil")
	}
	// Should be context deadline or ErrTurnInProgress
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, ErrTurnInProgress) {
		t.Errorf("Expected DeadlineExceeded or ErrTurnInProgress, got: %v", err)
	}
}

func TestSession_StreamBeforeSend(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	// Stream without Send — iterator should immediately return ErrNoTurnInProgress
	iter := session.Stream(ctx)
	_, err = iter.Next(ctx)
	if !errors.Is(err, ErrNoTurnInProgress) {
		t.Errorf("Expected ErrNoTurnInProgress, got: %v", err)
	}
}

func TestSession_CloseSession(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}

	// Close should succeed
	if err := session.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Send after close should fail
	err = session.Send(ctx, "too late")
	if !errors.Is(err, ErrSessionClosed) {
		t.Errorf("Expected ErrSessionClosed, got: %v", err)
	}
}

func TestSession_IteratorClose_Drains(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	// Enqueue messages for turn 1 (will be partially consumed then drained)
	transport.enqueueTurnMessages(
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "msg1"}}},
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "msg2"}}},
		&ResultMessage{SessionID: "sess-drain", Result: strPtr("drained")},
	)
	// Enqueue messages for turn 2
	transport.enqueueTurnMessages(
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "turn2-reply"}}},
		&ResultMessage{SessionID: "sess-drain", Result: strPtr("turn2-done")},
	)

	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	// Start turn 1
	if err := session.Send(ctx, "turn1"); err != nil {
		t.Fatalf("Send (turn 1) failed: %v", err)
	}

	// Read only first message
	iter := session.Stream(ctx)
	msg, err := iter.Next(ctx)
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if am, ok := msg.(*AssistantMessage); !ok {
		t.Fatalf("Expected AssistantMessage, got %T", msg)
	} else {
		tb, ok := am.Content[0].(*TextBlock)
		if !ok || tb.Text != "msg1" {
			t.Fatalf("Expected text 'msg1'")
		}
	}

	// Close the iterator prematurely — should drain remaining messages
	if err := iter.Close(); err != nil {
		t.Fatalf("iterator Close failed: %v", err)
	}

	// Wait briefly for the drain goroutine to process
	time.Sleep(200 * time.Millisecond)

	// Now turn 2 should work
	if err := session.Send(ctx, "turn2"); err != nil {
		t.Fatalf("Send (turn 2) failed: %v", err)
	}
	iter2 := session.Stream(ctx)
	msgs := consumeAllMessages(t, ctx, iter2)

	if len(msgs) != 2 {
		t.Fatalf("Expected 2 messages in turn 2, got %d", len(msgs))
	}

	assertLastResultText(t, session, "turn2-done")
}

func TestResumeSession(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()

	// We need to use CreateSessionWithTransport for testing since ResumeSession
	// uses CreateSession internally (which would try to find the real CLI).
	// Instead, test the underlying mechanics: WithResume option is applied.
	session, err := CreateSessionWithTransport(ctx, transport, WithResume("existing-sess-id"))
	if err != nil {
		t.Fatalf("CreateSessionWithTransport with WithResume failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	// Manually set sessionID as ResumeSession would
	session.mu.Lock()
	session.sessionID = "existing-sess-id"
	session.mu.Unlock()

	if session.SessionID() != "existing-sess-id" {
		t.Errorf("Expected SessionID 'existing-sess-id', got %q", session.SessionID())
	}

	assertSessionTransportConnected(t, transport)
}

func TestPrompt_OneShot(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	transport.enqueueTurnMessages(
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "The answer is 4"}}},
		&ResultMessage{SessionID: "prompt-sess", Result: strPtr("4")},
	)

	// We can't use the real Prompt() since it calls CreateSession which needs CLI.
	// Instead test the equivalent logic with transport.
	session, err := CreateSessionWithTransport(ctx, transport)
	if err != nil {
		t.Fatalf("CreateSessionWithTransport failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	if err := session.Send(ctx, "What is 2+2?"); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	var result *ResultMessage
	iter := session.Stream(ctx)
	for {
		msg, err := iter.Next(ctx)
		if errors.Is(err, ErrNoMoreMessages) || errors.Is(err, ErrNoTurnInProgress) {
			break
		}
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
		if r, ok := msg.(*ResultMessage); ok {
			result = r
			break
		}
	}

	if result == nil {
		t.Fatal("Expected ResultMessage, got nil")
	}
	if result.SessionID != "prompt-sess" {
		t.Errorf("Expected session_id 'prompt-sess', got %q", result.SessionID)
	}
	if result.Result == nil || *result.Result != "4" {
		t.Errorf("Expected result '4', got %v", result.Result)
	}
}

func TestWithSession_ResourceManagement(t *testing.T) {
	ctx, cancel := setupSessionTestContext(t, 5*time.Second)
	defer cancel()

	transport := newSessionMockTransport()
	transport.enqueueTurnMessages(
		&ResultMessage{SessionID: "ws-sess", Result: strPtr("ok")},
	)

	// Create session, run function, verify cleanup
	var capturedSessionID string
	err := withSessionTransport(ctx, transport, func(s *Session) error {
		if err := s.Send(ctx, "test"); err != nil {
			return err
		}
		iter := s.Stream(ctx)
		consumeAllMessages(t, ctx, iter)
		capturedSessionID = s.SessionID()
		return nil
	})
	if err != nil {
		t.Fatalf("WithSession failed: %v", err)
	}

	if capturedSessionID != "ws-sess" {
		t.Errorf("Expected SessionID 'ws-sess', got %q", capturedSessionID)
	}

	// Transport should be disconnected after WithSession returns
	transport.mu.Lock()
	connected := transport.connected
	transport.mu.Unlock()
	if connected {
		t.Error("Expected transport to be disconnected after WithSession")
	}
}

// ---------------------------------------------------------------------------
// Mock Transport
// ---------------------------------------------------------------------------

// sessionMockTransport is a mock transport for Session tests.
// It supports enqueueing messages for multiple turns and delivers them
// through channels, matching the real transport behavior.
type sessionMockTransport struct {
	mu           sync.Mutex
	connected    bool
	closed       bool
	sentMessages []StreamMessage
	allMessages  []Message // all messages across all turns, in order
	msgChan      chan Message
	errChan      chan error
}

func newSessionMockTransport() *sessionMockTransport {
	return &sessionMockTransport{}
}

// enqueueTurnMessages adds messages for a turn. Messages are enqueued in order
// across all calls, and are delivered via the message channel on Connect.
func (t *sessionMockTransport) enqueueTurnMessages(msgs ...Message) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.allMessages = append(t.allMessages, msgs...)
}

func (t *sessionMockTransport) Connect(_ context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		t.closed = false
	}

	t.connected = true
	return nil
}

func (t *sessionMockTransport) SendMessage(_ context.Context, message StreamMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return fmt.Errorf("not connected")
	}
	t.sentMessages = append(t.sentMessages, message)

	// When a message is sent, deliver the next batch of enqueued messages
	// up to and including the next ResultMessage (one turn's worth).
	if t.msgChan != nil {
		var remaining []Message
		delivered := false
		for i, msg := range t.allMessages {
			if !delivered {
				t.msgChan <- msg
				if _, isResult := msg.(*ResultMessage); isResult {
					delivered = true
					remaining = t.allMessages[i+1:]
				}
			}
		}
		if !delivered {
			// No ResultMessage found — deliver all remaining
			remaining = nil
		}
		t.allMessages = remaining
	}
	return nil
}

func (t *sessionMockTransport) ReceiveMessages(_ context.Context) (<-chan Message, <-chan error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.msgChan == nil {
		t.msgChan = make(chan Message, 100)
		t.errChan = make(chan error, 10)
	}

	return t.msgChan, t.errChan
}

func (t *sessionMockTransport) Interrupt(_ context.Context) error {
	return nil
}

func (t *sessionMockTransport) SetModel(_ context.Context, _ *string) error {
	return nil
}

func (t *sessionMockTransport) SetPermissionMode(_ context.Context, _ string) error {
	return nil
}

func (t *sessionMockTransport) RewindFiles(_ context.Context, _ string) error {
	return nil
}

func (t *sessionMockTransport) GetMcpStatus(_ context.Context) ([]McpServerStatusEntry, error) {
	return nil, nil
}

func (t *sessionMockTransport) ReconnectMcpServer(_ context.Context, _ string) error {
	return nil
}

func (t *sessionMockTransport) ToggleMcpServer(_ context.Context, _ string, _ bool) error {
	return nil
}

func (t *sessionMockTransport) SetMcpServers(_ context.Context, _ map[string]any) (map[string]any, error) {
	return nil, nil
}

func (t *sessionMockTransport) StopTask(_ context.Context, _ string) error {
	return nil
}

func (t *sessionMockTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.connected = false
	t.closed = true

	if t.msgChan != nil {
		close(t.msgChan)
		t.msgChan = nil
	}
	if t.errChan != nil {
		close(t.errChan)
		t.errChan = nil
	}

	return nil
}

func (t *sessionMockTransport) GetValidator() *StreamValidator {
	return &StreamValidator{}
}

// ---------------------------------------------------------------------------
// Test Helpers
// ---------------------------------------------------------------------------

func setupSessionTestContext(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

func assertSessionTransportConnected(t *testing.T, transport *sessionMockTransport) {
	t.Helper()
	transport.mu.Lock()
	connected := transport.connected
	transport.mu.Unlock()
	if !connected {
		t.Error("Expected transport to be connected")
	}
}

func consumeAllMessages(t *testing.T, ctx context.Context, iter *TurnIterator) []Message {
	t.Helper()
	var messages []Message
	for {
		msg, err := iter.Next(ctx)
		if errors.Is(err, ErrNoMoreMessages) || errors.Is(err, ErrNoTurnInProgress) {
			break
		}
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
		messages = append(messages, msg)
	}
	return messages
}

func assertLastResultText(t *testing.T, session *Session, expected string) {
	t.Helper()
	lr := session.LastResult()
	if lr == nil {
		t.Fatalf("Expected LastResult to be non-nil")
	}
	if lr.Result == nil || *lr.Result != expected {
		t.Errorf("Expected LastResult.Result = %q, got %v", expected, lr.Result)
	}
}

func strPtr(s string) *string {
	return &s
}

// withSessionTransport is a test helper that mirrors WithSession but uses a custom transport.
func withSessionTransport(ctx context.Context, transport Transport, fn func(*Session) error, opts ...Option) error {
	session, err := CreateSessionWithTransport(ctx, transport, opts...)
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()
	return fn(session)
}
