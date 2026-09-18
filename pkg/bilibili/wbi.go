package bilibili

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type WBIKey struct {
	ImgKey string
	SubKey string
}

type NavResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data struct {
		WbiImg struct {
			ImgURL string `json:"img_url"`
			SubURL string `json:"sub_url"`
		} `json:"wbi_img"`
	} `json:"data"`
}

func GetWBIKey() (WBIKey, error) {
	return GetWBIKeyContext(context.Background())
}

func GetWBIKeyContext(ctx context.Context) (WBIKey, error) {
	return defaultBilibiliClient.getWBIKey(ctx)
}

func (c *BilibiliClient) getWBIKey(ctx context.Context) (WBIKey, error) {
	if c == nil {
		return WBIKey{}, fmt.Errorf("bilibili client is nil")
	}
	if c.wbiCache == nil {
		c.wbiCache = &wbiCache{}
	}
	if c.wbiTTL <= 0 {
		c.wbiTTL = 30 * time.Minute
	}

	// This lock intentionally spans refresh so concurrent callers collapse into
	// one /nav request instead of stampeding the upstream endpoint.
	c.wbiCache.mu.Lock()
	defer c.wbiCache.mu.Unlock()

	if time.Now().Before(c.wbiCache.expires) {
		if err := validateWBIKey(c.wbiCache.key); err == nil {
			return c.wbiCache.key, nil
		}
	}

	body, err := c.SendRequestContext(ctx, c.navURL)
	if err != nil {
		return WBIKey{}, fmt.Errorf("获取 WBI 密钥失败: %w", err)
	}

	var navResp NavResponse
	if err := json.Unmarshal(body, &navResp); err != nil {
		return WBIKey{}, fmt.Errorf("解析 WBI 响应失败: %w", err)
	}
	if navResp.Code != 0 {
		return WBIKey{}, fmt.Errorf("WBI API 返回错误，错误码: %d, 错误信息: %s", navResp.Code, navResp.Message)
	}

	key := WBIKey{
		ImgKey: extractKeyFromURL(navResp.Data.WbiImg.ImgURL),
		SubKey: extractKeyFromURL(navResp.Data.WbiImg.SubURL),
	}
	if err := validateWBIKey(key); err != nil {
		return WBIKey{}, err
	}

	c.wbiCache.key = key
	c.wbiCache.expires = time.Now().Add(c.wbiTTL)
	return key, nil
}

func validateWBIKey(key WBIKey) error {
	if len(key.ImgKey) < 32 || len(key.SubKey) < 32 {
		return fmt.Errorf("invalid WBI key lengths: img=%d sub=%d", len(key.ImgKey), len(key.SubKey))
	}
	return nil
}

func extractKeyFromURL(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	filename := parsed.Path
	if idx := strings.LastIndex(filename, "/"); idx >= 0 {
		filename = filename[idx+1:]
	}
	if idx := strings.LastIndex(filename, "."); idx > 0 {
		filename = filename[:idx]
	}
	return filename
}

var mixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

func getMixinKey(orig string) (string, error) {
	var builder strings.Builder
	for _, index := range mixinKeyEncTab {
		if index >= len(orig) {
			return "", fmt.Errorf("WBI key material too short: need index %d, length %d", index, len(orig))
		}
		builder.WriteByte(orig[index])
	}
	mixed := builder.String()
	if len(mixed) < 32 {
		return "", fmt.Errorf("WBI mixin key too short: %d", len(mixed))
	}
	return mixed[:32], nil
}

func SignParams(params url.Values, wbiKey WBIKey) (url.Values, error) {
	if err := validateWBIKey(wbiKey); err != nil {
		return nil, err
	}

	signedParams := url.Values{}
	for key, values := range params {
		signedParams[key] = append([]string(nil), values...)
	}
	signedParams.Set("wts", strconv.FormatInt(time.Now().Unix(), 10))

	keys := make([]string, 0, len(signedParams))
	for key := range signedParams {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var query strings.Builder
	for _, key := range keys {
		if query.Len() > 0 {
			query.WriteByte('&')
		}
		value := signedParams.Get(key)
		value = strings.NewReplacer(
			"!", "",
			"'", "",
			"(", "",
			")", "",
			"*", "",
		).Replace(value)

		query.WriteString(key)
		query.WriteByte('=')
		query.WriteString(value)
	}

	mixinKey, err := getMixinKey(wbiKey.ImgKey + wbiKey.SubKey)
	if err != nil {
		return nil, err
	}
	hash := md5.Sum([]byte(query.String() + mixinKey))
	signedParams.Set("w_rid", hex.EncodeToString(hash[:]))
	return signedParams, nil
}

func (c *BilibiliClient) signParams(ctx context.Context, params url.Values) (url.Values, error) {
	key, err := c.getWBIKey(ctx)
	if err != nil {
		return nil, err
	}
	return SignParams(params, key)
}
