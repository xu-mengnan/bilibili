package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallLLMStreamMockEmitsDeltaChunks(t *testing.T) {
	service := NewAnalysisService("", "", "")
	var chunks []string

	result, err := service.CallLLMStream(context.Background(), func(chunk string) {
		chunks = append(chunks, chunk)
	}, "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple delta chunks, got %d", len(chunks))
	}
	if joined := strings.Join(chunks, ""); joined != result {
		t.Fatalf("joined delta chunks differ from final result")
	}

	totalBytes := 0
	for _, chunk := range chunks {
		totalBytes += len(chunk)
	}
	if totalBytes != len(result) {
		t.Fatalf("streamed bytes=%d result bytes=%d; expected linear delta streaming", totalBytes, len(result))
	}
}

func TestCallLLMStreamRealSSEEmitsDeltas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("missing authorization header")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		for _, chunk := range []string{"hello ", "world"} {
			_, _ = fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", chunk)
			if flusher != nil {
				flusher.Flush()
			}
		}
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
		if flusher != nil {
			flusher.Flush()
		}
	}))
	defer server.Close()

	service := NewAnalysisService(server.URL, "test-key", "test-model")
	service.httpClient = server.Client()

	var chunks []string
	result, err := service.CallLLMStream(context.Background(), func(chunk string) {
		chunks = append(chunks, chunk)
	}, "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if result != "hello world" {
		t.Fatalf("result=%q", result)
	}
	if len(chunks) != 2 || chunks[0] != "hello " || chunks[1] != "world" {
		t.Fatalf("unexpected chunks: %#v", chunks)
	}
}
