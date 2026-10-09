package session

// State is the lifecycle state of a database session.
type State string

const (
	StateConnecting   State = "connecting"
	StateReady        State = "ready"
	StateBusy         State = "busy"
	StateReconnecting State = "reconnecting"
	StateFailed       State = "failed"
	StateClosed       State = "closed"
)

// WriteCheck is the outcome of the write guard for one statement.
type WriteCheck int

const (
	WriteAllowed WriteCheck = iota
	WriteBlocked
	WriteNeedsConfirmation
)

// StatementClass classifies SQL before it reaches the database.
type StatementClass int

const (
	StatementRead StatementClass = iota
	StatementWrite
)
