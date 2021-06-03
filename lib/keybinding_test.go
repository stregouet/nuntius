package lib

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestParseKeyStroke(t *testing.T) {
	testCases := []struct {
		input    string
		expected []*KeyStroke
		err      error
	}{
		{
			input:    "c-space",
			expected: []*KeyStroke{&KeyStroke{tcell.KeyCtrlSpace, 0}},
		},
		{
			input: "space a",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, ' '},
				&KeyStroke{tcell.KeyRune, 'a'},
			},
		},
		{
			input: "space ;",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, ' '},
				&KeyStroke{tcell.KeyRune, ';'},
			},
		},
		{
			input: ". '",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, '.'},
				&KeyStroke{tcell.KeyRune, '\''},
			},
		},
		{
			input: "c-a u ",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyCtrlA, 0},
				&KeyStroke{tcell.KeyRune, 'u'},
			},
		},
		{
			input: "s-a",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, 'A'},
			},
		},
		{
			input: "«",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, '«'},
			},
		},
		{
			input: "s-é",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, 'É'},
			},
		},
	}
	for _, tc := range testCases {
		res, err := ParseKeyStroke(tc.input)
		if tc.err != nil {
			if err != tc.err {
				t.Errorf("unexpected error (input: %s) %v", tc.input, err)
			}
		} else if err != nil {
			t.Errorf("unexpected error (input: %s) %v", tc.input, err)
		} else if !reflect.DeepEqual(res, tc.expected) {
			t.Errorf("parsed keystroke is not equal to expected (input: %s) %#v", tc.input, res)
		}
	}
}
