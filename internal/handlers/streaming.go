package handlers

import (
	"context"

	"bilibili/internal/services"
)

type llmStreamer interface {
	CallLLMStream(context.Context, services.ChunkCallback, string) (string, error)
}

type analysisStreamEvent struct {
	Chunk string
	Err   error
	Done  bool
}

// runAnalysisStream is shared by both legacy and v2 SSE handlers.
// It applies backpressure and makes every channel send cancellation-aware.
func runAnalysisStream(ctx context.Context, streamer llmStreamer, prompt string) <-chan analysisStreamEvent {
	events := make(chan analysisStreamEvent, 32)

	send := func(event analysisStreamEvent) bool {
		select {
		case events <- event:
			return true
		case <-ctx.Done():
			return false
		}
	}

	go func() {
		defer close(events)

		_, err := streamer.CallLLMStream(ctx, func(chunk string) {
			send(analysisStreamEvent{Chunk: chunk})
		}, prompt)
		if err != nil {
			if ctx.Err() == nil {
				send(analysisStreamEvent{Err: err})
			}
			return
		}
		send(analysisStreamEvent{Done: true})
	}()

	return events
}
