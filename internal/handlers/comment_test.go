package handlers

import "testing"

func TestValidateAndNormalizeScrapeRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     ScrapeRequest
		wantID  string
		wantErr bool
	}{
		{
			name:   "bvid url",
			req:    ScrapeRequest{VideoID: "https://www.bilibili.com/video/BV1xx411c7mD", PageLimit: 10, DelayMs: 300, AuthType: "none", SortMode: "time"},
			wantID: "BV1xx411c7mD",
		},
		{
			name:   "av id",
			req:    ScrapeRequest{VideoID: "av123456", PageLimit: 2, DelayMs: 300, AuthType: "none", SortMode: "hot"},
			wantID: "av123456",
		},
		{
			name:    "invalid video",
			req:     ScrapeRequest{VideoID: "not-a-video", PageLimit: 2, DelayMs: 300},
			wantErr: true,
		},
		{
			name:    "page too large",
			req:     ScrapeRequest{VideoID: "BV1xx411c7mD", PageLimit: 51, DelayMs: 300},
			wantErr: true,
		},
		{
			name:    "delay too small",
			req:     ScrapeRequest{VideoID: "BV1xx411c7mD", PageLimit: 2, DelayMs: 50},
			wantErr: true,
		},
		{
			name:    "invalid auth type",
			req:     ScrapeRequest{VideoID: "BV1xx411c7mD", PageLimit: 2, DelayMs: 300, AuthType: "magic"},
			wantErr: true,
		},
		{
			name:    "cookie requires value",
			req:     ScrapeRequest{VideoID: "BV1xx411c7mD", PageLimit: 2, DelayMs: 300, AuthType: "cookie"},
			wantErr: true,
		},
		{
			name:    "app requires both credentials",
			req:     ScrapeRequest{VideoID: "BV1xx411c7mD", PageLimit: 2, DelayMs: 300, AuthType: "app", AppKey: "key"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.req
			err := validateAndNormalizeScrapeRequest(&req)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected validation error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if req.VideoID != tt.wantID {
				t.Fatalf("VideoID=%q want=%q", req.VideoID, tt.wantID)
			}
		})
	}
}

func TestValidateScrapeRequestDefaults(t *testing.T) {
	req := ScrapeRequest{VideoID: "BV1xx411c7mD"}
	if err := validateAndNormalizeScrapeRequest(&req); err != nil {
		t.Fatal(err)
	}
	if req.PageLimit != 2 || req.DelayMs != 300 || req.AuthType != "none" || req.SortMode != "time" {
		t.Fatalf("unexpected defaults: %#v", req)
	}
}
