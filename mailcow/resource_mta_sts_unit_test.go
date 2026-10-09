package mailcow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestMtaStsParsePolicy(t *testing.T) {
	testCases := []struct {
		name      string
		text      string
		expected  *mtaStsPolicy
		expectErr bool
	}{
		{
			name:     "policy as served by Mailcow",
			text:     "version: STSv1\nmode: enforce\nmax_age: 604800\nmx: mail.example.com\n",
			expected: &mtaStsPolicy{mode: "enforce", mx: []string{"mail.example.com"}, maxAge: 604800},
		},
		{
			name:     "several mx and CRLF line endings",
			text:     "version: STSv1\r\nmode: testing\r\nmax_age: 86400\r\nmx: mx1.example.com\r\nmx: *.example.net\r\n",
			expected: &mtaStsPolicy{mode: "testing", mx: []string{"mx1.example.com", "*.example.net"}, maxAge: 86400},
		},
		{
			name:      "missing mx",
			text:      "version: STSv1\nmode: enforce\nmax_age: 604800\n",
			expectErr: true,
		},
		{
			name:      "invalid max_age",
			text:      "version: STSv1\nmode: enforce\nmax_age: soon\nmx: mail.example.com\n",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			policy, err := mtaStsParsePolicy(tc.text)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected an error, got %+v", policy)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(policy, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, policy)
			}
		})
	}
}

func TestMtaStsFetchPolicy(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/mta-sts.txt" {
			http.NotFound(w, r)
			return
		}
		switch r.Host {
		case "mta-sts.active.example":
			_, _ = w.Write([]byte("version: STSv1\nmode: enforce\nmax_age: 604800\nmx: mail.example.com\n"))
		case "mta-sts.broken.example":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	hostName := server.Listener.Addr().String()

	policy, err := mtaStsFetchPolicy(ctx, server.Client(), hostName, "active.example")
	if err != nil {
		t.Fatalf("active policy: unexpected error: %v", err)
	}
	expected := &mtaStsPolicy{mode: "enforce", mx: []string{"mail.example.com"}, maxAge: 604800}
	if !reflect.DeepEqual(policy, expected) {
		t.Errorf("active policy: expected %+v, got %+v", expected, policy)
	}

	policy, err = mtaStsFetchPolicy(ctx, server.Client(), hostName, "missing.example")
	if err != nil || policy != nil {
		t.Errorf("missing policy: expected nil, nil; got %+v, %v", policy, err)
	}

	if _, err = mtaStsFetchPolicy(ctx, server.Client(), hostName, "broken.example"); err == nil {
		t.Error("server error: expected an error")
	}
}

func TestMtaStsMx(t *testing.T) {
	got := mtaStsMx([]interface{}{"mail.example.com", " mx2.example.com "})
	if got != "mail.example.com,mx2.example.com" {
		t.Errorf("unexpected mx string %q", got)
	}
}

func TestMtaStsPolicyId(t *testing.T) {
	ts := time.Date(2026, 10, 9, 14, 5, 7, 0, time.FixedZone("EDT", -4*3600))
	if got := mtaStsPolicyId(ts); got != "20261009180507" {
		t.Errorf("expected UTC id 20261009180507, got %s", got)
	}
}
