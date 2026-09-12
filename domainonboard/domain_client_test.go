package domainonboard

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerifyDomainRetriesRateLimitWithSameIdempotencyKey(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "domain-verify:mail.school.test" {
			t.Fatalf("idempotency key = %q", got)
		}
		if requests == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"ok":false,"error":{"message":"retry later"}}`)
			return
		}
		fmt.Fprint(w, `{"ok":true,"data":{"domain":"mail.school.test","verification":{"status":"pending"}},"metadata":{}}`)
	}))
	defer server.Close()

	client := New("test-key")
	client.BaseURL = server.URL
	client.Sleep = func(_ context.Context, delay time.Duration) error {
		if delay != 2*time.Second {
			t.Fatalf("delay = %s", delay)
		}
		return nil
	}

	domain, err := client.VerifyDomain(context.Background(), "mail.school.test")
	if err != nil {
		t.Fatal(err)
	}
	if domain.Verification.Status != "pending" || requests != 2 {
		t.Fatalf("domain = %+v, requests = %d", domain, requests)
	}
}
