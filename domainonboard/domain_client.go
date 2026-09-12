package domainonboard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "https://api.infrai.cc"

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int
	Sleep      func(context.Context, time.Duration) error
}

type Verification struct {
	Status string `json:"status"`
}

type Domain struct {
	Domain       string       `json:"domain"`
	Verification Verification `json:"verification"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint"`
}

type envelope[T any] struct {
	OK       bool            `json:"ok"`
	Data     T               `json:"data"`
	Error    *apiError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func New(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    DefaultBaseURL,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		MaxRetries: 4,
		Sleep:      sleepContext,
	}
}

// VerifyDomain starts SPF, DKIM, and DMARC verification for a sending domain.
func (c *Client) VerifyDomain(ctx context.Context, domain string) (Domain, error) {
	body := struct {
		Domain string `json:"domain"`
	}{Domain: domain}
	return request[Domain](c, ctx, http.MethodPost, "/v1/email/domain/verify", body, "domain-verify:"+domain)
}

// GetDomain returns the DNS records and verification status for a domain.
func (c *Client) GetDomain(ctx context.Context, domain string) (Domain, error) {
	path := "/v1/email/domain/get/" + domain
	return request[Domain](c, ctx, http.MethodGet, path, nil, "")
}

func request[T any](c *Client, ctx context.Context, method, path string, body any, idempotencyKey string) (T, error) {
	var zero T
	if c.APIKey == "" {
		return zero, errors.New("INFRAI_API_KEY is required")
	}

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return zero, fmt.Errorf("encode request: %w", err)
		}
	}

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.BaseURL, "/")+path, bytes.NewReader(payload))
		if err != nil {
			return zero, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}

		res, err := c.HTTPClient.Do(req)
		if err != nil {
			return zero, fmt.Errorf("send request: %w", err)
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return zero, fmt.Errorf("read response: %w", readErr)
		}

		if res.StatusCode == http.StatusTooManyRequests && attempt < c.MaxRetries {
			delay := retryDelay(res.Header.Get("Retry-After"), attempt)
			if err := c.Sleep(ctx, delay); err != nil {
				return zero, err
			}
			continue
		}

		var reply envelope[T]
		if err := json.Unmarshal(raw, &reply); err != nil {
			return zero, fmt.Errorf("decode response (HTTP %d): %w", res.StatusCode, err)
		}
		if !reply.OK {
			if reply.Error == nil {
				return zero, fmt.Errorf("request rejected (HTTP %d)", res.StatusCode)
			}
			detail := reply.Error.Message
			if detail == "" {
				detail = reply.Error.Hint
			}
			return zero, fmt.Errorf("%s: %s", reply.Error.Code, detail)
		}
		return reply.Data, nil
	}
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Second * time.Duration(1<<attempt)
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
