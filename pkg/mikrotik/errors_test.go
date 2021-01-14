package mikrotik

import "testing"

func TestConstErr_Error(t *testing.T) {
	cases := []struct {
		name       string
		giveConst  Error
		wantString string
	}{
		{
			name:       "ErrEmptyFields",
			giveConst:  ErrEmptyFields,
			wantString: "required fields does not filled",
		},
		{
			name:       "0",
			giveConst:  Error(0),
			wantString: "unknown error",
		},
		{
			name:       "255",
			giveConst:  Error(255),
			wantString: "unknown error",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.giveConst.Error(); tt.wantString != got {
				t.Errorf(`want: "%s", got: "%s"`, tt.wantString, got)
			}
		})
	}
}
