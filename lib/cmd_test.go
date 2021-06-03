package lib

import (
	"reflect"
	"testing"
)

func cmdEqual(one, other *Command) bool {
	return one.Partial == other.Partial &&
		one.Name == other.Name &&
		reflect.DeepEqual(one.Args, other.Args)
}

func TestParseCmd(t *testing.T) {
	testCases := []struct {
		query    string
		expected *Command
		err      error
	}{
		{
			query:    "search subject:toto from:jean",
			expected: &Command{false, "search", CmdArgs{"subject": "toto", "from": "jean"}},
			err:      nil,
		},
		{
			query:    "search subject:\"toto funny\" from:jean",
			expected: &Command{false, "search", CmdArgs{"subject": "toto funny", "from": "jean"}},
			err:      nil,
		},
		{
			query:    ">search subject:funny",
			expected: &Command{true, "search", CmdArgs{"subject": "funny"}},
			err:      nil,
		},
		{
			query:    "search subject:\"toto funny from:jean",
			expected: &Command{false, "search", CmdArgs{"subject": "toto funny", "from": "jean"}},
			err:      UnfinishedValueErr,
		},
	}
	for _, tc := range testCases {
		res, err := ParseCmd(tc.query)
		if tc.err != nil {
			if err != tc.err {
				t.Errorf("unexpected error (query: %s) %v", tc.query, err)
			}
		} else if err != nil {
			t.Errorf("unexpected error (query: %s) %v", tc.query, err)
		} else if !reflect.DeepEqual(res, tc.expected) {
			t.Errorf("parsed cmd is not equal to expected (query: %s) %v", tc.query, res)
		}

	}

}
