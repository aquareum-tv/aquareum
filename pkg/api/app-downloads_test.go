package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegexAquareum(t *testing.T) {
	tests := []struct {
		filename       string
		shouldMatch    bool
		expectedGroups []string
	}{
		// Test cases for the 're' regex
		{"aquareum-v1.2.3-abcdef-foo-bar.txt", true, []string{"v1.2.3", "-abcdef", "foo", "bar", "txt"}},
		{"aquareum-v1.0.0-123456-hello-world.csv", true, []string{"v1.0.0", "-123456", "hello", "world", "csv"}},
		{"aquareum-v2.5.1-abc123-done-done.xml", true, []string{"v2.5.1", "-abc123", "done", "done", "xml"}},
		{"aquareum-v3.2.1-xyz-abc.json", true, []string{"v3.2.1", "", "xyz", "abc", "json"}},
		{"aquareum-v3.2.1-nohash-xyz.json", true, []string{"v3.2.1", "", "nohash", "xyz", "json"}},
		{"aquareum-v10.2.10-abc123-linux-amd64.json", true, []string{"v10.2.10", "-abc123", "linux", "amd64", "json"}},
		{"aquareum-v10.2.10-darwin-arm64.json", true, []string{"v10.2.10", "", "darwin", "arm64", "json"}},

		// Test cases where the regex should not match
		{"aquareum-123-abc.txt", false, nil},
		{"aquareum-v1.2.3-abc.txt", false, nil},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			match := re.FindStringSubmatch(test.filename)
			if test.shouldMatch {
				require.NotNil(t, match, "Expected match for filename %s", test.filename)
				require.Len(t, match, 6, "Unexpected number of capture groups for filename %s", test.filename)
				for i, expected := range test.expectedGroups {
					require.Equal(t, expected, match[i+1], "Unexpected group %d for filename %s", i, test.filename)
				}
			} else {
				require.Nil(t, match, "Expected no match for filename %s", test.filename)
			}
		})
	}
}

func TestRegexInput(t *testing.T) {
	tests := []struct {
		filename       string
		shouldMatch    bool
		expectedGroups []string
	}{
		// Test cases for the 'inputRe' regex
		{"aquareum-foo-bar.txt", true, []string{"foo", "bar", "txt"}},
		{"aquareum-abc-def.csv", true, []string{"abc", "def", "csv"}},
		{"aquareum-x-y.xml", true, []string{"x", "y", "xml"}},
		{"aquareum-hello-world.json", true, []string{"hello", "world", "json"}},

		// Test cases where the regex should not match
		{"aquareum-foo.txt", false, nil},
		{"aquareum-foo-bar-baz.txt", false, nil},
		{"aquareum-foo-bar-baz-qux.txt", false, nil},
		{"aquareumfoo-bar.txt", false, nil},
		{"aquareum-foo-bar.", false, nil},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			match := inputRe.FindStringSubmatch(test.filename)
			if test.shouldMatch {
				require.NotNil(t, match, "Expected match for filename %s", test.filename)
				require.Len(t, match, 4, "Unexpected number of capture groups for filename %s", test.filename)
				for i, expected := range test.expectedGroups {
					require.Equal(t, expected, match[i+1], "Unexpected group %d for filename %s", i, test.filename)
				}
			} else {
				require.Nil(t, match, "Expected no match for filename %s", test.filename)
			}
		})
	}
}
