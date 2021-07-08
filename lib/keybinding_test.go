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
			expected: []*KeyStroke{&KeyStroke{tcell.KeyCtrlSpace, 0, nil}},
		},
		{
			input: "space a",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, ' ', nil},
				&KeyStroke{tcell.KeyRune, 'a', nil},
			},
		},
		{
			input: "space ;",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, ' ', nil},
				&KeyStroke{tcell.KeyRune, ';', nil},
			},
		},
		{
			input: ". '",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, '.', nil},
				&KeyStroke{tcell.KeyRune, '\'', nil},
			},
		},
		{
			input: "c-a u ",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyCtrlA, 0, nil},
				&KeyStroke{tcell.KeyRune, 'u', nil},
			},
		},
		{
			input: "s-a",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, 'A', nil},
			},
		},
		{
			input: "«",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, '«', nil},
			},
		},
		{
			input: "s-é",
			expected: []*KeyStroke{
				&KeyStroke{tcell.KeyRune, 'É', nil},
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
