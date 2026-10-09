package api

import (
	"encoding/json"
	"testing"
)

func TestMailcowResponseArrayUnmarshal(t *testing.T) {
	testCases := []struct {
		name         string
		body         string
		expectedLen  int
		expectedType string
	}{
		{
			name:         "array, as most endpoints answer",
			body:         `[{"type":"success","log":["mailbox","edit","mta_sts"],"msg":["object_modified","example.com"]}]`,
			expectedLen:  1,
			expectedType: "success",
		},
		{
			name:         "several entries, the last one counts",
			body:         `[{"type":"success","msg":"dkim_added"},{"type":"danger","msg":"access_denied"}]`,
			expectedLen:  2,
			expectedType: "danger",
		},
		{
			name:         "bare object, as add/mta-sts answers on success",
			body:         ` {"type":"success","msg":"Task completed"}`,
			expectedLen:  1,
			expectedType: "success",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var response MailcowResponseArray
			if err := json.Unmarshal([]byte(tc.body), &response); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(response) != tc.expectedLen {
				t.Fatalf("expected %d entries, got %d", tc.expectedLen, len(response))
			}
			if got := *response.GetFinalType(); got != tc.expectedType {
				t.Errorf("expected final type %q, got %q", tc.expectedType, got)
			}
		})
	}
}
