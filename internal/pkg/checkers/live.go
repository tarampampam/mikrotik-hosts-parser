package checkers

type LiveChecker struct{}

func NewLiveChecker() *LiveChecker { return &LiveChecker{} }

func (*LiveChecker) Check() error { return nil } // TODO implement me
