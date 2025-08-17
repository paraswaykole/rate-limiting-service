package sdk

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v3"
)

type ExtractFunc func(c fiber.Ctx) (args []string, err error)

type Config struct {
	// URL to Rate Limiter /check, e.g. http://localhost:3123/check
	CheckURL string

	// Extract args from the incoming request.
	// key: who to limit (userId/api-key/ip/tenant)
	ArgsExtractor ExtractFunc

	// How long to wait for /check
	Timeout time.Duration

	// FailOpen=true -> allow traffic on RL errors; false -> block with 503
	FailOpen bool

	// Add a static label for this service/endpoint
	Key string

	// Optional: custom http.Client (reused across requests)
	HTTPClient *http.Client
}

func DefaultArgsExtractor() ExtractFunc {
	return func(c fiber.Ctx) ([]string, error) {
		args0 := c.Path()
		args1 := c.Method()
		args2 := c.IP()
		return []string{args0, args1, args2}, nil
	}
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 500 * time.Millisecond, KeepAlive: 60 * time.Second}).DialContext,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		MaxConnsPerHost:       0,
		MaxIdleConnsPerHost:   10,
		ForceAttemptHTTP2:     true,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// Middleware returns a Fiber middleware that checks with RL /check before proceeding.
func Middleware(cfg Config) fiber.Handler {
	if cfg.ArgsExtractor == nil {
		cfg.ArgsExtractor = DefaultArgsExtractor()
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = NewHTTPClient(cfg.Timeout)
	}
	checkURL, err := url.Parse(cfg.CheckURL)
	if err != nil {
		panic(fmt.Errorf("rlsdk: invalid CheckURL: %w", err))
	}

	return func(c fiber.Ctx) error {
		key := cfg.Key
		args, err := cfg.ArgsExtractor(c)
		if err != nil || key == "" {
			// If extractor fails, choose fail-open/closed behavior
			if cfg.FailOpen {
				return c.Next()
			}
			return c.Status(fiber.StatusBadRequest).SendString("rate limit: missing key")
		}

		// Build /check?key=...&args[0]=..&args[1]=...
		q := checkURL.Query()
		q.Set("key", key)
		for _, arg := range args {
			q.Add("args", arg)
		}

		u := *checkURL
		u.RawQuery = q.Encode()

		// Context with the request deadline tied to Fiber's context
		ctx, cancel := context.WithTimeout(c.RequestCtx(), cfg.HTTPClient.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			if cfg.FailOpen {
				return c.Next()
			}
			return c.Status(fiber.StatusServiceUnavailable).SendString("rate limit: request build error")
		}

		resp, err := cfg.HTTPClient.Do(req)
		if err != nil {
			// Timeout / network error — choose fail-open vs fail-closed
			if cfg.FailOpen {
				return c.Next()
			}
			return c.Status(fiber.StatusServiceUnavailable).SendString("rate limit: unavailable")
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)

		// Forward X-RateLimit-* headers to client if provided
		if v := resp.Header.Get("X-RateLimit-Limit"); v != "" {
			c.Set("X-RateLimit-Limit", v)
		}
		if v := resp.Header.Get("X-RateLimit-Remaining"); v != "" {
			c.Set("X-RateLimit-Remaining", v)
		}
		if v := resp.Header.Get("X-RateLimit-Reset"); v != "" {
			c.Set("X-RateLimit-Reset", v)
		}

		// Allow / Deny
		switch resp.StatusCode {
		case http.StatusOK, http.StatusNoContent:
			// Allowed
			return c.Next()
		case http.StatusTooManyRequests:
			// Denied
			return c.Status(fiber.StatusTooManyRequests).SendString("Too Many Requests")
		default:
			// Unexpected RL status
			if cfg.FailOpen {
				return c.Next()
			}
			return c.Status(fiber.StatusServiceUnavailable).SendString("rate limit: error")
		}
	}
}
