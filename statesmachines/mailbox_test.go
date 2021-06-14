package statesmachines

import (
	"testing"

	"github.com/stregouet/nuntius/lib"
)

func TestParseNbLine(t *testing.T) {
	testCases := []struct {
		input    lib.Event
		expected int
	}{
		{
			input:    lib.Event{},
			expected: 1,
		},
		{
			input:    lib.Event{Payload: lib.CmdArgs{"line": "5"}},
			expected: 5,
		},
		{
			input:    lib.Event{Payload: lib.CmdArgs{"line": "»"}},
			expected: 1,
		},
		{
			input:    lib.Event{Payload: lib.CmdArgs{"line": "10"}},
			expected: 10,
		},
	}
	for _, tc := range testCases {
		res := getNblines(&tc.input)

		if res != tc.expected {
			t.Errorf("(input: %#v) expected %d, found %d", tc.input, tc.expected, res)
		}
	}

}
