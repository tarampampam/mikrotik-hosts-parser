package checkers

type ReadyChecker struct{}

func NewReadyChecker() *ReadyChecker { return &ReadyChecker{} }

func (*ReadyChecker) Check() error { return nil }
