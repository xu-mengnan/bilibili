package handlers

import (
	"context"
	"testing"
	"time"

	"bilibili/internal/services"
)

type blockingTestStreamer struct {
	started chan struct{}
}

func (s *blockingTestStreamer) CallLLMStream(ctx context.Context, callback services.ChunkCallback, prompt string) (string, error) {
	callback("first")
	close(s.started)
	<-ctx.Done()
	return "", ctx.Err()
}

func TestRunAnalysisStreamClosesOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	streamer := &blockingTestStreamer{started: make(chan struct{})}
	events := runAnalysisStream(ctx, streamer, "prompt")

	select {
	case event := <-events:
		if event.Chunk != "first" || event.Err != nil || event.Done {
			t.Fatalf("unexpected first event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("did not receive first stream event")
	}

	select {
	case <-streamer.started:
	case <-time.After(time.Second):
		t.Fatal("streamer did not start")
	}

	cancel()

	select {
	case _, ok := <-events:
		if ok {
			for range events {
			}
		}
	case <-time.After(time.Second):
		t.Fatal("stream event channel did not close after cancellation")
	}
}

type burstTestStreamer struct{}

func (s *burstTestStreamer) CallLLMStream(ctx context.Context, callback services.ChunkCallback, prompt string) (string, error) {
	for i := 0; i < 1000; i++ {
		callback("x")
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
	}
	return "done", nil
}

func TestRunAnalysisStreamBackpressureUnblocksOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	events := runAnalysisStream(ctx, &burstTestStreamer{}, "prompt")

	// Do not consume immediately; let the bounded event buffer apply backpressure.
	time.Sleep(20 * time.Millisecond)
	cancel()

	done := make(chan struct{})
	go func() {
		for range events {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stream producer remained blocked after cancellation")
	}
}
