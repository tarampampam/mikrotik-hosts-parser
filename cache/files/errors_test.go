package files

import (
	"errors"
	"testing"
)

func TestErrorType_String(t *testing.T) {
	t.Parallel()

	var unknownType ErrorType = 255

	tests := []struct {
		giveType   ErrorType
		wantString string
	}{
		{giveType: ErrUnknown, wantString: "unknown"},
		{giveType: ErrFileOpening, wantString: "cannot open file"},
		{giveType: ErrFileReading, wantString: "cannot read file"},
		{giveType: ErrFileWriting, wantString: "cannot write file"},
		{giveType: ErrExpirationDataNotAvailable, wantString: "expiration data is not available"},
		{giveType: unknownType, wantString: "unrecognized error type"},
	}

	for _, tt := range tests {
		t.Run(tt.wantString, func(t *testing.T) {
			if s := tt.giveType.String(); tt.wantString != s {
				t.Errorf("Wrong error type to string convertation. Want: %v, got: %v", tt.wantString, s)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	t.Parallel()

	msg := " foo bar "
	e := Error{Message: msg}

	if e.Error() != msg {
		t.Errorf("Wrong error message returned. Want: %v, got: %v", msg, e.Error())
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	prev := errors.New("foo")
	e := Error{previous: prev}

	if e.Unwrap() != prev {
		t.Errorf("Wrong previous error returned. Want: %v, got: %v", prev, e.Unwrap())
	}
}

func TestError_newError(t *testing.T) {
	t.Parallel()

	msg := "foo"
	prev := errors.New("foo")
	tp := ErrFileWriting

	e := newError(tp, msg, prev)

	if e.Message != msg {
		t.Errorf("Wrong error message returned. Want: %v, got: %v", msg, e.Message)
	}

	if e.Unwrap() != prev {
		t.Errorf("Wrong previous error returned. Want: %v, got: %v", prev, e.Unwrap())
	}

	if e.Type != tp {
		t.Errorf("Wrong error type returned. Want: %v, got: %v", tp, e.Type)
	}
}
