package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMikrotikDnsStaticEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		element      func() reflect.StructField
		wantComment  string
		wantProperty string
	}{
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(MikrotikDnsStaticEntry{}).FieldByName("Address")
				return field
			},
			wantComment:  "IP address",
			wantProperty: "address",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(MikrotikDnsStaticEntry{}).FieldByName("Comment")
				return field
			},
			wantComment:  "Short description of the item",
			wantProperty: "comment",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(MikrotikDnsStaticEntry{}).FieldByName("Disabled")
				return field
			},
			wantComment:  "Defines whether item is ignored or used",
			wantProperty: "disabled",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(MikrotikDnsStaticEntry{}).FieldByName("Name")
				return field
			},
			wantComment:  "Host name",
			wantProperty: "name",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(MikrotikDnsStaticEntry{}).FieldByName("Regexp")
				return field
			},
			wantProperty: "regexp",
		},
		{
			element: func() reflect.StructField {
				field, _ := reflect.TypeOf(MikrotikDnsStaticEntry{}).FieldByName("TTL")
				return field
			},
			wantComment:  "Time To Live",
			wantProperty: "ttl",
		},
	}
	for _, tt := range tests {
		t.Run("property "+tt.wantProperty, func(t *testing.T) {
			el := tt.element()

			// required tag
			value, _ := el.Tag.Lookup("property")
			if value != tt.wantProperty {
				t.Errorf("Wrong value for 'property' tag. Want: %v, got: %v", tt.wantProperty, value)
			}

			if tt.wantComment != "" {
				value, _ := el.Tag.Lookup("comment")
				if value != tt.wantComment {
					t.Errorf("Wrong value for 'comment' tag. Want: %v, got: %v", tt.wantComment, value)
				}
			}
		})
	}
}

func TestMikrotikDnsStaticEntries_Render(t *testing.T) {
	tests := []struct {
		name          string
		entries       *MikrotikDnsStaticEntries
		renderOptions *RenderOptions
		wantResult    string
		wantError     error
	}{
		{
			name:          "Empty input",
			entries:       &MikrotikDnsStaticEntries{{}},
			renderOptions: &RenderOptions{},
			wantResult:    "",
		},
		{
			name: "Address with comment",
			entries: &MikrotikDnsStaticEntries{{
				Address: "0.0.0.0",
				Comment: "foo comment",
			}},
			renderOptions: &RenderOptions{},
			wantResult:    `address=0.0.0.0 comment="foo comment" disabled=no`,
		},
		{
			name: "Two entries with addresses",
			entries: &MikrotikDnsStaticEntries{{
				Address: "0.0.0.0",
			}, {
				Address: "8.8.8.8",
			}},
			renderOptions: &RenderOptions{},
			wantResult:    "address=0.0.0.0 disabled=no\naddress=8.8.8.8 disabled=no",
		},
		{
			name: "Two entries (one is empty)",
			entries: &MikrotikDnsStaticEntries{{}, {
				Address: "8.8.8.8",
			}},
			renderOptions: &RenderOptions{},
			wantResult:    "address=8.8.8.8 disabled=no",
			wantError:     nil,
		},
		{
			name: "Two entries with Prefix and Postfix",
			entries: &MikrotikDnsStaticEntries{{
				Address: "0.0.0.0",
			}, {
				Address: "8.8.8.8",
			}},
			renderOptions: &RenderOptions{
				RenderEntryOptions: RenderEntryOptions{
					Prefix:  "foo",
					Postfix: "bar",
				},
			},
			wantResult: "foo address=0.0.0.0 disabled=no bar\nfoo address=8.8.8.8 disabled=no bar",
		},
		{
			name: "Entry with all fields",
			entries: &MikrotikDnsStaticEntries{{
				Address:  "1.2.3.4",
				Comment:  "foo comment",
				Disabled: true,
				Name:     "Bar name",
				Regexp:   `.*\.example\.com`,
				TTL:      "1d",
			}},
			renderOptions: &RenderOptions{},
			wantResult:    `address=1.2.3.4 comment="foo comment" disabled=yes name="Bar name" regexp=".*\.example\.com" ttl="1d"`,
		},
		{
			name: "Force empty fields render",
			entries: &MikrotikDnsStaticEntries{{
				Address: "1.2.3.4",
			}},
			renderOptions: &RenderOptions{
				RenderEmpty: true,
			},
			wantResult: `address=1.2.3.4 comment="" disabled=no name="" regexp="" ttl=""`,
		},
		{
			name: "Regular use-case with address, name and comment",
			entries: &MikrotikDnsStaticEntries{{
				Address: "1.2.3.4",
				Comment: "Foo comment",
				Name:    "Foo entry",
			}, {
				Address: "4.3.2.1",
				Comment: "Bar comment",
				Name:    "Bar entry",
			}},
			renderOptions: &RenderOptions{},
			wantResult: `address=1.2.3.4 comment="Foo comment" disabled=no name="Foo entry"` + "\n" +
				`address=4.3.2.1 comment="Bar comment" disabled=no name="Bar entry"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			l, err := tt.entries.Render(&buf, tt.renderOptions)

			if resLen := len(tt.wantResult); resLen != l {
				t.Errorf("Unexpected wrote bytes length: want %d, got %d", resLen, l)
			}

			if tt.wantError != nil && tt.wantError.Error() != err.Error() {
				t.Errorf("Unexpected error: want %v, got %v", tt.wantError, err)
			}

			if res := buf.String(); res != tt.wantResult {
				t.Errorf("Unexpected result. Want:\n[%s]\nGot:\n[%s]", tt.wantResult, res)
			}
		})
	}
}

func TestMikrotikDnsStaticEntries_getStructTagValue(t *testing.T) {
	type T struct {
		F string `one:"1" blank:""`
	}

	entries := MikrotikDnsStaticEntries{}
	ref := reflect.TypeOf(T{})

	if r := entries.getStructTagValue(ref, "F", "one"); r != "1" {
		t.Errorf("Struct tag getter returns %v, but want %v", r, "1")
	}

	if r := entries.getStructTagValue(ref, "F", "blank"); r != "" {
		t.Errorf("Struct tag getter returns %v for blank tag", r)
	}

	if r := entries.getStructTagValue(ref, "F", "miss"); r != "" {
		t.Errorf("Struct tag getter returns %v for non-existing tag", r)
	}

	if r := entries.getStructTagValue(ref, "foo", "bar"); r != "" {
		t.Errorf("Struct tag getter returns %v for non-existing property and tag", r)
	}
}
