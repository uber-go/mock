package gomock

import (
	"fmt"
	"testing"
	"time"
)

func TestGetString(t *testing.T) {
	now := time.Now()
	var nilTime *time.Time

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "nil stringer should not panic",
			input: nilTime,
			want:  "<nil>",
		},
		{
			name:  "non-nil stringer",
			input: &now,
			want:  now.String(),
		},
		{
			name:  "non-stringer value",
			input: 42,
			want:  "42",
		},
		{
			name:  "nil interface value",
			input: nil,
			want:  "<nil>",
		},
		{
			name:  "panicking stringer handles panic",
			input: panickingStringer{},
			want:  fmt.Sprintf("%v", panickingStringer{}),
		},
		{
			name:  "generated mock struct pointer",
			input: &mockStruct{},
			want:  "*gomock.mockStruct",
		},
		{
			name:  "generated mock struct value",
			input: mockStruct{},
			want:  "gomock.mockStruct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getString(tt.input)
			if got != tt.want {
				t.Errorf("getString() = %q, want %q", got, tt.want)
			}
		})
	}
}

type panickingStringer struct{}

func (panickingStringer) String() string {
	panic("panicking stringer")
}

type mockStruct struct {
	isgomock struct{}
}

