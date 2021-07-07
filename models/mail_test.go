package models

import (
	"reflect"
	"testing"

	"github.com/emersion/go-imap"
)

func TestBuildBodyPath(t *testing.T) {
	testCases := []struct {
		input    []int
		expected BodyPath
	}{
		{
			input:    []int{0, 1, 0},
			expected: "/0/1/0",
		},
	}
	for _, tc := range testCases {
		bp := BodyPathFromMessagePath(tc.input)
		if tc.expected != bp {
			t.Errorf("body path not build correctly (expected: %s, got: %s)", tc.expected, bp)
		}
	}
}

func TestBodyPathToMessagePath(t *testing.T) {
	testcases := []struct {
		input    BodyPath
		expected []int
		err      error
	}{
		{
			input:    "/0/1/0",
			expected: []int{0, 1, 0},
		},
		{
			input:    "/",
			expected: []int{},
		},
	}
	for _, tc := range testcases {
		got, err := tc.input.ToMessagePath()
		if tc.err != nil && err == nil {
			t.Error("expected error got nil")
		} else if err != nil && tc.err == nil {
			t.Errorf("unexpected error: %v", err)
		} else if err != tc.err {
			t.Errorf("unexpected error (expected: %v, got: %v)", tc.err, err)
		} else if !reflect.DeepEqual(tc.expected, got) {
			t.Errorf("body path not converted correctly (expected: %v, got: %v)", tc.expected, got)
		}
	}
}

func TestBodyStructure(t *testing.T) {
	testcases := []struct {
		input    imap.BodyStructure
		expected []*BodyPart
	}{
		{
			input: imap.BodyStructure{
				MIMEType:    "multipart",
				MIMESubType: "mixed",
				Parts: []*imap.BodyStructure{
					&imap.BodyStructure{
						MIMEType:    "text",
						MIMESubType: "plain",
					},
					&imap.BodyStructure{
						MIMEType:    "message",
						MIMESubType: "rfc822",
					},
					&imap.BodyStructure{
						MIMEType:    "message",
						MIMESubType: "rfc822",
					},
				},
			},
			expected: []*BodyPart{
				&BodyPart{
					Path:        "/",
					MIMEType:    "multipart",
					MIMESubType: "mixed",
				},
				&BodyPart{
					Path:        "/0",
					MIMEType:    "text",
					MIMESubType: "plain",
				},
				&BodyPart{
					Path:        "/1",
					MIMEType:    "message",
					MIMESubType: "rfc822",
				},
				&BodyPart{
					Path:        "/2",
					MIMEType:    "message",
					MIMESubType: "rfc822",
				},
			},
		},
		{
			input: imap.BodyStructure{
				MIMEType:    "multipart",
				MIMESubType: "mixed",
				Parts: []*imap.BodyStructure{
					&imap.BodyStructure{
						MIMEType:    "multipart",
						MIMESubType: "alternative",
						Parts: []*imap.BodyStructure{
							&imap.BodyStructure{
								MIMEType:    "text",
								MIMESubType: "plain",
							},
							&imap.BodyStructure{
								MIMEType:    "text",
								MIMESubType: "html",
							},
						},
					},
					&imap.BodyStructure{
						MIMEType:    "application",
						MIMESubType: "pgp-keys",
					},
				},
			},
			expected: []*BodyPart{
				&BodyPart{
					Path:        "/",
					MIMEType:    "multipart",
					MIMESubType: "mixed",
				},
				&BodyPart{
					Path:        "/0",
					MIMEType:    "multipart",
					MIMESubType: "alternative",
				},
				&BodyPart{
					Path:        "/0/0",
					MIMEType:    "text",
					MIMESubType: "plain",
				},
				&BodyPart{
					Path:        "/0/1",
					MIMEType:    "text",
					MIMESubType: "html",
				},
				&BodyPart{
					Path:        "/1",
					MIMEType:    "application",
					MIMESubType: "pgp-keys",
				},
			},
		},
	}
	for _, tc := range testcases {
		got := BodyPartsFromImap(&tc.input)
		if !reflect.DeepEqual(got, tc.expected) {
			t.Errorf("body structure not correctly built (expected: %#v, got: %#v)", tc.expected, got)
		}
	}
}
