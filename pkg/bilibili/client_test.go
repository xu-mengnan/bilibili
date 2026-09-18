package bilibili

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func testClient(server *httptest.Server) *BilibiliClient {
	client := NewBilibiliClient()
	client.client = server.Client()
	client.limiter = newRequestLimiter(0)
	client.wbiCache = &wbiCache{}
	client.baseBackoff = time.Millisecond
	client.maxRetries = 2
	return client
}

func TestSendRequestRetriesRetryableStatus(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			http.Error(w, "temporary", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := testClient(server)
	body, err := client.SendRequestContext(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "ok" {
		t.Fatalf("unexpected body: %q", body)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestSendRequestContextCancellation(t *testing.T) {
	started := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		<-r.Context().Done()
	}))
	defer server.Close()

	client := testClient(server)
	client.maxRetries = 0

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := client.SendRequestContext(ctx, server.URL)
		errCh <- err
	}()

	select {
	case <-started:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}

	select {
	case err := <-errCh:
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("request did not cancel promptly")
	}
}

func TestWBIKeyCacheCollapsesRefreshes(t *testing.T) {
	var calls int32
	img := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	sub := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w,
			`{"code":0,"data":{"wbi_img":{"img_url":"https://i0.hdslb.com/bfs/wbi/%s.png","sub_url":"https://i0.hdslb.com/bfs/wbi/%s.png"}}}`,
			img, sub,
		)
	}))
	defer server.Close()

	client := testClient(server)
	client.navURL = server.URL
	client.wbiTTL = time.Hour

	first, err := client.getWBIKey(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := client.getWBIKey(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("cached key mismatch: %#v %#v", first, second)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected one /nav request, got %d", got)
	}
}

func TestMalformedWBIKeyReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"wbi_img":{"img_url":"https://x/a.png","sub_url":"https://x/b.png"}}}`))
	}))
	defer server.Close()

	client := testClient(server)
	client.navURL = server.URL

	if _, err := client.getWBIKey(context.Background()); err == nil {
		t.Fatal("expected malformed WBI key to fail")
	}
}

func TestGetMixinKeyRejectsShortMaterial(t *testing.T) {
	if _, err := getMixinKey("short"); err == nil {
		t.Fatal("expected short WBI material to return an error")
	}
}
