package main

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v32/github"
)

func TestParseDependabotPullRequest(t *testing.T) {
	testCases := []struct {
		issue    *github.Issue
		expected *dependabotPullRequest
	}{
		{
			issue: &github.Issue{
				Title: stringPtr("Bump jsdom from 12.2.0 to 16.3.0"),
				Labels: []*github.Label{
					{
						Name: stringPtr("javascript"),
					},
				},
			},
			expected: &dependabotPullRequest{
				Library:     "jsdom",
				Language:    "javascript",
				FromVersion: "12.2.0",
				ToVersion:   "16.3.0",
				Directory:   "",
				Security:    false,
			},
		},
		{
			issue: &github.Issue{
				Title: stringPtr("Bump foo-lib from 1.11.0 to 1.12.0 in /foo-service"),
				Labels: []*github.Label{
					{
						Name: stringPtr("ruby"),
					},
				},
			},
			expected: &dependabotPullRequest{
				Library:     "foo-lib",
				Language:    "ruby",
				FromVersion: "1.11.0",
				ToVersion:   "1.12.0",
				Directory:   "foo-service",
				Security:    false,
			},
		},
		{
			issue: &github.Issue{
				Title: stringPtr("[Security] Bump bar-lib from 6.0.0 to 6.0.1 in /bar-service"),
				Labels: []*github.Label{
					{
						Name: stringPtr("javascript"),
					},
				},
			},
			expected: &dependabotPullRequest{
				Library:     "bar-lib",
				Language:    "javascript",
				FromVersion: "6.0.0",
				ToVersion:   "6.0.1",
				Directory:   "bar-service",
				Security:    true,
			},
		},
		{
			issue: &github.Issue{
				Title: stringPtr("Update grpc requirement from >= 1.19, < 1.29 to >= 1.19, < 1.31 in /baz-lib"),
				Labels: []*github.Label{
					{
						Name: stringPtr("go"),
					},
				},
			},
			expected: &dependabotPullRequest{
				Library:     "grpc",
				Language:    "go",
				FromVersion: ">= 1.19, < 1.29",
				ToVersion:   ">= 1.19, < 1.31",
				Directory:   "baz-lib",
			},
		},
	}

	for _, testCase := range testCases {
		actual, err := parseDependabotPullRequest(testCase.issue)
		if err != nil {
			t.Fatalf("Failed to parse: %s", err)
		}

		if !reflect.DeepEqual(testCase.expected, actual) {
			t.Errorf("Assertion failed:\n  expected: %#v\n  actual:   %#v", testCase.expected, actual)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
