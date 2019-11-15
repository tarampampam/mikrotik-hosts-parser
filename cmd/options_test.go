package cmd

import (
	"reflect"
	"testing"
)

func TestOptions_Struct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		element         func() reflect.StructField
		wantCommand     string
		wantAlias       string
		wantDescription string
	}{
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Options{}).FieldByName("Serve")
				return field
			},
			wantCommand:     "serve",
			wantAlias:       "s",
			wantDescription: "Start web-server",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(Options{}).FieldByName("Version")
				return field
			},
			wantCommand:     "version",
			wantAlias:       "v",
			wantDescription: "Display application version",
		},
	}
	for _, tt := range tests {
		t.Run(tt.wantDescription, func(t *testing.T) {
			el := tt.element()
			if tt.wantCommand != "" {
				value, _ := el.Tag.Lookup("command")
				if value != tt.wantCommand {
					t.Errorf("Wrong value for 'command' tag. Want: %v, got: %v", tt.wantCommand, value)
				}
			}

			if tt.wantAlias != "" {
				value, _ := el.Tag.Lookup("alias")
				if value != tt.wantAlias {
					t.Errorf("Wrong value for 'alias' tag. Want: %v, got: %v", tt.wantAlias, value)
				}
			}

			if tt.wantDescription != "" {
				value, _ := el.Tag.Lookup("description")
				if value != tt.wantDescription {
					t.Errorf("Wrong value for 'description' tag. Want: %v, got: %v", tt.wantDescription, value)
				}
			}
		})
	}
}
