package checkers

// ReadyChecker is a readiness checker.
type ReadyChecker struct{}

// NewReadyChecker creates readiness checker.
func NewReadyChecker() *ReadyChecker { return &ReadyChecker{} }

// Check application is ready for incoming requests processing?
func (*ReadyChecker) Check() error { return nil } // TODO implement me
