package modules

import (
	"strings"
	"testing"
)

func TestWriteBlock(t *testing.T) {
	block := "line11"
	beginMarker := []byte("# BEGIN MARKER NCO")
	endMarker := []byte("# END MARKER NCO")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"block is not in file",
			`line0
line1`,
			`line0
line1
# BEGIN MARKER NCO
line11
# END MARKER NCO`,
		},
		{
			"block is in file",
			`line0
# BEGIN MARKER NCO
line11
# END MARKER NCO
line1`,
			`line0
# BEGIN MARKER NCO
line11
# END MARKER NCO
line1`,
		},
		{
			"block is in file but needs updating",
			`line0
# BEGIN MARKER NCO
line99
# END MARKER NCO
line1`,
			`line0
# BEGIN MARKER NCO
line11
# END MARKER NCO
line1`,
		},
		{
			"block is in file but is not closed",
			`line0
line1
# BEGIN MARKER NCO
line11`,
			`line0
line1
# BEGIN MARKER NCO
line11
# END MARKER NCO`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := writeBlock(strings.NewReader(tt.input), beginMarker, endMarker, []byte(block))
			if err != nil {
				t.Error(err)
			}
			if tt.expected != string(out) {
				t.Errorf("expected:\n%s\n\nactual:\n%s", tt.expected, out)
			}
		})
	}
}

func TestDeleteBlock(t *testing.T) {
	beginMarker := []byte("# BEGIN MARKER NCO")
	endMarker := []byte("# END MARKER NCO")

	input := `line0
# BEGIN MARKER NCO
line11
# END MARKER NCO
line1`

	expected := `line0
line1`

	out, err := deleteBlock(strings.NewReader(input), beginMarker, endMarker)
	if err != nil {
		t.Error(err)
	}

	if expected != string(out) {
		t.Errorf("expected:\n%s\n\nactual:\n%s", expected, out)
	}
}
