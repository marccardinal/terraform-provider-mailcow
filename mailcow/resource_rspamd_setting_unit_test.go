package mailcow

import (
	"reflect"
	"strings"
	"testing"
)

func TestRspamdSettingContent(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "heredoc with trailing newline",
			content:  "priority = high;\napply {\n  want_spam = no;\n}\n",
			expected: "priority = high;\napply {\n  want_spam = no;\n}",
		},
		{
			name:     "CRLF line endings, as stored by the admin UI",
			content:  "priority = high;\r\napply {\r\n  want_spam = no;\r\n}",
			expected: "priority = high;\napply {\n  want_spam = no;\n}",
		},
		{
			name:     "whitespace only",
			content:  " \r\n\t\n",
			expected: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rspamdSettingContent(tc.content); got != tc.expected {
				t.Errorf("rspamdSettingContent(%q) = %q, want %q", tc.content, got, tc.expected)
			}
		})
	}
}

func TestRspamdSettingNewId(t *testing.T) {
	rule := func(id interface{}, desc, content string) map[string]interface{} {
		return map[string]interface{}{"id": id, "desc": desc, "content": content, "active": 1}
	}
	before := []map[string]interface{}{rule(float64(1), "existing", "a;")}

	testCases := []struct {
		name      string
		after     []map[string]interface{}
		expected  string
		expectErr string
	}{
		{
			name:     "one new rule",
			after:    []map[string]interface{}{rule(float64(1), "existing", "a;"), rule(float64(2), "new", "b;")},
			expected: "2",
		},
		{
			name:     "id as a string, content stored with CRLF",
			after:    []map[string]interface{}{rule("1", "existing", "a;"), rule("7", "new", "b;\r\n")},
			expected: "7",
		},
		{
			name:     "an older rule with the same description and content is ignored",
			after:    []map[string]interface{}{rule(float64(1), "new", "b;"), rule(float64(2), "new", "b;")},
			expected: "2",
		},
		{
			name:      "nothing new",
			after:     []map[string]interface{}{rule(float64(1), "existing", "a;")},
			expectErr: "not found after create",
		},
		{
			name:      "two new identical rules",
			after:     []map[string]interface{}{rule(float64(2), "new", "b;"), rule(float64(3), "new", "b;")},
			expectErr: "2 new rules match",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := rspamdSettingNewId(before, tc.after, "new", "b;")
			if tc.expectErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectErr) {
					t.Fatalf("expected error containing %q, got %v (id %q)", tc.expectErr, err, got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.expected {
				t.Errorf("id = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestDecodeList(t *testing.T) {
	testCases := []struct {
		name      string
		body      string
		expected  []map[string]interface{}
		expectErr bool
	}{
		{name: "empty list answered as {}", body: "{}", expected: []map[string]interface{}{}},
		{name: "empty list answered as {} with whitespace", body: " {}\n", expected: []map[string]interface{}{}},
		{name: "empty array", body: "[]", expected: []map[string]interface{}{}},
		{name: "list", body: `[{"id": 1}]`, expected: []map[string]interface{}{{"id": float64(1)}}},
		{name: "non-empty object", body: `{"type": "error"}`, expectErr: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := decodeList(strings.NewReader(tc.body))
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected an error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("decodeList(%q) = %v, want %v", tc.body, got, tc.expected)
			}
		})
	}
}
