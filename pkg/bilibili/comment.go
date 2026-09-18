package bilibili

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

type CommentOptions struct {
	client   *BilibiliClient
	sortMode string
}

type CommentOption func(*CommentOptions)

func WithClient(client *BilibiliClient) CommentOption {
	return func(opts *CommentOptions) {
		if client != nil {
			opts.client = client.Clone()
		}
	}
}

func WithCookie(sessdata string) CommentOption {
	return func(opts *CommentOptions) {
		ensureCommentClient(opts)
		opts.client.SetCookies(map[string]string{"SESSDATA": sessdata})
	}
}

func WithAppAuth(appkey, appsec string) CommentOption {
	return func(opts *CommentOptions) {
		ensureCommentClient(opts)
		opts.client.SetAppAuth(appkey, appsec)
	}
}

func WithSortMode(sortMode string) CommentOption {
	return func(opts *CommentOptions) {
		opts.sortMode = sortMode
	}
}

type AuthOption = CommentOption

func ensureCommentClient(opts *CommentOptions) {
	if opts.client == nil {
		opts.client = DefaultClient()
	}
}

func resolveCommentOptions(commentOptions ...CommentOption) *CommentOptions {
	opts := &CommentOptions{
		client:   DefaultClient(),
		sortMode: "time",
	}
	for _, option := range commentOptions {
		option(opts)
	}
	ensureCommentClient(opts)
	return opts
}

func GetComments(oid int64, pn int, ps int, next int, commentOptions ...CommentOption) (*CommentResponse, error) {
	return GetCommentsContext(context.Background(), oid, pn, ps, next, commentOptions...)
}

func GetCommentsContext(ctx context.Context, oid int64, pn int, ps int, next int, commentOptions ...CommentOption) (*CommentResponse, error) {
	return GetCommentsWithOffsetContext(ctx, oid, pn, ps, next, "", commentOptions...)
}

func GetCommentsWithOffset(oid int64, pn int, ps int, next int, nextOffset string, commentOptions ...CommentOption) (*CommentResponse, error) {
	return GetCommentsWithOffsetContext(context.Background(), oid, pn, ps, next, nextOffset, commentOptions...)
}

func GetCommentsWithOffsetContext(ctx context.Context, oid int64, pn int, ps int, next int, nextOffset string, commentOptions ...CommentOption) (*CommentResponse, error) {
	opts := resolveCommentOptions(commentOptions...)
	resp, _, err := fetchMainComments(ctx, opts, oid, pn, ps, next, nextOffset)
	return resp, err
}

func fetchMainComments(ctx context.Context, opts *CommentOptions, oid int64, pn int, ps int, next int, nextOffset string) (*CommentResponse, bool, error) {
	params := url.Values{}
	params.Set("oid", fmt.Sprintf("%d", oid))
	params.Set("type", "1")
	if opts.sortMode == "hot" {
		params.Set("mode", "3")
	} else {
		params.Set("mode", "2")
	}

	switch {
	case pn <= 1:
		params.Set("pagination_str", `{"offset":"{\"type\":1,\"direction\":1,\"data\":{}}"}`)
	case nextOffset != "":
		if nextOffset[0] == '{' {
			params.Set("pagination_str", nextOffset)
		} else {
			offsetJSON, err := json.Marshal(nextOffset)
			if err != nil {
				return nil, false, fmt.Errorf("编码 next_offset 失败: %w", err)
			}
			params.Set("pagination_str", fmt.Sprintf(`{"offset":%s}`, string(offsetJSON)))
		}
	case next != 0:
		params.Set("pagination_str", fmt.Sprintf(`{"offset":"{\"type\":1,\"direction\":1,\"data\":{\"cursor\":%d}}"}`, next))
	default:
		return nil, false, fmt.Errorf("missing pagination cursor for page %d", pn)
	}

	signedParams, err := opts.client.signParams(ctx, params)
	if err != nil {
		return nil, false, fmt.Errorf("WBI 签名失败: %w", err)
	}

	body, err := opts.client.SendRequestContext(ctx, opts.client.replyMainURL+"?"+signedParams.Encode())
	if err != nil {
		return nil, false, err
	}

	var resp CommentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, false, fmt.Errorf("解析评论 JSON 失败: %w", err)
	}
	if resp.Code == -403 {
		fallback, err := fetchFallbackComments(ctx, opts, oid, pn, ps)
		return fallback, true, err
	}
	if resp.Code != 0 {
		return nil, false, fmt.Errorf("评论 API 返回错误，错误码: %d, 错误信息: %s", resp.Code, resp.Message)
	}
	return &resp, false, nil
}

func GetCommentsFallback(oid int64, pn int, ps int, commentOptions ...CommentOption) (*CommentResponse, error) {
	return GetCommentsFallbackContext(context.Background(), oid, pn, ps, commentOptions...)
}

func GetCommentsFallbackContext(ctx context.Context, oid int64, pn int, ps int, commentOptions ...CommentOption) (*CommentResponse, error) {
	return fetchFallbackComments(ctx, resolveCommentOptions(commentOptions...), oid, pn, ps)
}

func fetchFallbackComments(ctx context.Context, opts *CommentOptions, oid int64, pn int, ps int) (*CommentResponse, error) {
	params := url.Values{}
	params.Set("oid", fmt.Sprintf("%d", oid))
	params.Set("pn", fmt.Sprintf("%d", pn))
	params.Set("ps", fmt.Sprintf("%d", ps))
	params.Set("type", "1")
	if opts.sortMode == "hot" {
		params.Set("sort", "1")
	} else {
		params.Set("sort", "2")
	}

	signedParams, err := opts.client.signParams(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("WBI 签名失败: %w", err)
	}
	body, err := opts.client.SendRequestContext(ctx, opts.client.replyFallbackURL+"?"+signedParams.Encode())
	if err != nil {
		return nil, err
	}

	var resp CommentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析备用评论 JSON 失败: %w", err)
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("备用评论 API 返回错误，错误码: %d, 错误信息: %s", resp.Code, resp.Message)
	}
	return &resp, nil
}

func GetHotComments(oid int64, pn int, ps int, commentOptions ...CommentOption) (*CommentResponse, error) {
	opts := append(commentOptions, WithSortMode("hot"))
	return GetComments(oid, pn, ps, 0, opts...)
}

// CommentPaginator is the single source of truth for comment pagination.
// It carries cursor/next_offset between pages and transparently sticks to the
// fallback page-number API when the main endpoint returns -403.
type CommentPaginator struct {
	oid      int64
	pageSize int
	maxPages int
	page     int

	nextCursor int
	nextOffset string
	done       bool
	fallback   bool
	opts       *CommentOptions
}

func NewCommentPaginator(oid int64, pageSize, maxPages int, commentOptions ...CommentOption) *CommentPaginator {
	if pageSize <= 0 {
		pageSize = 20
	}
	if maxPages <= 0 {
		maxPages = 100
	}
	return &CommentPaginator{
		oid:      oid,
		pageSize: pageSize,
		maxPages: maxPages,
		opts:     resolveCommentOptions(commentOptions...),
	}
}

func (p *CommentPaginator) Next(ctx context.Context) (*CommentResponse, error) {
	if p == nil || p.done || p.page >= p.maxPages {
		return nil, io.EOF
	}
	nextPage := p.page + 1

	var (
		resp         *CommentResponse
		usedFallback bool
		err          error
	)
	if p.fallback {
		resp, err = fetchFallbackComments(ctx, p.opts, p.oid, nextPage, p.pageSize)
		usedFallback = true
	} else {
		resp, usedFallback, err = fetchMainComments(ctx, p.opts, p.oid, nextPage, p.pageSize, p.nextCursor, p.nextOffset)
	}
	if err != nil {
		return nil, err
	}

	p.page = nextPage
	if usedFallback {
		p.fallback = true
		total := resp.Data.Page.Count
		if len(resp.Data.Replies) == 0 || (total > 0 && p.page*p.pageSize >= total) || p.page >= p.maxPages {
			p.done = true
		}
		return resp, nil
	}

	p.nextCursor = resp.Data.Cursor.Next
	p.nextOffset = resp.Data.Cursor.PaginationReply.NextOffset
	if (p.nextCursor == 0 && p.nextOffset == "") || p.page >= p.maxPages {
		p.done = true
	}
	return resp, nil
}

func (p *CommentPaginator) Done() bool {
	return p == nil || p.done
}

func (p *CommentPaginator) Page() int {
	if p == nil {
		return 0
	}
	return p.page
}

func GetAllComments(oid int64, commentOptions ...CommentOption) ([]CommentData, error) {
	return GetAllCommentsContext(context.Background(), oid, commentOptions...)
}

func GetAllCommentsContext(ctx context.Context, oid int64, commentOptions ...CommentOption) ([]CommentData, error) {
	paginator := NewCommentPaginator(oid, 20, 100, commentOptions...)
	unique := make(map[int64]struct{})
	all := make([]CommentData, 0)

	for {
		resp, err := paginator.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("获取第 %d 页评论失败: %w", paginator.Page()+1, err)
		}

		for _, comment := range resp.Data.Replies {
			if _, exists := unique[comment.RPID]; exists {
				continue
			}
			unique[comment.RPID] = struct{}{}
			all = append(all, comment)
		}

		if paginator.Done() {
			break
		}
	}
	return all, nil
}

func GetSubComments(oid int64, root int64, commentOptions ...CommentOption) ([]CommentData, error) {
	return GetSubCommentsContext(context.Background(), oid, root, commentOptions...)
}

func GetSubCommentsContext(ctx context.Context, oid int64, root int64, commentOptions ...CommentOption) ([]CommentData, error) {
	opts := resolveCommentOptions(commentOptions...)
	params := url.Values{}
	params.Set("oid", fmt.Sprintf("%d", oid))
	params.Set("root", fmt.Sprintf("%d", root))
	params.Set("type", "1")
	params.Set("pn", "1")
	params.Set("ps", "3")

	signedParams, err := opts.client.signParams(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("WBI 签名失败: %w", err)
	}
	body, err := opts.client.SendRequestContext(ctx, opts.client.subReplyURL+"?"+signedParams.Encode())
	if err != nil {
		return nil, err
	}

	var resp CommentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析子评论 JSON 失败: %w", err)
	}
	if resp.Code != 0 {
		// 子评论失败不应中断主评论任务。
		return []CommentData{}, nil
	}
	return resp.Data.Replies, nil
}
