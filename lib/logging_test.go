package lib

import (
	"errors"
	"testing"
)

func TestParseLevel(t *testing.T) {
	type testCase struct {
		expectederr    error
		expectedresult Level
		param          string
	}
	tcases := []testCase{
		{
			expectederr: errors.New(""),
			param:       "yo",
		},
		{
			expectedresult: DEBUG,
			param:          "debug",
		},
		{
			expectedresult: LOG,
			param:          "log",
		},
		{
			expectedresult: WARN,
			param:          "warn",
		},
		{
			expectedresult: ERROR,
			param:          "erRor",
		},
	}
	for _, tc := range tcases {
		res, err := LogParseLevel(tc.param)
		if tc.expectederr != nil && err == nil {
			t.Errorf("expected error but found nil (with param %s)", tc.param)
		}
		if tc.expectederr == nil && err != nil {
			t.Errorf("found unexpected error `%v` (with param %s)", err, tc.param)
		}
		if res != tc.expectedresult {
			t.Errorf("unexpected result, found `%v` (with param %s)", res, tc.param)
		}
	}
}

func TestIsEnable(t *testing.T) {
	type testCase struct {
		param          Level
		loglevel       Level
		expectedresult bool
	}
	tcases := []testCase{
		{
			loglevel:       WARN,
			param:          DEBUG,
			expectedresult: false,
		},
		{
			loglevel:       DEBUG,
			param:          WARN,
			expectedresult: true,
		},
		{
			loglevel:       WARN,
			param:          WARN,
			expectedresult: true,
		},
		{
			loglevel:       WARN,
			param:          LOG,
			expectedresult: false,
		},
	}
	for _, tc := range tcases {
		l := &Logger{lvl: tc.loglevel}
		res := l.isEnable(tc.param)
		if res != tc.expectedresult {
			t.Errorf(
				"unexpected result, found `%v` (with param %s, and initlog %s)",
				res,
				tc.param.ToString(),
				tc.loglevel.ToString(),
			)
		}
	}
}
