package bilibili

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultUserAgent        = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0 Safari/537.36"
	defaultResponseBodyLimit = int64(16 << 20) // 16 MiB
	errorBodyLimit           = int64(8 << 10)  // 8 KiB
)

type requestLimiter struct {
	mu       sync.Mutex
	next     time.Time
	interval time.Duration
}

func newRequestLimiter(interval time.Duration) *requestLimiter {
	return &requestLimiter{interval: interval}
}

func (l *requestLimiter) wait(ctx context.Context) error {
	if l == nil || l.interval <= 0 {
		return nil
	}

	l.mu.Lock()
	now := time.Now()
	scheduled := now
	if l.next.After(now) {
		scheduled = l.next
	}
	l.next = scheduled.Add(l.interval)
	wait := time.Until(scheduled)
	l.mu.Unlock()

	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type wbiCache struct {
	mu      sync.Mutex
	key     WBIKey
	expires time.Time
}

type BilibiliClient struct {
	client   *http.Client
	limiter  *requestLimiter
	wbiCache *wbiCache

	cookies map[string]string
	appkey  string
	appsec  string

	maxRetries        int
	baseBackoff       time.Duration
	responseBodyLimit int64
	wbiTTL            time.Duration

	navURL           string
	replyMainURL     string
	replyFallbackURL string
	subReplyURL      string
	videoViewURL     string
	userInfoURL      string
}

var sharedTransport = func() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Keep the original project's direct-connect behavior.
	transport.Proxy = nil
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 20
	transport.IdleConnTimeout = 90 * time.Second
	return transport
}()

var sharedHTTPClient = &http.Client{
	Timeout:   15 * time.Second,
	Transport: sharedTransport,
}

var (
	sharedBilibiliLimiter = newRequestLimiter(120 * time.Millisecond)
	sharedWBICache         = &wbiCache{}
	defaultBilibiliClient  = NewBilibiliClient()
)

func NewBilibiliClient() *BilibiliClient {
	return &BilibiliClient{
		client:            sharedHTTPClient,
		limiter:           sharedBilibiliLimiter,
		wbiCache:          sharedWBICache,
		maxRetries:        2,
		baseBackoff:       200 * time.Millisecond,
		responseBodyLimit: defaultResponseBodyLimit,
		wbiTTL:            30 * time.Minute,
		navURL:            "https://api.bilibili.com/x/web-interface/nav",
		replyMainURL:      "https://api.bilibili.com/x/v2/reply/main",
		replyFallbackURL:  "https://api.bilibili.com/x/v2/reply",
		subReplyURL:       "https://api.bilibili.com/x/v2/reply/reply",
		videoViewURL:      "https://api.bilibili.com/x/web-interface/view",
		userInfoURL:       "https://api.bilibili.com/x/space/acc/info",
	}
}

func DefaultClient() *BilibiliClient {
	return defaultBilibiliClient.Clone()
}

func (c *BilibiliClient) Clone() *BilibiliClient {
	if c == nil {
		return NewBilibiliClient()
	}
	clone := *c
	if c.cookies != nil {
		clone.cookies = make(map[string]string, len(c.cookies))
		for key, value := range c.cookies {
			clone.cookies[key] = value
		}
	}
	return &clone
}

func (c *BilibiliClient) SetCookies(cookies map[string]string) {
	c.cookies = make(map[string]string, len(cookies))
	for key, value := range cookies {
		c.cookies[key] = value
	}
}

func (c *BilibiliClient) SetAppAuth(appkey, appsec string) {
	c.appkey = appkey
	c.appsec = appsec
}

func (c *BilibiliClient) SendRequest(rawURL string) ([]byte, error) {
	return c.SendRequestContext(context.Background(), rawURL)
}

func (c *BilibiliClient) SendRequestContext(ctx context.Context, rawURL string) ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("bilibili client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.limiter.wait(ctx); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		c.applyHeaders(req)

		resp, err := c.client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			lastErr = fmt.Errorf("发送请求失败: %w", err)
			if attempt < c.maxRetries {
				if err := c.waitRetry(ctx, attempt); err != nil {
					return nil, err
				}
				continue
			}
			break
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			body, readErr := readLimitedBody(resp.Body, c.responseBodyLimit)
			closeErr := resp.Body.Close()
			if readErr != nil {
				return nil, fmt.Errorf("读取响应失败: %w", readErr)
			}
			if closeErr != nil {
				return nil, fmt.Errorf("关闭响应失败: %w", closeErr)
			}
			return body, nil
		}

		errorBody, _ := readLimitedBody(resp.Body, errorBodyLimit)
		_ = resp.Body.Close()
		lastErr = fmt.Errorf("Bilibili HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(errorBody)))

		if !isRetryableStatus(resp.StatusCode) || attempt >= c.maxRetries {
			break
		}
		if err := c.waitRetry(ctx, attempt); err != nil {
			return nil, err
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("Bilibili request failed")
	}
	return nil, lastErr
}

func (c *BilibiliClient) applyHeaders(req *http.Request) {
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://www.bilibili.com/")

	if len(c.cookies) > 0 {
		keys := make([]string, 0, len(c.cookies))
		for key := range c.cookies {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, key+"="+c.cookies[key])
		}
		req.Header.Set("Cookie", strings.Join(parts, "; "))
	}
	if c.appkey != "" {
		req.Header.Set("APP-KEY", c.appkey)
	}
}

func (c *BilibiliClient) waitRetry(ctx context.Context, attempt int) error {
	base := c.baseBackoff
	if base <= 0 {
		return nil
	}
	delay := base << attempt
	// Lightweight jitter without shared random state.
	jitterRange := delay / 2
	if jitterRange > 0 {
		delay += time.Duration(time.Now().UnixNano() % int64(jitterRange))
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusRequestTimeout || status >= 500
}

func readLimitedBody(reader io.Reader, limit int64) ([]byte, error) {
	if limit <= 0 {
		limit = defaultResponseBodyLimit
	}
	limited := io.LimitReader(reader, limit+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return body, nil
}
