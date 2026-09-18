package bilibili

import (\n\t"time"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestGetAllCommentsAdvancesCursorAndDeduplicates(t *testing.T) {
	img := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	sub := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	var (
		mu             sync.Mutex
		paginationSeen []string
		mainCalls      int
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/nav", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"wbi_img": map[string]string{
					"img_url": "https://i0.hdslb.com/bfs/wbi/" + img + ".png",
					"sub_url": "https://i0.hdslb.com/bfs/wbi/" + sub + ".png",
				},
			},
		})
	})
	mux.HandleFunc("/reply/main", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		mainCalls++
		call := mainCalls
		paginationSeen = append(paginationSeen, r.URL.Query().Get("pagination_str"))
		mu.Unlock()

		var resp CommentResponse
		resp.Code = 0
		switch call {
		case 1:
			resp.Data.Replies = []CommentData{{RPID: 1}, {RPID: 2}}
			resp.Data.Cursor.Next = 10
			resp.Data.Cursor.PaginationReply.NextOffset = `{"offset":"page2"}`
		case 2:
			resp.Data.Replies = []CommentData{{RPID: 2}, {RPID: 3}}
			resp.Data.Cursor.Next = 20
			resp.Data.Cursor.PaginationReply.NextOffset = `{"offset":"page3"}`
		case 3:
			resp.Data.Replies = []CommentData{{RPID: 4}}
		default:
			t.Fatalf("unexpected fourth page request")
		}
		_ = json.NewEncoder(w).Encode(&resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := testClient(server)
	client.navURL = server.URL + "/nav"
	client.replyMainURL = server.URL + "/reply/main"
	client.wbiTTL = timeHourForTest()

	comments, err := GetAllCommentsContext(context.Background(), 123, WithClient(client))
	if err != nil {
		t.Fatal(err)
	}

	if len(comments) != 4 {
		t.Fatalf("expected 4 unique comments, got %d: %#v", len(comments), comments)
	}
	for i, want := range []int64{1, 2, 3, 4} {
		if comments[i].RPID != want {
			t.Fatalf("comments[%d].RPID=%d want=%d", i, comments[i].RPID, want)
		}
	}

	mu.Lock()
	seen := append([]string(nil), paginationSeen...)
	mu.Unlock()
	if len(seen) != 3 {
		t.Fatalf("expected 3 page requests, got %d", len(seen))
	}
	if !strings.Contains(seen[0], `"data":{}`) {
		t.Fatalf("first page did not use initial pagination state: %q", seen[0])
	}
	if seen[1] != `{"offset":"page2"}` {
		t.Fatalf("second page did not carry next_offset: %q", seen[1])
	}
	if seen[2] != `{"offset":"page3"}` {
		t.Fatalf("third page did not carry next_offset: %q", seen[2])
	}
}

func timeHourForTest() time.Duration { return time.Hour }
