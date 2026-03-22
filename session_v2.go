package claudecode

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Session error sentinels.
var (
	// ErrTurnInProgress indicates that a previous turn is still active.
	// Consume or close the stream before calling Send() again.
	ErrTurnInProgress = errors.New("previous turn still in progress; consume or close the stream first")

	// ErrSessionClosed indicates the session has been closed and cannot accept new messages.
	ErrSessionClosed = errors.New("session is closed")

	// ErrNoTurnInProgress indicates that no turn is active.
	// Call Send() before calling Stream().
	ErrNoTurnInProgress = errors.New("no turn in progress; call Send() first")
)

// Session provides a multi-turn conversation interface to Claude Code CLI.
// It wraps a ClientImpl and manages per-turn streaming, tracking the session ID
// across turns and providing turn-level flow control.
//
// A typical usage pattern is:
//
//	session, err := claudecode.CreateSession(ctx)
//	if err != nil { ... }
//	defer session.Close()
//
//	// First turn
//	session.Send(ctx, "Hello")
//	iter := session.Stream(ctx)
//	for {
//	    msg, err := iter.Next(ctx)
//	    if errors.Is(err, claudecode.ErrNoMoreMessages) { break }
//	    // handle msg
//	}
//
//	// Second turn
//	session.Send(ctx, "Tell me more")
//	iter = session.Stream(ctx)
//	// ...
type Session struct {
	mu         sync.Mutex
	client     *ClientImpl
	sessionID  string
	closed     bool
	turnActive bool
	turnDone   chan struct{} // signals drain completion
	lastResult *ResultMessage
}

// SessionID returns the session ID for this conversation.
// The session ID is extracted from the first ResultMessage or SystemMessage received.
func (s *Session) SessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionID
}

// setSessionID sets the session ID if not already set or updates it.
func (s *Session) setSessionID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id != "" {
		s.sessionID = id
	}
}

// setLastResult stores the last ResultMessage from a completed turn.
func (s *Session) setLastResult(result *ResultMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastResult = result
}

// turnCompleted marks the current turn as complete and signals any waiters.
func (s *Session) turnCompleted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.turnActive = false
	if s.turnDone != nil {
		close(s.turnDone)
	}
}

// Send sends a user message to start a new turn in the conversation.
// Returns ErrTurnInProgress if the previous turn has not been consumed or closed.
// Returns ErrSessionClosed if the session has been closed.
func (s *Session) Send(ctx context.Context, message string) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrSessionClosed
	}
	if s.turnActive {
		// Wait for drain to complete
		done := s.turnDone
		s.mu.Unlock()
		if done != nil {
			select {
			case <-done:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		s.mu.Lock()
		if s.turnActive {
			s.mu.Unlock()
			return ErrTurnInProgress
		}
	}
	s.turnActive = true
	s.turnDone = make(chan struct{})

	// Determine which session ID to use for the query
	sid := s.sessionID
	s.mu.Unlock()

	if sid == "" {
		sid = defaultSessionID
	}

	return s.client.queryWithSession(ctx, message, sid)
}

// SendMessage sends a StreamMessage to start a new turn. This is the advanced
// variant of Send that allows full control over the message payload.
// Returns ErrTurnInProgress if the previous turn has not been consumed or closed.
// Returns ErrSessionClosed if the session has been closed.
func (s *Session) SendMessage(ctx context.Context, msg StreamMessage) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrSessionClosed
	}
	if s.turnActive {
		done := s.turnDone
		s.mu.Unlock()
		if done != nil {
			select {
			case <-done:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		s.mu.Lock()
		if s.turnActive {
			s.mu.Unlock()
			return ErrTurnInProgress
		}
	}
	s.turnActive = true
	s.turnDone = make(chan struct{})
	s.mu.Unlock()

	// Check connection status
	s.client.mu.RLock()
	connected := s.client.connected
	transport := s.client.transport
	s.client.mu.RUnlock()

	if !connected || transport == nil {
		s.turnCompleted()
		return fmt.Errorf("client not connected")
	}

	return transport.SendMessage(ctx, msg)
}

// Stream returns a TurnIterator that yields messages for the current turn.
// The iterator stops after yielding a ResultMessage (the turn boundary).
// You must call Send() or SendMessage() before calling Stream().
func (s *Session) Stream(ctx context.Context) *TurnIterator {
	s.mu.Lock()
	active := s.turnActive
	msgChan := s.client.msgChan
	errChan := s.client.errChan
	s.mu.Unlock()

	return &TurnIterator{
		session:    s,
		msgChan:    msgChan,
		errChan:    errChan,
		done:       !active,
		ctx:        ctx,
		noTurnSent: !active,
	}
}

// Close disconnects the underlying client. It is safe to call multiple times.
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	return s.client.Disconnect()
}

// LastResult returns the ResultMessage from the most recently completed turn,
// or nil if no turn has completed yet.
func (s *Session) LastResult() *ResultMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastResult
}

// TurnIterator yields messages for a single conversation turn.
// It stops after returning a ResultMessage, which marks the turn boundary.
type TurnIterator struct {
	mu         sync.Mutex
	session    *Session
	msgChan    <-chan Message
	errChan    <-chan error
	done       bool
	ctx        context.Context
	noTurnSent bool // true if no turn was active when Stream() was called
}

// Next returns the next message in the current turn.
// When a ResultMessage is encountered, it is returned as the final message.
// Subsequent calls return ErrNoMoreMessages.
// If no turn was started before calling Stream(), returns ErrNoTurnInProgress.
func (ti *TurnIterator) Next(ctx context.Context) (Message, error) {
	ti.mu.Lock()
	if ti.noTurnSent {
		ti.mu.Unlock()
		return nil, ErrNoTurnInProgress
	}
	if ti.done {
		ti.mu.Unlock()
		return nil, ErrNoMoreMessages
	}
	ti.mu.Unlock()

	select {
	case msg, ok := <-ti.msgChan:
		if !ok {
			ti.mu.Lock()
			ti.done = true
			ti.mu.Unlock()
			ti.session.turnCompleted()
			return nil, ErrNoMoreMessages
		}

		// Check for SystemMessage with session_id
		if sysMsg, ok := msg.(*SystemMessage); ok {
			if sid, exists := sysMsg.Data["session_id"]; exists {
				if sidStr, ok := sid.(string); ok {
					currentID := ti.session.SessionID()
					if currentID == "" {
						ti.session.setSessionID(sidStr)
					}
				}
			}
		}

		// Check for ResultMessage (turn boundary)
		if result, ok := msg.(*ResultMessage); ok {
			ti.session.setSessionID(result.SessionID)
			ti.session.setLastResult(result)
			ti.mu.Lock()
			ti.done = true
			ti.mu.Unlock()
			ti.session.turnCompleted()
			return msg, nil
		}

		return msg, nil

	case err, ok := <-ti.errChan:
		if !ok {
			ti.mu.Lock()
			ti.done = true
			ti.mu.Unlock()
			ti.session.turnCompleted()
			return nil, ErrNoMoreMessages
		}
		ti.mu.Lock()
		ti.done = true
		ti.mu.Unlock()
		ti.session.turnCompleted()
		return nil, err

	case <-ctx.Done():
		ti.mu.Lock()
		ti.done = true
		ti.mu.Unlock()
		ti.session.turnCompleted()
		return nil, ctx.Err()
	}
}

// Close stops the iterator. If the turn has not completed naturally,
// it spawns a background goroutine to drain remaining messages until a
// ResultMessage is received or the channel closes. This ensures the next
// Send() can proceed without blocking.
func (ti *TurnIterator) Close() error {
	ti.mu.Lock()
	if ti.done {
		ti.mu.Unlock()
		return nil
	}
	ti.done = true
	ti.mu.Unlock()

	// Drain in background until ResultMessage or channel close
	go func() {
		for {
			select {
			case msg, ok := <-ti.msgChan:
				if !ok {
					ti.session.turnCompleted()
					return
				}
				if result, ok := msg.(*ResultMessage); ok {
					ti.session.setSessionID(result.SessionID)
					ti.session.setLastResult(result)
					ti.session.turnCompleted()
					return
				}
			case <-ti.ctx.Done():
				ti.session.turnCompleted()
				return
			}
		}
	}()
	return nil
}

// CreateSession creates a new Session connected to Claude Code CLI.
// The session is ready for multi-turn conversation after creation.
func CreateSession(ctx context.Context, opts ...Option) (*Session, error) {
	client := NewClient(opts...)

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect session: %w", err)
	}

	impl, ok := client.(*ClientImpl)
	if !ok {
		return nil, fmt.Errorf("unexpected client type")
	}

	return &Session{
		client: impl,
	}, nil
}

// CreateSessionWithTransport creates a Session with a custom transport.
// This is primarily used for testing with mock transports.
func CreateSessionWithTransport(ctx context.Context, transport Transport, opts ...Option) (*Session, error) {
	client := NewClientWithTransport(transport, opts...)

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect session: %w", err)
	}

	impl, ok := client.(*ClientImpl)
	if !ok {
		return nil, fmt.Errorf("unexpected client type")
	}

	return &Session{
		client: impl,
	}, nil
}

// ResumeSession resumes an existing session by ID.
// It creates a new session with the WithResume option applied,
// connecting to the same conversation state identified by sessionID.
func ResumeSession(ctx context.Context, sessionID string, opts ...Option) (*Session, error) {
	opts = append(opts, WithResume(sessionID))
	session, err := CreateSession(ctx, opts...)
	if err != nil {
		return nil, err
	}
	session.sessionID = sessionID
	return session, nil
}

// Prompt is a one-shot convenience function that creates a session, sends a message,
// collects the full result, and closes the session. It returns the final ResultMessage.
func Prompt(ctx context.Context, message string, opts ...Option) (*ResultMessage, error) {
	var result *ResultMessage
	err := WithSession(ctx, func(s *Session) error {
		if err := s.Send(ctx, message); err != nil {
			return err
		}
		iter := s.Stream(ctx)
		for {
			msg, err := iter.Next(ctx)
			if errors.Is(err, ErrNoMoreMessages) || errors.Is(err, ErrNoTurnInProgress) {
				break
			}
			if err != nil {
				return err
			}
			if r, ok := msg.(*ResultMessage); ok {
				result = r
				break
			}
		}
		return nil
	}, opts...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// WithSession provides Go-idiomatic resource management for sessions.
// It creates a session, executes the provided function, and ensures the session
// is properly closed regardless of how fn returns.
func WithSession(ctx context.Context, fn func(*Session) error, opts ...Option) error {
	session, err := CreateSession(ctx, opts...)
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	return fn(session)
}
