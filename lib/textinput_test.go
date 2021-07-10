package lib

import (
	"reflect"
	"testing"
)

func TestInsertRune(t *testing.T) {

	testCases := []struct {
		input    string
		expected string
		chr      rune
		pos      int
	}{
		{
			input:    "it's coo",
			pos:      8,
			chr:      'l',
			expected: "it's cool",
		},
		{
			input:    "it's coo",
			pos:      4,
			chr:      'X',
			expected: "it'sX coo",
		},
	}
	for _, tc := range testCases {
		res := InsertRune([]rune(tc.input), tc.pos, tc.chr)
		if !reflect.DeepEqual(res, []rune(tc.expected)) {
			t.Errorf("inserted runes is not equal to expected (input: %s) %s", tc.input, string(res))
		}
	}
}

func TestRmRune(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
		pos      int
	}{
		{
			input:    "it's coo",
			pos:      8,
			expected: "it's co",
		},
		{
			input:    "it's coo",
			pos:      4,
			expected: "it' coo",
		},
	}
	for _, tc := range testCases {
		res := RemoveRuneBackward([]rune(tc.input), tc.pos)
		if !reflect.DeepEqual(res, []rune(tc.expected)) {
			t.Errorf("removed runes is not equal to expected (input: %s) %s", tc.input, string(res))
		}
	}
}

func TestRmRuneWord(t *testing.T) {
	testCases := []struct {
		input       string
		expected    string
		expectedlen int
		pos         int
	}{
		{
			input:       "it's coo",
			pos:         8,
			expected:    "it's ",
			expectedlen: 3,
		},
		{
			input:       "it's cool",
			pos:         4,
			expected:    "it' cool",
			expectedlen: 1,
		},
		{
			input:       "cmd arg1:val1 arg2:val2",
			pos:         8,
			expected:    "cmd :val1 arg2:val2",
			expectedlen: 4,
		},
		{
			input:       "cmd arg1:val1 arg2:val2",
			pos:         13,
			expected:    "cmd arg1: arg2:val2",
			expectedlen: 4,
		},
		{
			input:       "cmd arg1:val1 ",
			pos:         14,
			expected:    "cmd arg1:",
			expectedlen: 5,
		},
	}
	for _, tc := range testCases {
		res, nbremoved := RemoveRuneWordBackward([]rune(tc.input), tc.pos)
		if !reflect.DeepEqual(res, []rune(tc.expected)) {
			t.Errorf("removed runes word is not equal to expected (input: %s) %s", tc.input, string(res))
		}
		if nbremoved != tc.expectedlen {
			t.Errorf("removed word has not expected length (input: %s) %d", tc.input, nbremoved)
		}
	}
}
